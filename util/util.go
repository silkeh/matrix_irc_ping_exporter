package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandString returns a random string of a specified length.
// Only lower case letters are used.
func RandString(len int) (s string) {
	for i := 0; i < len; i++ {
		s += string(rand.Intn(26) + 97)
	}
	return
}
