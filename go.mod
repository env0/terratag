module github.com/env0/terratag

go 1.13

require (
	github.com/bmatcuk/doublestar v1.2.2
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/logutils v1.0.0
	github.com/minamijoyo/tfschema v0.5.0
	github.com/onsi/gomega v1.10.1
	github.com/otiai10/copy v1.2.0
	github.com/thoas/go-funk v0.5.0
	github.com/zclconf/go-cty v1.5.1
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)

// remove the following directive when https://github.com/hashicorp/hcl/issues/402 gets fixed

replace github.com/hashicorp/hcl/v2 => github.com/env0/hcl/v2 v2.2.1-0.20201012055633-9ccfb031dba0
