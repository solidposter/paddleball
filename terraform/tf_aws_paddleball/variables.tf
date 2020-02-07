variable "key_name" {
  type        = string
  description = "Access key"
}

variable "region" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type to use"
  default     = "t2.nano"
}

variable "instance_tags" {
  type = map
  default = {
    Name = "paddleball"
  }
}

variable "user_data" {
  type        = string
  description = "User data that will be executerd when the instance is launched"
  default     = <<EOF
#!/bin/bash
echo "starting paddleball!"
docker run --rm -d --expose 2222 -p 2222:2222/udp gcr.io/public-registry-265316/kindred/paddleball -k 1984 -s 2222
echo "starting HTTP health endpoint"
touch index.html
python -m SimpleHTTPServer 2220
EOF
}

variable "ingress_cidr_block" {
  type        = string
  description = "Ingress block to allow traffic from the load balancer"
}

variable "health_ingress_cidr_block" {
  type        = string
  description = "Ingress block to allow load balancer health checks to reach through"
}

variable "route53_zone" {
  type        = string
  description = "Hosted zone, such as example.com."
}


