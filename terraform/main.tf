terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }

  backend "s3" {
    bucket         = "katamichi-go-bot-v3"
    key            = "terraform/katamichi-bot/terraform.tfstate"
    region         = "ap-northeast-1"
    dynamodb_table = "katamichi-bot-tflock"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}
