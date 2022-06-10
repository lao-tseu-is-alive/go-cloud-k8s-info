resource "aws_s3_bucket" "state" {
  bucket        = var.state_bucket
}

resource "aws_s3_bucket_acl" "acl" {
  bucket    = aws_s3_bucket.state.id
  acl       = "private"
}

resource "aws_s3_bucket_public_access_block" "state-bucket-public-access-block" {
  bucket = aws_s3_bucket.state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}