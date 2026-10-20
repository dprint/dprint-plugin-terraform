variables {
  bucket_prefix = "test"
  name = "${var.base}"
}

run "valid_name" {
  command = plan
  assert {
    condition = aws_s3_bucket.b.bucket == "${var.bucket_prefix}-bucket"
    error_message = "bad name"
  }
}
