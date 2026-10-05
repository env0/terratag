resource "aws_s3_bucket" "untagged" {
  bucket = "terratag-tofu-untagged"
}

resource "aws_s3_bucket" "tagged" {
  bucket = "terratag-tofu-tagged"

  tags = {
    a = "b"
  }
}
