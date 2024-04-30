package random

import "math/rand"

const (
	defaultLenOfAlias = 10
)

func Alias(lens ...int) string {
	var length int

	if len(lens) == 0 || len(lens) > 1 {
		length = defaultLenOfAlias
	} else {
		length = lens[0]
	}

	result := make([]byte, length)

	symbols := "abcdefghigklmnopqrstuvwxyzABCDEFGHIGKLMNOPQRSTUVWXYZ1234567890"

	for i := range result {
		result[i] = symbols[rand.Int()%length]
	}

	return string(result)
}
