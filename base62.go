package main

import (
	"errors"
	"strings"
)

// Characters for base62 encoding
const CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// base62Encode encodes a number using base62 encoding.
func base62Encode(num int) string {
	if num == 0 {
		return string(CHARS[0])
	}
	var encoding strings.Builder
	for num > 0 {
		remainder := num % 62
		num /= 62
		encoding.WriteString(string(CHARS[remainder]))
	}

	// Reverse the result to match the desired encoding
	encoded := encoding.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// base62Decode decodes a base62-encoded string to its original number.
func base62Decode(encoded string) (int, error) {
	result := 0
	base := len(CHARS)

	for _, char := range encoded {
		index := strings.IndexRune(CHARS, char)
		if index == -1 {
			return 0, errors.New("invalid character in base62 string")
		}
		result = result*base + index
	}

	return result, nil
}
