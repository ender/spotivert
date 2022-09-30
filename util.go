package main

func min[T float64 | int](a, b T) T {
	if a <= b {
		return a
	}
	return b
}

func splitArray[T any](elems []T, size int) (batch [][]T) {
	for i := 0; i < len(elems); i += size {
		batch = append(batch, elems[i:min(i+size, len(elems))])
	}
	return batch
}
