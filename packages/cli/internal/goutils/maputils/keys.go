package maputils

import "sort"

func GetKeys[V any](mp map[string]V) []string {
	var keys = []string{}

	for key := range mp {
		keys = append(keys, key)
	}

	return keys
}

func GetSortedKeys[V any](mp map[string]V) []string {
	var keys = GetKeys(mp)
	sort.Strings(keys)
	return keys
}
