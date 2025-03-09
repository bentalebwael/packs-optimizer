package utils

import "testing"

func TestCalculateArrayHash(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected string
	}{
		{
			name:     "Empty array",
			input:    []int{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Single element",
			input:    []int{1},
			expected: "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
		},
		{
			name:     "Multiple elements",
			input:    []int{1, 2, 3},
			expected: "8a6ae15122001229edb8866f56e342af12ae8187203c3e3b33931743e7c0c48d",
		},
		{
			name:     "Negative numbers",
			input:    []int{-1, 0, 1},
			expected: "134f7be29b90b375a969cd795e3e73037907515cc90435eb922aa77db7157447",
		},
		{
			name:     "Large numbers",
			input:    []int{1000000, 9999999},
			expected: "24038b7d9368f5f3ba7fb89c7fe783d6f62559b9c8ceb2c58e8090d62575f97f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateArrayHash(tt.input)
			if result != tt.expected {
				t.Errorf("CalculateArrayHash(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateArrayHash_Deterministic(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	// Calculate hash multiple times
	hash1 := CalculateArrayHash(input)
	hash2 := CalculateArrayHash(input)
	hash3 := CalculateArrayHash(input)

	// Verify all hashes are identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Hash function is not deterministic: got %v, %v, %v for same input", hash1, hash2, hash3)
	}
}
