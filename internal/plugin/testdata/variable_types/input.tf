variable "a" {
  type = list
}

variable "b" {
  type = map
}

variable "c" {
  type = set
}

variable "d" {
  type = "string"
}

variable "e" {
  type = "list"
}

variable "f" {
  type = "map"
}

variable "g" {
  type = list(string)
}

variable "h" {
  type = "number"
}

module "not_a_variable" {
  source = "./mod"
  type = list
}

variable "nested" {
  validation {
    type = list
    condition = true
    error_message = "x"
  }
}
