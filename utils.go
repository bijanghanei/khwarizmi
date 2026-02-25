package main

import "strings"

// normalizeProblemInput converts either a full LeetCode URL or a slug
// into a valid titleSlug.
//
// Examples:
//   https://leetcode.com/problems/coin-change/ → coin-change
//   coin-change → coin-change
func normalizeProblemInput(input string) string {
	input = strings.TrimSpace(input)

	// If it's a URL
	if strings.Contains(input, "leetcode.com") {
		parts := strings.Split(input, "/")
		for i := 0; i < len(parts); i++ {
			if parts[i] == "problems" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	// Assume it's already a slug
	return input
}