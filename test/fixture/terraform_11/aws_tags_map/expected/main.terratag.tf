provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = "${merge( map("Name"                    , "My bucket" ,    "Unquoted1"                 , "I wanna be quoted" ,    "AnotherName"             , "Yo" ,    "Unquoted2"                 , "I really wanna be quoted" ,    "Unquoted3"                 , "I really really wanna be quoted and I got a comma",    "${max(1, 2)}"            , "Test function" ,    "${local.localTagKey}"    , "Test expression"), local.terratag_added_main)}"
}

locals {
  localTagKey = "localTagKey"
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

