module github.com/env0/terratag

go 1.13

require (
	github.com/bmatcuk/doublestar v1.2.2
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/minamijoyo/tfschema v0.4.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/onsi/gomega v1.10.1
	github.com/otiai10/copy v1.2.0
	github.com/thoas/go-funk v0.5.0
	github.com/zclconf/go-cty v1.5.1
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
)

replace github.com/minamijoyo/tfschema => github.com/env0/tfschema v0.4.2-0.20200831101540-6e1bb6162348
