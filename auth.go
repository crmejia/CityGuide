package guide

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"regexp"
)

type user struct {
	Id                        int64
	Username, Password, Email string
}

func newUser(username, password, confirmPassword, email string) (user, error) {
	if username == "" {
		return user{}, errors.New("username cannot be empty")
	}
	if password == "" {
		return user{}, errors.New("password cannot be empty")
	}
	if len(password) < 8 {
		return user{}, errors.New("password has to be at least 8 characters long")
	}
	if password != confirmPassword {
		return user{}, errors.New("passwords do not match")
	}
	if email == "" {
		return user{}, errors.New("email cannot be empty")
	}
	match := rxEmail.Match([]byte(email))
	if !match {
		return user{}, errors.New("email has to be a valid address")
	}

	salt, err := generateSalt(saltSize)
	hashedPassword := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedPassword := base64.RawStdEncoding.EncodeToString(hashedPassword)
	encodedHash := fmt.Sprintf("%s%s", encodedSalt, encodedPassword)
	if err != nil {
		return user{}, err
	}
	u := user{
		Username: username,
		Password: encodedHash,
		Email:    email,
	}
	return u, nil
}

func generateSalt(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

var rxEmail = regexp.MustCompile(".+@.+\\..+")

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 24
	argon2KeyLen  = 32
	saltSize      = 16
)
