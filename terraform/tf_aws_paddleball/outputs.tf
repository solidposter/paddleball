output "paddleball_endpoint" {
  value = "paddleball.${data.aws_route53_zone.selected.name}"
}