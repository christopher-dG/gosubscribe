package gosubscribe

import (
	"errors"
	"fmt"
	"hash/fnv"
	"time"
)

// GetSecret retrieve's a user's secret.
func GetSecret(user User) (string, error) {
	if len(user.Secret) != 0 {
		return user.Secret, nil
	} else {
		return "", errors.New("You don't have a secret; run `.init` to get one.")
	}
}

// UserFromSecret retrieves a user from their unique secret.
func UserFromSecret(secret string) (User, error) {
	var user User
	DB.Where("secret = ?", secret).First(&user)
	if user.ID == 0 {
		return user, errors.New("Incorrect secret.")
	}
	return user, nil
}

// GenSecret generates a "random" string of digits.
func GenSecret() string {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprint(time.Now().UnixNano())))
	return fmt.Sprint(h.Sum32())

}
