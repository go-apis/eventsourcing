package types

func IndexOf[T any](slice []T, match func(T) bool) int {
	for i, v := range slice {
		if match(v) {
			return i
		}
	}
	return -1
}
