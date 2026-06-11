variable "aws_region" {
  default = "ap-northeast-1"
}

variable "ssm_slack_token" {
  description = "SSM Parameter Store ARN for SLACK_BOT_TOKEN (shared by sta and pro)"
  type        = string
  default     = "arn:aws:ssm:ap-northeast-1:581476295353:parameter/katamichi/slack-token"
}
