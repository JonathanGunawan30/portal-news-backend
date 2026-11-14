package conv

import (
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSlug(title string) string {
	var b strings.Builder
	prevDash := false

	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' || !unicode.IsLetter(r) {
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		} else {
			b.WriteRune(r)
			prevDash = false
		}
	}

	return strings.TrimRight(b.String(), "-")
}

func StringToInt64(str string) (int64, error) {
	result, err := strconv.ParseInt(str, 10, 64)
	return result, err
}

func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func StringToInt(str string) (int, error) {
	num, err := strconv.Atoi(str)
	return num, err
}
