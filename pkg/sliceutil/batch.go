// Package sliceutil provides small, generic helpers for working with slices.
package sliceutil

// SplitIntoBatches splits objects into consecutive batches of at most size
// elements. The final batch holds the remainder and may be smaller.
//
// It is useful when fanning a large slice out to a queue or a bulk database
// write, where sending one batch per message keeps the number of round-trips
// bounded instead of growing with the number of elements.
//
// Edge cases:
//   - An empty input returns an empty (non-nil) result with no batches.
//   - A non-positive size returns a single batch containing all objects, so
//     callers never have to guard against a zero size separately.
//
// The returned sub-slices share the backing array of objects (using a full
// three-index slice expression to cap their capacity), so callers must not
// append to a batch expecting it to be independent.
func SplitIntoBatches[T any](objects []T, size int) [][]T {
	if len(objects) == 0 {
		return [][]T{}
	}
	if size <= 0 {
		return [][]T{objects}
	}

	batches := make([][]T, 0, (len(objects)+size-1)/size)
	for size < len(objects) {
		objects, batches = objects[size:], append(batches, objects[0:size:size])
	}
	batches = append(batches, objects)

	return batches
}
