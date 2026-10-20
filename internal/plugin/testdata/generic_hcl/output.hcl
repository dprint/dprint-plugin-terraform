source "amazon-ebs" "example" {
  ami_name      = "${var.prefix}"
  instance_type = "t2.micro"
}

job example {
  datacenters = ["dc1"]
  group "cache" {
    count = 1
  }
}

variable "a" {
  type = list
}
