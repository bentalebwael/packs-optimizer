package postgres

import (
	"reflect"
	"testing"

	"github.com/lib/pq"
)

func TestInt64ArrayToIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    pq.Int64Array
		expected []int
	}{
		{
			name:     "empty array",
			input:    pq.Int64Array{},
			expected: []int{},
		},
		{
			name:     "single element",
			input:    pq.Int64Array{1},
			expected: []int{1},
		},
		{
			name:     "multiple elements",
			input:    pq.Int64Array{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "large numbers",
			input:    pq.Int64Array{1000000, 2000000},
			expected: []int{1000000, 2000000},
		},
		{
			name:     "negative numbers",
			input:    pq.Int64Array{-1, -2, -3},
			expected: []int{-1, -2, -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64ArrayToIntSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Int64ArrayToIntSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIntSliceToPqArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected pq.Int64Array
	}{
		{
			name:     "empty array",
			input:    []int{},
			expected: pq.Int64Array{},
		},
		{
			name:     "single element",
			input:    []int{1},
			expected: pq.Int64Array{1},
		},
		{
			name:     "multiple elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: pq.Int64Array{1, 2, 3, 4, 5},
		},
		{
			name:     "large numbers",
			input:    []int{1000000, 2000000},
			expected: pq.Int64Array{1000000, 2000000},
		},
		{
			name:     "negative numbers",
			input:    []int{-1, -2, -3},
			expected: pq.Int64Array{-1, -2, -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntSliceToPqArray(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("IntSliceToPqArray() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConversionRoundTrip(t *testing.T) {
	testCases := [][]int{
		{},
		{1},
		{1, 2, 3, 4, 5},
		{-1, 0, 1},
		{1000000, 2000000},
	}

	for _, input := range testCases {
		t.Run("round trip conversion", func(t *testing.T) {
			// Convert to pq.Int64Array and back
			pqArray := IntSliceToPqArray(input)
			result := Int64ArrayToIntSlice(pqArray)

			// Verify the round trip preserved the values
			if !reflect.DeepEqual(input, result) {
				t.Errorf("Round trip conversion failed: input %v, got %v", input, result)
			}
		})
	}
}
