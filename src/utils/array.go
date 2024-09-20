package utils

func Find[T any](arr []T, predicate func(T) bool) (T,bool) {
	for _, v := range arr {
		if predicate(v) {
			return v, true
		}
	}

	var zeroValue T
	return zeroValue, false 
}