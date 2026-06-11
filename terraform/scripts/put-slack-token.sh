#!/bin/bash
set -euo pipefail

PARAM_NAME="/katamichi/slack-token"
REGION="ap-northeast-1"
PROFILE="${AWS_PROFILE:-default}"

if [ -z "${SLACK_BOT_TOKEN:-}" ]; then
  read -rsp "SLACK_BOT_TOKEN: " SLACK_BOT_TOKEN
  echo
fi

aws ssm put-parameter \
  --profile "$PROFILE" \
  --region "$REGION" \
  --name "$PARAM_NAME" \
  --value "$SLACK_BOT_TOKEN" \
  --type SecureString \
  --overwrite

echo "Registered: $PARAM_NAME"
