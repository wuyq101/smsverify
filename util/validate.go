package util

import "unicode"

func ValidatePhone(phone string) bool {
	if len(phone) != 11 {
		return false
	}
	for _, r := range phone {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
