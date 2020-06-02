
# Terratag
Add tags to your AWS resources in both Terraform 0.11 and 0.12!

## Prerequisites
- [tfschema](https://github.com/minamijoyo/tfschema) to assert which resources support `tags` in AWS Provider
- Terraform 0.11 or 0.12

## Usage
1. Download the latest [release binary](https://github.com/env0/terratag/releases) or install the latest [node package](https://github.com/env0/terratag/packages)
2. ```bash    
    terraform init # needed to initialize provider schema and pull child terraform modules
    terratag -dir=foo/bar -tags={\"hello\": \"world\"}
    ```
      > Note Terratag receives two command line arguments:  
      > - `-dir` - optional, the directory to recursively search for any `.tf` file and try to terratag it.  
      > - `-tags` - tags, as valid JSON (NOT HCL)
      > - `-skipTerratagFiles` - optional. Default to `true`. Skips any previously tagged - (files with `terratag.tf` suffix)


## Notes
- Only AWS resources are supported (for now)
- Resources already having the exact same tag as the one being appeneded will be overridden

## How's it different from [env0/terratag.js](https://github.com/env0/terratag.js)?
- Multi version support! Terratag works on Terraform 0.11 as well as 0.12
- `terratag.js` relies on HCL to JSON translation to traverse and manipulate the HCL.    
Such conversion is [unsafe](https://github.com/hashicorp/terraform/issues/9354#issuecomment-512624185), and becomes even more fragile and difficult in HCL2 (introduced in Terraform 12).  
Instead, `terratag` uses HashiCorps' [`hclwrite`](https://godoc.org/github.com/hashicorp/terraform/vendor/github.com/hashicorp/hcl/v2/hclwrite) to surgically manipulate the tags attribute directrly in HCL 2.   

## Common Errors
- ```
  Failed to NewClient: Failed to find plugin: aws. Plugin binary was not found in any of the following directories
  ```  
  `tfschema` can't find your AWS Provider schema - You probably didn't run `terraform init`

## Develop

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

## TODO
- [ ] Support for resource block nested in for loops (?)
- [ ] Add godocs
- [ ] Add unit tests
- [ ] Remove `tfschema` as a prerequsite install
- [x] Automate and publish package release binaries
- [ ] Support tagging resources in `.tf.json` files (?)
