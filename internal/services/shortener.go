package services

import (
	"Linux-url-shortener/internal/database"
	"crypto/rand"
	"os"

	"math/big"
)

var maxAttempts = os.Getenv("maxAttempts")

func GenerateCode(lenght int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, lenght)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))

		if err != nil {
			panic(err)
		}

		b[i] = charset[num.Int64()]
	}

	return string(b)
}

func GenerateUniqueCode(repo database.Repository) (string, error) {

	for {
		code := GenerateCode(6)

		exists, err := repo.ShortCodeExist(code)

		if err != nil {
			return "", err
		}

		if !exists {
			return code, nil
		}
	}
}
