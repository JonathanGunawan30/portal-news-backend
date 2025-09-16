package conv

import (
	"strconv"
	"strings"

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
	return strings.ReplaceAll(strings.ToLower(title), " ", "-")
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
