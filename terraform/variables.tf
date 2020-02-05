
variable "assume_role_arn" {
  default = "role_arn = "arn:aws:iam::996268458099:role/TerraformCrossAccount""
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

variable "instance_type" {
  description = "EC2 instance type to use"
  default     = "t2.nano"
}

variable "key_name" {
  default = "CloudFormation-eu-central-1"
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


variable "user_data" {
  description = "User data that will be executerd when the instance is launched"
  default     = <<EOF
#!/bin/bash
echo "starting paddleball!"
docker run -d --expose 2222 -p 2222:2222/udp gcr.io/public-registry-265316/kindred/paddleball -k 1984 -s 2222
echo "starting HTTP health endpoint"
touch index.html
python -m SimpleHTTPServer 2220
EOF
}

variable "ingress_cidr_block" {
    description = "Ingress block to allow traffic from the load balancer"
    default = "10.0.0.0/8"
}

variable "health_ingress_cidr_block" {
    description = "Ingress block to allow load balancer health checks to reach through"
    default = "10.0.0.0/8"
}
