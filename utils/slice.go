package utils

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func RemoveIndex(list interface{}, index int) {
	//copy(list[index:], list[index+1:])
	//list = list[:len(list)-1]
}
