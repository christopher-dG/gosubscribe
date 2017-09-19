package gosubscribe

import (
	"errors"
	"fmt"
	"log"

	randomdata "github.com/Pallinder/go-randomdata"
)

// GetSecret retrieve's a user's secret.
func (user *User) GetSecret() (string, error) {
	if len(user.Secret) == 0 {
		log.Printf("User %d does not have a secret\n", user.ID)
		return "", errors.New("user does not have a secret")
	}
	return user.Secret, nil
}

// UserFromSecret retrieves a user from their unique secret.
func UserFromSecret(secret string) (*User, error) {
	user := new(User)
	DB.Where("secret = ?", secret).First(user)
	if user.ID == 0 {
		return nil, errors.New("secret does not exist")
	}
	return user, nil
}

// GenSecret generates a "random" string that is guaranteed to not already be used.
// Note: This is not even remotely cryptographically secure, but it'll do for our needs.
func GenSecret() string {
	for { // Retry until we get something unique (more than one try should be very rare).
		s := fmt.Sprint(
			randomdata.SillyName(),
			randomdata.SillyName(),
			randomdata.SillyName(),
		)
		_, err := UserFromSecret(s)
		if err != nil { // Secret does not exist.
			return s
		}
	}
}
