package utils

import "encoding/json"

func SortMap(tagsMap map[string]string) {
	// json.Marshal sorts the map by its keys lexicography
	// https://golang.org/pkg/encoding/json/#Marshal
	marshal, _ := json.Marshal(tagsMap)
	json.Unmarshal(marshal, &tagsMap)
}
