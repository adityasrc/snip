package utils

import ()

const base62Digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func ConvertToBase62(id int) string {

	if id == 0 {
		return string(base62Digits[0])
	}

	var base62 []byte // byte is an alias for unit8 (0-255)

	// conversion math
	for id != 0 {
		rem := id % 62
		base62 = append(base62, base62Digits[rem]) // appending remainder
		id = id / 62
	}

	// reverse the result
	for i, j := 0, len(base62); i < j; i, j = i+1, j-1 {
		base62[i], base62[j] = base62[j], base62[i]
	}

	return string(base62)
}
