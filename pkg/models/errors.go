package models

import "errors"

var (
	ErrNoSession = errors.New("cant find such session")

	ErrNoUser           = errors.New("no user found")
	ErrWrongCredentials = errors.New("wrong login or password")
	ErrAlreadyCreated   = errors.New("already created")

	ErrCorruptedCommentID = errors.New("bad comment id")
	ErrNoComment          = errors.New("cant find such comment")
	ErrDeleteComment      = errors.New("cant delete comment")

	ErrCorruptedPostID       = errors.New("bad post id")
	ErrUnrecognizedRate      = errors.New("unrecognized rate")
	ErrNoPost                = errors.New("cant find such post")
	ErrUpdatePost            = errors.New("cant update post")
	ErrDeletePost            = errors.New("cant delete post")
	ErrIncorrectPostCategory = errors.New("incorrect post category")
)
