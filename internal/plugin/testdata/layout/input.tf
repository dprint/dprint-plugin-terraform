# a comment
terraform {
required_version = ">= 1.0"
required_providers {
aws = {
source = "hashicorp/aws"
version = "~> 5.0"
}
}
}

resource "aws_s3_bucket" "b" {
    bucket = "my-bucket"   # trailing comment
    acl    =     "private"
    force_destroy = true


    lifecycle {
        prevent_destroy = true
    }
    policy = <<EOF
{
  "Version": "2012-10-17"
}
EOF
    count = var.enabled ? 1 : 0
    for_each = { for k, v in var.items : k => v if v != null }
    dynamic "rule" {
      for_each = var.rules
      content {
        id = rule.value.id
      }
    }
}

output "ids" {
  value = [for b in aws_s3_bucket.b : b.id]
  description="The IDs"
}
