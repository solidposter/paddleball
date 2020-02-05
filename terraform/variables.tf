variable "key_name" {
  description = "Access keys"
}

variable "region" {
}

variable "availability_zones" {
}
variable "vpc_id" {
}

variable "instance_type" {
  description = "EC2 instance type to use"
  default     = "t2.nano"
}

variable "instance_tags" {
  default = {
    Name = "paddleball"
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
}

variable "health_ingress_cidr_block" {
    description = "Ingress block to allow load balancer health checks to reach through"
}
