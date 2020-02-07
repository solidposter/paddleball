# A Terraform module to create a Paddleball installation in AWS

A terraform module providing a Paddleball installation in AWS.

## Usage

```hcl
module "paddleball" {
  source = "github.com/kindredgroup/paddleball.git//terraform/tf_aws_paddleball?ref=terraform"

  key_name                  = "CloudFormation-eu-central-1"
  region                    = "eu-central-1"
  availability_zones        = ["eu-central-1a"]
  vpc_id                    = "vpc-d998f8b1"
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
  route53_zone = "amo-pe.aws.kindredgroup.com."
}
```


## Authors

Created by [Anders Norrbom](https://github.com/norrbom)
