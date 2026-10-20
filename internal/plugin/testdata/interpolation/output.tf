locals {
  name    = var.name
  joined  = "${var.a}-${var.b}"
  nested  = lookup(var.map, "${var.key}")
  literal = "prefix-${var.x}"
  spaced  = var.x
  conditional = (var.enabled
    ? "yes"
  : "no")
  already_parens = (
    var.a
  )
  plain = "text"
  list  = ["${var.one}", "${var.two}"]
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  tags = {
    Name = "${var.name}"
  }
}
