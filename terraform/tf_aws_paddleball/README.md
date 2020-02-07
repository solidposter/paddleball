# A Terraform module to create a Paddleball installation in AWS

A terraform module providing a Paddleball installation in AWS.

## Usage

```hcl
provider "aws" {
  region = "eu-central-1"
}

data "aws_region" "current" {}

module "paddleball" {
  source  = "github.com/kindredgroup/paddleball.git//terraform/tf_aws_paddleball"

  region                    = data.aws_region.current.name
  vpc_id                    = "vpc-1111111"
  ingress_cidr_block        = "10.0.0.0/8"
  health_ingress_cidr_block = "10.0.0.0/8"
  instance_tags = {
    Component   = "paddleball"
    CostCenter  = "PE"
    Department  = "AMO"
    Team        = "Platform Engineering"
    Environment = "Prod"
    Name        = "paddleball"
  }
  route53_zone = "aws.company.com."
}
```


## Authors

Created by [Anders Norrbom](https://github.com/norrbom)
