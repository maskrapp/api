package utils

import (
	"math/rand"
	"strings"
)

var set = "1234567890"

func GenerateCode(length int) string {

	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(set[rand.Intn(len(set))])
	}
	return sb.String()
}
