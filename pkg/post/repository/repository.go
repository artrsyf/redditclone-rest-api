package repository

import "redditclone/pkg/models"

//go:generate mockgen -source=repository.go -destination=mock_repository/post_mock.go -package=mock_repository MockPostRepository
type PostRepo interface {
	GetAllPosts(category string, username string) ([]*models.Post, error)
	CreateNewPost(category string, title string, postType string, url string, text string, user *models.User) (*models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	UpvotePost(user *models.User, post *models.Post, rate int) error
	DeletePostComment(post *models.Post, comment *models.Comment) error
	AddPostComment(post *models.Post, comment *models.Comment) (*models.Post, error)
	DeletePost(post *models.Post) error
}
