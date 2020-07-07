package tagging

func tagNodeConfigBlocks(args TagBlockArgs) []string {
	var swappedTagsStrings []string

	for _, block := range args.Block.Body().Blocks() {
		blockArgs := args
		blockArgs.Block = block

		if block.Type() == "node_config" {
			swappedTagsStrings = append(swappedTagsStrings, TagBlock(blockArgs))
		} else {
			swappedTagsStrings = append(swappedTagsStrings, tagNodeConfigBlocks(blockArgs)...)
		}
	}

	return swappedTagsStrings
}

func tagContainerCluster(args TagBlockArgs) Result {
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	rootBlockSwappedTagsStrings := []string{TagBlock(rootBlockArgs)}

	return Result{SwappedTagsStrings: append(rootBlockSwappedTagsStrings, tagNodeConfigBlocks(args)...)}
}

func tagContainerNodePool(args TagBlockArgs) Result {
	return Result{SwappedTagsStrings: tagNodeConfigBlocks(args)}
}
