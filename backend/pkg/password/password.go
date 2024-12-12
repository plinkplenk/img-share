package password

import "golang.org/x/crypto/bcrypt"

func Compare(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
