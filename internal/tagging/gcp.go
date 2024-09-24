package tagging

func tagContainerCluster(args TagBlockArgs) (*Result, error) {
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	tagBlock, err := TagBlock(rootBlockArgs)

	if err != nil {
		return nil, err
	}

	rootBlockSwappedTagsStrings := []string{tagBlock}

	return &Result{SwappedTagsStrings: rootBlockSwappedTagsStrings}, nil
}
