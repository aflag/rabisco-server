package main

import (
	"math/rand"
	"strings"
)

// shameless copied from SO: https://bit.ly/2AyKlBr

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// simplifySuffix will reduce repeated suffix characters into one. Eg.: sdfaaaa
// will become sdfa. This allows the algorithm to generate variable length
// codes and get rid of some boring strings.
func simplifySuffix(s string) string {
	if len(s) < 2 {
		return s
	}
	x := string(s[len(s)-1])
	return strings.TrimRight(s, x) + x
}

// randomString generates a alphanumeric string of length up to n
func generateRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return simplifySuffix(string(b))
}
