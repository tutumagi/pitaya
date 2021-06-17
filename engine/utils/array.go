package utils

//WashShuffleArray 数组重排
func WashShuffleArray(items []int) []int {
	if len(items) < 1 {
		return items
	}
	for i := 0; i < len(items); i++ {
		idx := int(RandomInt32(int32(i), int32(len(items))))
		items[i], items[idx] = items[idx], items[i]
	}
	return items
}
