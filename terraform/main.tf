resource "aws_dynamodb_table" "trade-records" {
  name           = "TradeRecords"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "ID"

  global_secondary_index {
    name           = "StatusIndex"
    read_capacity  = 5
    write_capacity = 5

    hash_key           = "ID"
    range_key          = "Status"
    projection_type    = "INCLUDE"
    non_key_attributes = ["AlpacaOrderID"]
  }

  attribute {
    name = "ID"
    type = "S"
  }

  attribute {
    name = "Status"
    type = "N"
  }
}

resource "aws_dynamodb_table" "trade-records-test" {
  name           = "TradeRecordsTest"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "ID"

  global_secondary_index {
    name           = "StatusIndex"
    read_capacity  = 5
    write_capacity = 5

    hash_key           = "ID"
    range_key          = "Status"
    projection_type    = "INCLUDE"
    non_key_attributes = ["AlpacaOrderID"]
  }

  attribute {
    name = "ID"
    type = "S"
  }

  attribute {
    name = "Status"
    type = "N"
  }
}
