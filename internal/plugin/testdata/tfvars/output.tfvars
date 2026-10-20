region         = "us-east-1"
instance_count = 3
name           = var.prefix
tags = {
  Environment = "dev"
  Owner       = "team"
}
