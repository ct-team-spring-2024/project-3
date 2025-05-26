package internal
func merge(left, right []Pair) []Pair {
	result := make([]Pair, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i].Key < right[j].Key {
			result = append(result, left[i])
			i++
		} else if left[i].Key > right[j].Key {
			result = append(result, right[j])
			j++
		} else {
			// If keys are equal, prefer the right-side Pair (newer data)
			result = append(result, right[j])
			j++
		}
	}

	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}
