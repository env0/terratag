package common

import "github.com/hashicorp/hcl/v2/hclwrite"

type IACType string

const (
	Terraform  IACType = "terraform"
	Terragrunt IACType = "terragrunt"
)

type Version struct {
	Major int
	Minor int
}

type TaggingArgs struct {
	Filter              string
	Dir                 string
	Tags                string
	Matches             []string
	IsSkipTerratagFiles bool
	Rename              bool
	IACType             IACType
	TFVersion           Version
}

type TerratagLocal struct {
	Found map[string]hclwrite.Tokens
	Added string
}
