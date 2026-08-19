package main

type UserRepository interface {
	CreateUser(User) (User, error)
}

var ErrUserNotFound = error(nil)

const MaxRetries = 3
