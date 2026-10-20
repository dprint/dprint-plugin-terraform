variable "a" {
  type = list(any)
}

variable "b" {
  type = map(any)
}

variable "c" {
  type = set(any)
}

variable "d" {
  type = string
}

variable "e" {
  type = list(string)
}

variable "f" {
  type = map(string)
}

variable "g" {
  type = list(string)
}

variable "h" {
  type = "number"
}

module "not_a_variable" {
  source = "./mod"
  type   = list
}

variable "nested" {
  validation {
    type          = list
    condition     = true
    error_message = "x"
  }
}
