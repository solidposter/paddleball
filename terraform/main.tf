# Configure the AWS Provider
provider "aws" {
  region = var.region
}

data "aws_subnet_ids" "private" {
  vpc_id = var.vpc_id
  tags = {
    SubnetType = "Private"
  }
}

data "aws_ami" "latest_ecs" {
  most_recent = true
  owners = ["591542846629"] # AWS

  filter {
      name   = "name"
      values = ["*amazon-ecs-optimized"]
  }

  filter {
      name   = "virtualization-type"
      values = ["hvm"]
  }
}

resource "aws_security_group" "paddleball_sg" {
  name         = "paddleball-sg"
  description  = "Allow UDP inbound traffic and egress"
  vpc_id       = var.vpc_id

  ingress {
    from_port   = 2222
    to_port     = 2222
    protocol    = "udp"
    cidr_blocks = [var.ingress_cidr_block]
  }

  ingress {
    from_port   = 2220
    to_port     = 2220
    protocol    = "tcp"
    cidr_blocks = [var.health_ingress_cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "Paddleball Security Group"
  }
}

resource "aws_launch_configuration" "paddleball_lc" {
  image_id               = data.aws_ami.latest_ecs.id
  instance_type          = var.instance_type
  key_name               = var.key_name
  security_groups        = [aws_security_group.paddleball_sg.id]
  user_data              = var.user_data
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "paddleball_asg" {
  name                 = "paddleball-asg"
  launch_configuration = aws_launch_configuration.paddleball_lc.id
  availability_zones   = var.availability_zones
  target_group_arns    = [aws_alb_target_group.paddleball_alb_tg.arn]
  vpc_zone_identifier  = data.aws_subnet_ids.private.ids
  min_size             = 1
  max_size             = 2
  dynamic "tag" {
    for_each = var.instance_tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }

}

resource "aws_alb_target_group" "paddleball_alb_tg" {  
  name     = "paddleball-tg"
  port     = 2222
  protocol = "UDP"
  vpc_id   = var.vpc_id
  tags = {
    Name        = "PaddleBall"
    CostCenter  = "Platform Engineering"
    Environment = "Development"
  }
  health_check {
    port     = 2220
    protocol = "HTTP"
    path     = "/"
    matcher = "200-399"
  }
}

resource "aws_lb" "paddleball_lb" {
  name               = "paddleball-lb"
  load_balancer_type = "network"
  internal           = true
  ip_address_type    = "ipv4"
  subnets            = data.aws_subnet_ids.private.ids
}

resource "aws_lb_listener" "paddleball_listener" {
  load_balancer_arn = aws_lb.paddleball_lb.arn
  port              = "2222"
  protocol          = "UDP"

  default_action {
    type             = "forward"
    target_group_arn = aws_alb_target_group.paddleball_alb_tg.arn
  }
}