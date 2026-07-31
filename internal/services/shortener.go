package services

import (
	"Linux-url-shortener/internal/database"
	"crypto/rand"

	"math/big"
)


func GenerateCode(lenght int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 6)

	for i := range b {
		num , _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}

	return string(b)
}

func GenerateUniqueCode(repo database.Repository) (string, error){
	for {
		code := GenerateCode(6)

		exists, err := repo.ShortCodeExist(code)

		if err != nil{
			return "", err
		}

		if !exists {
			return code, nil
		}
	}
}
