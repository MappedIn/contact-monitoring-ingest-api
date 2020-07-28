package invitecode

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var numbers = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GenerateNumberString returns a pseudo random alpha-numeric
// uppercase string
func GenerateNumberString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}
