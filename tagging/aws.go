package tagging

import "github.com/env0/terratag/convert"

func tagAutoscalingGroup(args TagBlockArgs) Result {
	// https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html
	var foundTagBlock bool
	for _, block := range args.Block.Body().Blocks() {
		if block.Type() == "tag" {
			foundTagBlock = true
			break
		}
	}
	if foundTagBlock {
		convert.AppendTagBlocks(args.Block, args.Tags)
	} else {
		ConcatTagToTagsBlock(args)
	}
	return Result{}
}
