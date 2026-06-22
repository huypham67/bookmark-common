package sliceutil

import (
	"reflect"
	"testing"
)

func TestSplitIntoBatches(t *testing.T) {
	cases := []struct {
		name  string
		input []int
		size  int
		want  [][]int
	}{
		{
			name:  "empty input returns no batches",
			input: []int{},
			size:  3,
			want:  [][]int{},
		},
		{
			name:  "nil input returns no batches",
			input: nil,
			size:  3,
			want:  [][]int{},
		},
		{
			name:  "fewer than size fits in one batch",
			input: []int{1, 2},
			size:  3,
			want:  [][]int{{1, 2}},
		},
		{
			name:  "exactly size fits in one batch",
			input: []int{1, 2, 3},
			size:  3,
			want:  [][]int{{1, 2, 3}},
		},
		{
			name:  "more than size splits with remainder",
			input: []int{1, 2, 3, 4, 5},
			size:  2,
			want:  [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:  "evenly divisible splits cleanly",
			input: []int{1, 2, 3, 4},
			size:  2,
			want:  [][]int{{1, 2}, {3, 4}},
		},
		{
			name:  "zero size returns single batch",
			input: []int{1, 2, 3},
			size:  0,
			want:  [][]int{{1, 2, 3}},
		},
		{
			name:  "negative size returns single batch",
			input: []int{1, 2, 3},
			size:  -1,
			want:  [][]int{{1, 2, 3}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SplitIntoBatches(c.input, c.size)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("SplitIntoBatches(%v, %d) = %v, want %v", c.input, c.size, got, c.want)
			}
		})
	}
}

// TestSplitIntoBatchesCapacityIsCapped guards against a classic append bug:
// each batch must have its capacity capped to its length, so appending to one
// batch cannot overwrite the first element of the next.
func TestSplitIntoBatchesCapacityIsCapped(t *testing.T) {
	batches := SplitIntoBatches([]int{1, 2, 3, 4}, 2)
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(batches))
	}

	batches[0] = append(batches[0], 99)

	if batches[1][0] != 3 {
		t.Errorf("appending to batch 0 corrupted batch 1: got %v, want first element 3", batches[1])
	}
}
