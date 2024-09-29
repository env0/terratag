package tagging

func tagAksK8sCluster(args TagBlockArgs) (*Result, error) {
	var swappedTagsStrings []string

	// handle root block tags attribute
	tagBlock, err := TagBlock(args)
	if err != nil {
		return nil, err
	}

	swappedTagsStrings = append(swappedTagsStrings, tagBlock)

	// handle default_node_pool tags attribute
	nodePool := args.Block.Body().FirstMatchingBlock("default_node_pool", nil)
	if nodePool != nil {
		args.Block = nodePool

		tagBlock, err := TagBlock(args)
		if err != nil {
			return nil, err
		}

		swappedTagsStrings = append(swappedTagsStrings, tagBlock)
	}

	return &Result{SwappedTagsStrings: swappedTagsStrings}, nil
}
