package utils

func Find[T any](arr []T, predicate func(T) bool) (T,bool) {
	for _, v := range arr {
		if predicate(v) {
			return v, true
		}
	}

	var val T
	return val, false 
}

func FindIndex[T any](arr []T, predicate func(T) bool) int {
	index := -1

	for i, v := range arr {
		if predicate(v) {
			index = i
		}
	}

	return index
}


func Includes[T comparable](arr []T, value T) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}

	return false
}