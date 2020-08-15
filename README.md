# Camelid
> Alpacas are part of the camelid family; therefore camelid is just a personal-finance abstraction over an [alpaca](https://alpaca.markets)

<p align="center">
  <img src="https://upload.wikimedia.org/wikipedia/commons/8/82/2011_Trampeltier_1528.JPG" alt="A real chonker" width="250"/>
</p>

## What?
Noting my enthusiasm with [terraform](https://terraform.io) and [Bogleism](https://www.bogleheads.org/) someone once joked about me terraforming my personal finance. Hilarious. This is that.

It's just a cron that runs every day and invests any available funds, maintaining portfolio ratios as best it can.

## Architecture
- It's a [lambda](https://aws.amazon.com/lambda/) function
- The lambda is triggered by a cloudwatch scheduled event, once a day on weekdays
- It uses dynamodb for state
- It reconciles trades to make sure they go through, and won't conduct further action until all are settled
- All configured by terraform

## Configuration
Configuration is mainly done using terraform locals in [terraform/main.tf](terraform/main.tf).

Notable env vars:
```
CAMELID_RATIOS         = jsonencode({ VOO = 665, VXUS = 285, BND = 50 })  # ratios of the tickers you'd like to hold, does not need to add up to 100 (its based on dollar value ratios)
CAMELID_MAX_INVESTMENT = 5000  # max amount to invest in one run
CAMELID_DRY_RUN        = "1"  # whether to actually trade or dry-run

APCA_API_BASE_URL   = "https://paper-api.alpaca.markets"  # the alpaca endpoint to hit, useful for testing
```

API keys are read from `.env` - copy [.env.template](.env.template) to `.env` and fill in the values.

## Deployment
```shell
$ make build
$ ./scripts/terraform.sh apply
```
