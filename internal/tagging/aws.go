package tagging

import "github.com/env0/terratag/internal/convert"

func tagAutoscalingGroup(args TagBlockArgs) Result {
	// for now, we count on it that if there's a single "tag" in the schema (unlike "tags" block),
	// then no "tags" interpolation is used, but rather multiple instances of a "tag" block
	// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
	convert.AppendTagBlocks(args.Block, args.Tags)

	return Result{}
}
