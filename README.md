# [<img src="ttlogo.png" width="300" alt="Terratag Logo">](https://terratag.io) 
[![ci](https://github.com/env0/terratag/workflows/ci/badge.svg)](https://github.com/env0/terratag/actions?query=workflow%3Aci+branch%3Amaster) [![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fenv0%2Fterratag.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fenv0%2Fterratag?ref=badge_small)

> <sub>Terratag is brought to you with&nbsp;❤️&nbsp; by   
>[<img src="logo.svg" width="150">](https://env0.com)  
> Let your team manage their own environment in AWS, Azure and Google. <br/> 
> Governed by your policies and with complete visibility and cost management.      

## What?
Terratag is a CLI tool allowing for tags or labels to be applied across an entire set of Terraform files. Terratag will apply tags or labels to any AWS, GCP and Azure resources. 

### Terratag in action
![](https://assets.website-files.com/5dc3f52851595b160ba99670/5f62090d2d532ca35e143133_terratag.gif)

## Why?
Maintaining tags across your application is hard, especially when done manually. Terratag enables you to easily add tags to your existing IaC and benefit from some cross-resource tag applications you wish you had thought of when you had just started writing your Terraform, saving you tons of time and making future updates easy. [Read more](https://d1.awsstatic.com/whitepapers/aws-tagging-best-practices.pdf) on why tagging is important.

## How?
### Prerequisites
- Terraform 0.11, 0.12, 0.13 or 0.14

### Usage
1. Install from homebrew:
    ```
    ➜ brew install env0/terratag/terratag
    ```
    Or download the latest [release binary](https://github.com/env0/terratag/releases) .

1. Initialize Terraform modules to get provider schema and pull child modules:
   ```bash    
    terraform init  
    ```
1. Run Terratag  
      ```bash    
       terratag -dir=foo/bar -tags={\"environment_id\": \"prod\"}
   ```    
   
   Terratag supports the following arguments:  
   - `-dir` - optional, the directory to recursively search for any `.tf` file and try to terratag it.  
   - `-tags` - tags, as valid JSON (NOT HCL)
   - `-skipTerratagFiles` - optional. Default to `true`. Skips any previously tagged - (files with `terratag.tf` suffix)

### Example Output
#### Before Terratag
```
|- aws.tf
|- gcp.tf
```

```hcl
# aws.tf
provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags {
    Name        = "My bucket"
  }
}
```
```hcl
#gcp.tf
resource "google_storage_bucket" "static-site" {
  name          = "image-store.com"
  location      = "EU"
  force_destroy = true

  bucket_policy_only = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["http://image-store.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
  labels = {
    "foo" = "bar"
  }
}

```

#### After Terratag
Running `terratag -tags={\"env0_environment_id\":\"dev\",\"env0_project_id\":\"clientA\"}` will output:

```
|- aws.terratag.tf
|- gcp.terratag.tf
|- aws.tf.bak
|- gcp.tf.bak
```

```hcl
# aws.terratag.tf
provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = merge( map("Name", "My bucket" ), local.terratag_added_main)
}
locals {
  terratag_added_main = {"env0_environment_id"="dev","env0_project_id"="clientA"}
}
```
```hcl
# gcp.terratag.tf
resource "google_storage_bucket" "static-site" {
  name          = "image-store.com"
  location      = "EU"
  force_destroy = true

  bucket_policy_only = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["http://image-store.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
  labels = merge( map("foo" , "bar"), local.terratag_added_main)
}
locals {
  terratag_added_main = {"env0_environment_id"="dev","env0_project_id"="clientA"}
}
```

### Optional CLI flags

* `-dir=<path>` - defaults to `.`. Sets the terraform folder to tag `.tf` files in
* `-skipTerratagFiles=false` - Dont skip processing `*.terratag.tf` files (when running terratag a second time for the same directory)
* `-verbose=true` - Turn on verbose logging
* `-rename=false` - Instead of replacing files named `<basename>.tf` with `<basename>.terratag.tf`, keep the original filename

##### See more samples [here](https://github.com/env0/terratag/tree/master/test/fixture)

## Notes
- Resources already having the exact same tag as the one being appended will be overridden

## Develop
Issues and Pull Requests are very welcome!  

### Prerequisites
- Go > 1.13.5

### Build
```bash
git clone https://github.com/env0/terratag
go get
go build
```

### Test

#### Structure
The test suite will look for fixtures under `test/fixtures/terraform_xx`.  
Each fixture placed there should have the following directory structure:  
```
my_fixture
|+ input
  ...            // any depth under /input
     |- main.tf  // this is where we will run all terraform/terratag commands
|- expected
```

- `input` is where you should place the terraform files of your fixture.  
All commands will be executed wherever down the hierarchy where `main.tf` is located.  
We do that to allow cases where complex nested submodule resolution may take place, and one would like to test how a directory higher up the hierarchy gets resolved.  
- `expected` is a directory in which all `.terratag.tf` files will be matched with the output directory

#### What's being tested?
Each test will run:
- `terraform init`
- `terratag`
- `terraform validate`  

And finally, will compare the results in `out` with the `expected` directory 

#### Running Tests
Tests can only run on a specific Terraform version - 
```
go test -run TestTerraformXX
``` 

We use [tfenv](https://github.com/tfutils/tfenv) to switch between versions. The exact versions used in the CI tests can be found under `test/tfenvconf`.
