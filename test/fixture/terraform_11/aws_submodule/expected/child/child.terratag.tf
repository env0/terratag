resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = "${merge( map("Name"        , "My bucket" ,    "Environment" , "Dev"), local.terratag_added_child)}"
}
locals {
  terratag_added_child = {"env0_environment_id"="40907eff-cf7c-419a-8694-e1c6bf1d1168","env0_project_id"="43fd4ff1-8d37-4d9d-ac97-295bd850bf94"}
}

