package mongo

import (
	"context"
	"redditclone/pkg/models"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostMongoDBRepository struct {
	DB *mongo.Collection
}

var validPostCategories = map[string]bool{
	"music":       true,
	"funny":       true,
	"videos":      true,
	"programming": true,
	"news":        true,
	"fashion":     true,
}

func NewPostMongoDBMemoryRepo(postsCollection *mongo.Collection) *PostMongoDBRepository {
	return &PostMongoDBRepository{
		DB: postsCollection,
	}
}

func (repo *PostMongoDBRepository) GetAllPosts(category string, username string) ([]*models.Post, error) {
	posts := []*models.Post{}

	filter := bson.M{}
	if category != "" {
		if _, ok := validPostCategories[category]; !ok {
			return nil, models.ErrIncorrectPostCategory
		}
		filter["category"] = category
	} else if username != "" {
		filter["author.username"] = username
	}

	cursor, err := repo.DB.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (repo *PostMongoDBRepository) GetPostByID(id string) (*models.Post, error) {
	primitiveID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, models.ErrCorruptedPostID
	}
	filter := bson.M{"_id": primitiveID}
	post := models.Post{}

	err = repo.DB.FindOne(context.Background(), filter).Decode(&post)
	if err == mongo.ErrNoDocuments {
		return nil, models.ErrNoPost
	} else if err != nil {
		return nil, err
	}

	return &post, nil
}

func (repo *PostMongoDBRepository) CreateNewPost(category string, title string, postType string, url string, text string, user *models.User) (*models.Post, error) {
	newPostBSON := bson.M{
		"_id":              primitive.NewObjectID(),
		"score":            1,
		"views":            1,
		"type":             postType,
		"title":            title,
		"author":           user,
		"category":         category,
		"text":             text,
		"url":              url,
		"created":          time.Now(),
		"upvotePercentage": 100,
		"votes": []*models.Vote{
			{
				Author:   *user,
				AuthorID: user.ID,
				Vote:     1,
			},
		},
	}
	newPostDoc, err := bson.Marshal(newPostBSON)
	if err != nil {
		return nil, err
	}

	_, err = repo.DB.InsertOne(context.Background(), newPostBSON)
	if err != nil {
		return nil, err
	}

	var newPost *models.Post
	err = bson.Unmarshal(newPostDoc, &newPost)
	if err != nil {
		return nil, err
	}

	return newPost, nil
}

func (repo *PostMongoDBRepository) UpvotePost(user *models.User, post *models.Post, rate int) error {
	if rate > 1 || rate < -1 {
		return models.ErrUnrecognizedRate
	}

	newUpvote := &models.Vote{
		Author:   *user,
		AuthorID: user.ID,
		Vote:     rate,
	}

	voteIndex := slices.IndexFunc(post.Votes, func(vote *models.Vote) bool {
		return vote.Author.ID == user.ID
	})

	if voteIndex != -1 {
		post.Score -= post.Votes[voteIndex].Vote
		post.Votes = slices.Delete(post.Votes, voteIndex, voteIndex+1)
	}

	if rate == 1 || rate == -1 {
		post.Votes = append(post.Votes, newUpvote)
		post.Score += newUpvote.Vote
	}

	if post.Score < 0 || len(post.Votes) == 0 {
		post.UpvotePercentage = 0
	} else {
		post.UpvotePercentage = post.Score / len(post.Votes) * 100
	}

	filter := bson.M{"_id": post.ID}
	res, err := repo.DB.UpdateOne(
		context.Background(),
		filter,
		bson.M{"$set": bson.M{
			"votes":            post.Votes,
			"score":            post.Score,
			"upvotePercentage": post.UpvotePercentage,
		}},
	)
	if err != nil {
		return models.ErrUpdatePost
	} else if res.ModifiedCount == 0 {
		return models.ErrNoPost
	}

	return nil
}

func (repo *PostMongoDBRepository) DeletePostComment(post *models.Post, deleteComment *models.Comment) error {
	commentIndex := slices.IndexFunc(post.Comments, func(comment *models.Comment) bool {
		return comment.ID == deleteComment.ID
	})

	if commentIndex != -1 {
		post.Comments = slices.Delete(post.Comments, commentIndex, commentIndex+1)
	}

	filter := bson.M{"_id": post.ID}
	res, err := repo.DB.UpdateOne(
		context.Background(),
		filter,
		bson.M{"$set": bson.M{
			"comments": post.Comments,
		}},
	)

	if err != nil {
		return models.ErrUpdatePost
	} else if res.ModifiedCount == 0 {
		return models.ErrNoPost
	}

	return nil
}

func (repo *PostMongoDBRepository) AddPostComment(post *models.Post, comment *models.Comment) (*models.Post, error) {
	post.Comments = append(post.Comments, comment)

	filter := bson.M{"_id": post.ID}
	res, err := repo.DB.UpdateOne(
		context.Background(),
		filter,
		bson.M{"$set": bson.M{
			"comments": post.Comments,
		}},
	)

	if err != nil {
		return nil, models.ErrUpdatePost
	} else if res.ModifiedCount == 0 {
		return nil, models.ErrNoPost
	}

	return post, nil
}

func (repo *PostMongoDBRepository) DeletePost(post *models.Post) error {
	filter := bson.M{"_id": post.ID}

	res, err := repo.DB.DeleteOne(context.Background(), filter)
	if err != nil {
		return models.ErrDeletePost
	} else if res.DeletedCount == 0 {
		return models.ErrNoPost
	}

	return nil
}
