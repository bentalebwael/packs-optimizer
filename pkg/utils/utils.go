package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// CalculateArrayHash takes an array of integers and returns its SHA256 hash as a string.
// The function converts the array to a deterministic string representation before hashing.
func CalculateArrayHash(arr []int) string {
	// Convert integers to strings and join them with a delimiter
	// Using comma as delimiter since it's not typically part of integer representations
	elements := make([]string, len(arr))
	for i, num := range arr {
		elements[i] = strconv.Itoa(num)
	}
	str := strings.Join(elements, ",")

	// Calculate SHA256 hash
	hasher := sha256.New()
	hasher.Write([]byte(str))
	hash := hasher.Sum(nil)

	// Convert hash to hexadecimal string
	return hex.EncodeToString(hash)
}
