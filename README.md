[![ci](https://github.com/env0/terratag/workflows/ci/badge.svg)](https://github.com/env0/terratag/actions?query=workflow%3Aci+branch%3Amaster)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fenv0%2Fterratag.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fenv0%2Fterratag?ref=badge_small)
# Terratag by env0
Terratag is a CLI tool allowing for tags or labels to be applied across an entire set of targeted Terraform files directory.  

> Terratag is brought to you with ❤️ by [env0](https://env0.com) -   
> <img src="logo.svg">  
> Let your team manage their own environment in AWS, Azure and Google. Governed by your policies and with complete visibility and cost management.      
  

## Prerequisites
- Terraform 0.11 or 0.12

## Usage
1. Download the latest [release binary](https://github.com/env0/terratag/releases) or install the latest [node package](https://github.com/env0/terratag/packages)  

1. Initialize Terraform modules to get provider schema and pull child modules:
   ```bash    
    terraform init  
    ```
1. Run Terratag  
      ```bash    
       terratag -dir=foo/bar -tags={\"hello\": \"world\"}
   ```    
   
   Terratag supports the following arguments:  
   - `-dir` - optional, the directory to recursively search for any `.tf` file and try to terratag it.  
   - `-tags` - tags, as valid JSON (NOT HCL)
   - `-skipTerratagFiles` - optional. Default to `true`. Skips any previously tagged - (files with `terratag.tf` suffix)


## Notes
- Resources already having the exact same tag as the one being appeneded will be overridden

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
All commands will be executed wherever down the heirarchy where `main.tf` is located.  
We do that to allow cases where complex nested submodule resolution may take place, and one would like to test how a directory higher up the heirarchy is resolved.  
- `expected` is a directory in which all `.terratag.tf` files will be matched with the output directory

#### What's being tested?
Each test will run:
- `terraform init`
- `terratag`
- `terraform validate`  

And finally, will compare the results in `out` with the `expected` directory 

#### Running Tests
Focus on a praticular Terraform version:
```
go test -run TestTerraformXX
``` 

### Release
```bash
# release tags have to start with v to trigger the release workflow -
# https://github.com/env0/terratag/blob/master/.github/workflows/release.yml
git tag vx.x.x 
git push --tags
```

