resource "aws_dynamodb_table" "trade-records" {
  name           = "TradeRecords"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "ID"

  attribute {
    name = "ID"
    type = "S"
  }
}
