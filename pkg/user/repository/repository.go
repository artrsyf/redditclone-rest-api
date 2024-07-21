package repository

import "redditclone/pkg/models"

//go:generate mockgen -source=repository.go -destination=mock_repository/user_mock.go -package=mock_repository MockUserRepository
type UserRepo interface {
	GetUserFromRepo(login, pass string) (*models.User, error)
	CreateUser(login, pass string) (*models.User, error)
	GetUserByID(userID int) (*models.User, error)
}
