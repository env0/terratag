package tagging

import "github.com/env0/terratag/convert"

func tagAutoscalingGroup(args TagBlockArgs) Result {
	// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
	var found bool
	for _, block := range args.Block.Body().Blocks() {
		if block.Type() == "tag" {
			found = true
			break
		}
	}
	if found {
		convert.AppendTagBlocks(args.Block, args.Tags)
	} else {
		ConcatTagsToTagsBlock(args)
	}
	return Result{}
}
