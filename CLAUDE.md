# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

トヨタレンタカー「片道GO！」ページの車種一覧を監視し、変更をSlackに通知するGoボット。

- 対象URL: https://cp.toyota.jp/rentacar/?padid=ag270_fr_sptop_onewayma
- 監視対象: `#service-items-shop-type-start .service-item` 要素（出発方向の車種一覧）
- チェック間隔: 10秒
- 変更検知: 追加・削除されたアイテムをSlack Incoming Webhookで通知
- 状態保存: `/data/state.json`（Dockerボリュームにマウント）

## Commands

```bash
# 起動（.envにSLACK_WEBHOOK_URLを設定してから）
docker compose up -d

# ログ確認
docker compose logs -f

# 停止
docker compose down

# ビルドのみ（ローカルGoが必要な場合）
go build ./...
go test ./...
```

## Architecture

```
main.go           – メインループ（10秒ごとにcheck()、1時間ごとにS3バックアップ）
scraper/          – HTTP取得 + goquery でHTML解析、CarItemのリストを返す
notifier/         – Slack Incoming Webhookへのメッセージ送信
storage/          – /data/state.json への状態読み書き（前回の車種キー一覧）
storage/s3.go     – S3へのアップロード・ダウンロード（バックアップ・リストア）
```

### 変更検知の仕組み

1. `scraper.Fetch()` で現在の車種一覧を取得
2. `storage.Load()` で前回保存した `map[key]bool` を読み込む
3. キーは `出発店舗|車種|期間` の複合キー
4. 新規キー = 追加、消えたキー = 削除
5. 差分があればSlack通知後に `storage.Save()` で状態を更新

### S3バックアップ

- 起動時: S3から `state.json` をリストア（存在しない場合は初回扱い）
- 毎時: `state.json` を S3 の `state.json` キーへ上書きアップロード
- EC2ではIAMロールで認証（アクセスキー不要）
- `S3_BUCKET` / `AWS_REGION` が未設定の場合はバックアップ無効で起動継続

## Setup

```bash
cp .env.example .env
# .envにSlack Webhook URLとS3設定を記入
docker compose up -d
```
