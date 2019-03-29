package middleware

import "strings"

func mergeStringSlice(s1 []string, s2 []string) []string {
	mergedList := s1
	for _, el := range s2 {
		if !contains(s1, el) {
			mergedList = append(mergedList, el)
		}
	}

	return mergedList
}

func contains(stringList []string, element string) bool {
	for _, el := range stringList {
		if strings.ToLower(el) == strings.ToLower(element) {
			return true
		}
	}
	return false
}
