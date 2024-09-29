package utils

import "sort"

func SortObjectKeys(tagsMap map[string]string) []string {
	keys := []string{}

	for key := range tagsMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
