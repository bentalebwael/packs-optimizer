package postgres

import (
	"github.com/lib/pq"
)

func Int64ArrayToIntSlice(int64Array pq.Int64Array) []int {
	intSlice := make([]int, len(int64Array))
	for i, v := range int64Array {
		intSlice[i] = int(v)
	}
	return intSlice
}

func IntSliceToPqArray(intSlice []int) pq.Int64Array {
	int64Array := make(pq.Int64Array, len(intSlice))
	for i, v := range intSlice {
		int64Array[i] = int64(v)
	}
	return int64Array
}
