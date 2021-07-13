package utils

// GetSlicesIntersect returns the intersect of s1 and s2
func GetSlicesIntersect(s1, s2 []string) (intersect []string) {
	for i := 0; i < len(s1); i++ {
		if Contains(s2, s1[i]) {
			intersect = append(intersect, s1[i])
		}
	}
	return intersect
}

// LimitStringSlice limits the slice of strings to a maximum length
func LimitStringSlice(slice []string, limit uint) []string {
	if len(slice) <= int(limit) {
		return slice
	}

	return slice[:limit]
}

// Contains returns true if s is in the slice
func Contains(slice []string, s string) bool {
	for _, a := range slice {
		if a == s {
			return true
		}
	}
	return false
}
