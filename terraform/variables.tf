variable "aws_region" {
  default = "ap-northeast-1"
}

variable "ssm_slack_token_sta" {
  description = "SSM Parameter Store ARN for sta SLACK_BOT_TOKEN"
  type        = string
}

variable "ssm_slack_token_pro" {
  description = "SSM Parameter Store ARN for pro SLACK_BOT_TOKEN"
  type        = string
}
