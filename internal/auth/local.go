package auth

import "crypto/subtle"

type LocalAuth struct {
	username string
	password string
}

func NewLocalAuth(username, password string) *LocalAuth {
	return &LocalAuth{username: username, password: password}
}

func (la *LocalAuth) Verify(username, password string) bool {
	usernameMatch := subtle.ConstantTimeCompare([]byte(la.username), []byte(username)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(la.password), []byte(password)) == 1
	return usernameMatch && passwordMatch
}
