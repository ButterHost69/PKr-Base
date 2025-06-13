package utils

import (
	"math/rand"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
)

func CreateSlug() string {
	var gamerTag []string
	for range 1024 {
		gamerTag = append(gamerTag, fake.Gamertag())
	}
	g := rand.Intn(1024)
	return gamerTag[g]
}

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[r.Intn(len(letters))]
	}
	return string(s)
}
