package util

import "math/rand"

func RandomSample(letters string, n int) string {
	b := make([]byte, n)
	llen := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(llen)]
	}
	return string(b)
}
