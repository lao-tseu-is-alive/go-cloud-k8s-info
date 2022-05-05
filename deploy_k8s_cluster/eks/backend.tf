terraform {
  backend "s3" {
    bucket = "doc-20220505103159"
    key    = "terraform/state"
  }
}
