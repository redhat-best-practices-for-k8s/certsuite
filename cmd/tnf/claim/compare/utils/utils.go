package utils

import "sort"

type Object interface {
	ID() string
	Data() interface{}
}

func GetMapFromObjectsList(objects []Object) map[string]interface{} {
	m := map[string]interface{}{}
	for _, o := range objects {
		m[o.ID()] = o.Data()
	}

	return m
}

func GetUniqueIDs(objects1, objects2 map[string]Object) []string {
	m := map[string]struct{}{}
	for ID := range objects1 {
		m[ID] = struct{}{}
	}

	for ID := range objects2 {
		m[ID] = struct{}{}
	}

	uniqueIDs := []string{}
	for ID := range m {
		uniqueIDs = append(uniqueIDs, ID)
	}

	sort.Strings(uniqueIDs)
	return uniqueIDs
}

func GetNotFoundObjects(objects1, objects2 map[string]Object) (notFoundInObjects1, notFoundInObjects2 []Object) {
	for ID, o := range objects1 {
		if _, found := objects2[ID]; !found {
			notFoundInObjects2 = append(notFoundInObjects2, o)
		}
	}

	for ID, o := range objects2 {
		if _, found := objects1[ID]; !found {
			notFoundInObjects1 = append(notFoundInObjects1, o)
		}
	}

	return
}
