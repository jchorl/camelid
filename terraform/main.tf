# config values
locals {
  conf = {
    CAMELID_RATIOS         = jsonencode({ VOO = 665, VXUS = 285, BND = 50 })
    CAMELID_MAX_INVESTMENT = 5000

    APCA_API_BASE_URL   = "https://paper-api.alpaca.markets"
    APCA_API_KEY_ID     = var.alpaca_api_key
    APCA_API_SECRET_KEY = var.alpaca_secret_key
  }
}

# dynamo config for trade record keeping and recon
resource "aws_dynamodb_table" "trade_records" {
  name           = "CamelidTradeRecords"
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

resource "aws_dynamodb_table" "trade_records_test" {
  name           = "CamelidRecordsTest"
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

# iam permissions for the cron job
resource "aws_iam_role" "role" {
  name = "camelid-lambda-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_iam_policy_document" "policy_document" {
  # lambda needs to talk to dynamo
  statement {
    sid = "DBAccess"

    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:GetItem",
      "dynamodb:BatchGetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem"
    ]

    resources = [
      aws_dynamodb_table.trade_records_test.arn
    ]
  }

  # logging may/may not be required
  statement {
    sid    = "Logging"
    effect = "Allow"

    resources = [
      "arn:aws:logs:*:*:*"
    ]

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
  }
}

resource "aws_iam_policy" "policy" {
  name        = "camelid-policy"
  description = "Policy for Camelid lambda"

  policy = data.aws_iam_policy_document.policy_document.json
}

resource "aws_iam_role_policy_attachment" "camelid-policy-attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy.arn
}

# define the lambda
resource "aws_lambda_function" "camelid_lambda" {
  filename      = "build/camelid_payload.zip"
  function_name = "camelid"
  role          = aws_iam_role.role.arn
  handler       = "main"

  source_code_hash = filebase64sha256("build/camelid_payload.zip")

  runtime = "go1.x"

  environment {
    variables = local.conf
  }
}

# cron up the lambda
resource "aws_cloudwatch_event_rule" "cron" {
  name                = "camelid-cron"
  description         = "Sends event to camelid cron"
  schedule_expression = "cron(38 16 ? * MON-FRI *)" # 16:38 UTC should be 12:38p ET
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule = aws_cloudwatch_event_rule.cron.name
  arn  = aws_lambda_function.camelid_lambda.arn
}

resource "aws_lambda_permission" "cloudwatch" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.camelid_lambda.arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.cron.arn
}
