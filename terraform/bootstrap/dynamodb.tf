resource "aws_dynamodb_table" "tflock" {
  name         = "katamichi-bot-tflock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
