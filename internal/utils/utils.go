package utils

import "sort"

func SortObjectKeys(tagsMap map[string]string) []string {
	var keys []string
	for key, _ := range tagsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
