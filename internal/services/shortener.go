package services

import (
	"Linux-url-shortener/internal/database"
	"crypto/rand"
	"errors"

	"math/big"
)

var ErrMaxCodeGenerationAttempts = errors.New(
	"maximum shortcode generation attempts exceeded",
)

const DefaultMaxAttempts = 10

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

func GenerateUniqueCode(repo database.Repository, url string, maxAttempts int) (string, error) {

	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}

	for attempts := 0; attempts < maxAttempts; attempts++ {
		code := GenerateCode(6)

		err := repo.SaveUrl(code, url)

		if err != nil {
			if errors.Is(err, database.ErrShortCodeExist) {
				continue
			}

			return "", err
		}
		return code, nil
	}

	return "", ErrMaxCodeGenerationAttempts
}
