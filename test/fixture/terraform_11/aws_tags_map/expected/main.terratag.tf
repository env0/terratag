provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = "${merge( map("Name","Mybucket","Unquoted1","Iwannabequoted","AnotherName","Yo","Unquoted2","Ireallywannabequoted","Unquoted3","IreallyreallywannabequotedandIgotacomma","${max(1,2)}","Testfunction","${local.localTagKey}","Testexpression","${local.localTagKey2}","Testvariableaskey","Yo-${local.localTagKey}","Testvariableinsidekey","Testvariableasvalue","${local.localTagKey}","Testvariableinsidevalue","Yo-${local.localTagKey}"), local.terratag_added_main)}"
}

resource "aws_s3_bucket" "a" {
  bucket = "my-another-tf-test-bucket"
  acl    = "private"

  tags = "${merge( "${local.localMap}", local.terratag_added_main)}"
}


locals {
  localTagKey  = "localTagKey"
  localTagKey2 = "localTagKey2"
  localMap = {
    key = "value"
  }
}
locals {
  terratag_added_main = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

