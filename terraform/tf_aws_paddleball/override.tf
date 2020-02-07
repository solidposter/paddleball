variable "key_name" {
  description = "Access keys"
  default     = "CloudFormation-eu-central-1"
}

variable "region" {
  default = "eu-central-1"
}

variable "availability_zones" {
  default = ["eu-central-1a"]
}
variable "vpc_id" {
  default = "vpc-d998f8b1"
}
variable "key_name" {
  default = "CloudFormation-eu-central-1"
}

variable "ingress_cidr_block" {
  default = "10.0.0.0/8"
}

variable "health_ingress_cidr_block" {
  default = "10.0.0.0/8"
}

variable "instance_tags" {
  default = {
    Component   = "paddleball"
    CostCenter  = "PE"
    Department  = "AMO"
    Team        = "Platform Engineering"
    Environment = "Prod"
    Name        = "paddleball"
  }
}

variable "route53_zone" {
  default = "amo-pe.aws.kindredgroup.com."
}