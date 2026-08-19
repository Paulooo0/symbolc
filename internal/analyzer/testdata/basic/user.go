package main

type User struct {
	ID   string
	Name string
}

type CreateUserInput struct {
	Name string
}

type UserService struct{}

func (UserService) CreateUser(input CreateUserInput) (User, error) {
	user := User{}
	return user, nil
}
