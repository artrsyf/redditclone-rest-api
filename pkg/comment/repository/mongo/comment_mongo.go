package mongo

import (
	"context"
	"redditclone/pkg/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentMongoDBRepository struct {
	DB *mongo.Collection
}

func NewCommentMongoDBRepository(commentCollection *mongo.Collection) *CommentMongoDBRepository {
	return &CommentMongoDBRepository{
		DB: commentCollection,
	}
}

func (repo *CommentMongoDBRepository) GetCommentByID(commentID string) (*models.Comment, error) {
	primitiveID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return nil, models.ErrCorruptedCommentID
	}
	filter := bson.M{"_id": primitiveID}
	comment := models.Comment{}

	err = repo.DB.FindOne(context.Background(), filter).Decode(&comment)
	if err != nil {
		return nil, models.ErrNoComment
	}

	return &comment, nil
}

func (repo *CommentMongoDBRepository) CreateComment(post *models.Post, user *models.User, commentText string) (*models.Comment, error) {
	newCommentBSON := bson.M{
		"_id":     primitive.NewObjectID(),
		"text":    commentText,
		"author":  user,
		"created": time.Now(),
	}
	newCommentDoc, err := bson.Marshal(newCommentBSON)
	if err != nil {
		return nil, err
	}

	_, err = repo.DB.InsertOne(context.Background(), newCommentBSON)
	if err != nil {
		return nil, err
	}

	var newComment *models.Comment
	err = bson.Unmarshal(newCommentDoc, &newComment)
	if err != nil {
		return nil, err
	}

	return newComment, nil
}

func (repo *CommentMongoDBRepository) DeleteComment(comment *models.Comment) error {
	filter := bson.M{"_id": comment.ID}

	res, err := repo.DB.DeleteOne(context.Background(), filter)
	if err != nil {
		return models.ErrDeleteComment
	} else if res.DeletedCount == 0 {
		return models.ErrNoComment
	}

	return nil
}
