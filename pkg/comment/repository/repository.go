package repository

import "redditclone/pkg/models"

//go:generate mockgen -source=repository.go -destination=mock_repository/comment_mock.go -package=mock_repository MockCommentRepository
type CommentRepo interface {
	GetCommentByID(commentID string) (*models.Comment, error)
	CreateComment(post *models.Post, user *models.User, commentText string) (*models.Comment, error)
	DeleteComment(comment *models.Comment) error
}
