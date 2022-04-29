package tagging

func tagAksK8sCluster(args TagBlockArgs) Result {
	var swappedTagsStrings []string

	// handle root block tags attribute
	swappedTagsStrings = append(swappedTagsStrings, TagBlock(args))

	// handle default_node_pool tags attribute
	nodePool := args.Block.Body().FirstMatchingBlock("default_node_pool", nil)
	if nodePool != nil {
		args.Block = nodePool
		swappedTagsStrings = append(swappedTagsStrings, TagBlock(args))
	}

	return Result{SwappedTagsStrings: swappedTagsStrings}
}
