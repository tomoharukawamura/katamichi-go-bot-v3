output "bucket_name" {
  value = aws_s3_bucket.main.bucket
}

output "dynamodb_table" {
  value = aws_dynamodb_table.tflock.name
}

output "github_actions_role_arn" {
  description = "GitHub Actions に設定する IAM ロール ARN (secrets.AWS_ROLE_ARN)"
  value       = aws_iam_role.github_actions.arn
}
