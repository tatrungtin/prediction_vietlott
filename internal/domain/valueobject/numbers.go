package valueobject

import (
	"fmt"
	"sort"
)

// Numbers represents a set of 6 unique lottery numbers
type Numbers []int

// NewNumbers creates a new Numbers value object with validation
func NewNumbers(nums []int) (Numbers, error) {
	if len(nums) != 6 {
		return nil, fmt.Errorf("must have exactly 6 numbers, got %d", len(nums))
	}

	// Validate range and uniqueness
	seen := make(map[int]bool)
	for _, n := range nums {
		if n < 1 || n > 55 {
			return nil, fmt.Errorf("numbers must be between 1-55, got %d", n)
		}
		if seen[n] {
			return nil, fmt.Errorf("numbers must be unique, duplicate found: %d", n)
		}
		seen[n] = true
	}

	// Sort and return as a copy
	sorted := make(Numbers, 6)
	copy(sorted, nums)
	sort.Ints(sorted)

	return sorted, nil
}

// MustNewNumbers creates a Numbers value object and panics on error
// Useful for tests with known valid data
func MustNewNumbers(nums []int) Numbers {
	n, err := NewNumbers(nums)
	if err != nil {
		panic(err)
	}
	return n
}

// MatchCount returns the count of numbers that match between two Numbers sets
func (n Numbers) MatchCount(other Numbers) int {
	count := 0
	nSet := make(map[int]bool)
	for _, num := range n {
		nSet[num] = true
	}

	for _, num := range other {
		if nSet[num] {
			count++
		}
	}
	return count
}

// Contains checks if a number is present in the set
func (n Numbers) Contains(num int) bool {
	for _, v := range n {
		if v == num {
			return true
		}
	}
	return false
}

// Sum returns the sum of all numbers
func (n Numbers) Sum() int {
	sum := 0
	for _, num := range n {
		sum += num
	}
	return sum
}

// AsSlice returns the numbers as a slice
func (n Numbers) AsSlice() []int {
	return []int(n)
}

// String returns a string representation of the numbers
func (n Numbers) String() string {
	result := "["
	for i, num := range n {
		if i > 0 {
			result += ", "
		}
		if num < 10 {
			result += fmt.Sprintf("0%d", num)
		} else {
			result += fmt.Sprintf("%d", num)
		}
	}
	result += "]"
	return result
}
