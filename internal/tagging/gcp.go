package tagging

func tagContainerCluster(args TagBlockArgs) Result {
	rootBlockArgs := args
	rootBlockArgs.TagId = "resource_labels"
	rootBlockSwappedTagsStrings := []string{TagBlock(rootBlockArgs)}

	return Result{SwappedTagsStrings: rootBlockSwappedTagsStrings}
}
