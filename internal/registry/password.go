package registry

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"magnetron/internal/config"
	"syscall"
)

func EncryptPassword(password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hash), err
}

func PromptUserForPassword() (string, error) {
	fmt.Print("Enter password: ")

	bytePassword, err := term.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}

func CheckPassword(password string, passwordConfig config.PasswordConfig) bool {

	for _, entry := range passwordConfig.PasswordEntries {
		if err := bcrypt.CompareHashAndPassword([]byte(entry.Password), []byte(password)); err == nil {
			return true
		}
	}

	return false
}
