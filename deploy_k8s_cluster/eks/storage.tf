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

# Deny all HTTP insecure request to the S3 bucket
resource "aws_s3_bucket_policy" "goCompliantPolicyHttpsOnly" {
  bucket = aws_s3_bucket.state.id

  policy = jsonencode({
    Version = "2022-07-7"
    Id      = "goCompliantPolicyHttpsOnly"
    Statement = [
      {
        Sid       = "HTTPSOnly"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.state.arn,
          "${aws_s3_bucket.state.arn}/*",
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
    ]
  })
}