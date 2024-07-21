package mongo

import (
	"redditclone/pkg/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewPostMongoDBMemoryRepo(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("correct query", func(mt *mtest.T) {
		postsCollection := mt.Coll
		repo := NewPostMongoDBMemoryRepo(postsCollection)

		assert.NotNil(t, repo)
		assert.Equal(t, postsCollection, repo.DB)
	})
}

func TestGetPostByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	// defer mt.Close()

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		postAuthor := models.User{
			ID:    1,
			Login: "alex12345",
		}

		createdTime := time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

		expectedPost := &models.Post{
			ID:               primitive.NewObjectID(),
			Title:            "some title",
			Score:            1,
			Views:            1,
			Type:             "text",
			Author:           postAuthor,
			Category:         "news",
			Text:             "post content",
			Created:          createdTime,
			UpvotePercentage: 100,
			Votes: []*models.Vote{
				{
					Author:   postAuthor,
					AuthorID: postAuthor.ID,
					Vote:     1,
				},
			},
		}

		response := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			bson.E{Key: "_id", Value: expectedPost.ID},
			bson.E{Key: "title", Value: expectedPost.Title},
			bson.E{Key: "score", Value: expectedPost.Score},
			bson.E{Key: "views", Value: expectedPost.Views},
			bson.E{Key: "type", Value: expectedPost.Type},
			bson.E{Key: "author", Value: expectedPost.Author},
			bson.E{Key: "category", Value: expectedPost.Category},
			bson.E{Key: "text", Value: expectedPost.Text},
			bson.E{Key: "url", Value: expectedPost.URL},
			bson.E{Key: "created", Value: expectedPost.Created},
			bson.E{Key: "upvotePercentage", Value: expectedPost.UpvotePercentage},
			bson.E{Key: "votes", Value: expectedPost.Votes},
		})

		mt.AddMockResponses(response)

		post, err := repo.GetPostByID(expectedPost.ID.Hex())
		assert.Nil(t, err)
		assert.Equal(t, expectedPost, post)
	})

	mt.Run("ErrCorruptedPostID", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		incorrectPostID := "qwe"

		_, err := repo.GetPostByID(incorrectPostID)
		assert.Equal(t, models.ErrCorruptedPostID, err)
	})

	mt.Run("ErrNoPost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		response := mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch)

		mt.AddMockResponses(response)

		_, err := repo.GetPostByID(primitive.NewObjectID().Hex())

		assert.Equal(t, models.ErrNoPost, err)
	})

	mt.Run("unexpected error", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{bson.E{Key: "ok", Value: 0}})

		_, err := repo.GetPostByID(primitive.NewObjectID().Hex())

		assert.NotNil(t, err)
	})
}

func TestGetAllPosts(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var firstPost = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            1,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	var secondPost = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some second title",
		Score:            1,
		Views:            1,
		Type:             "url",
		Author:           postAuthor,
		Category:         "funny",
		URL:              "www.google.con",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		expectedPosts := []*models.Post{
			&firstPost,
			&secondPost,
		}

		for idx, expectedPost := range expectedPosts {
			var batchID mtest.BatchIdentifier
			if idx == 0 {
				batchID = mtest.FirstBatch
			} else {
				batchID = mtest.NextBatch
			}
			response := mtest.CreateCursorResponse(1, "foo.bar", batchID, bson.D{
				bson.E{Key: "_id", Value: expectedPost.ID},
				bson.E{Key: "title", Value: expectedPost.Title},
				bson.E{Key: "score", Value: expectedPost.Score},
				bson.E{Key: "views", Value: expectedPost.Views},
				bson.E{Key: "type", Value: expectedPost.Type},
				bson.E{Key: "author", Value: expectedPost.Author},
				bson.E{Key: "category", Value: expectedPost.Category},
				bson.E{Key: "text", Value: expectedPost.Text},
				bson.E{Key: "url", Value: expectedPost.URL},
				bson.E{Key: "created", Value: expectedPost.Created},
				bson.E{Key: "upvotePercentage", Value: expectedPost.UpvotePercentage},
				bson.E{Key: "votes", Value: expectedPost.Votes},
			})

			mt.AddMockResponses(response)
		}
		killCursor := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(killCursor)

		posts, err := repo.GetAllPosts("", "")
		assert.Nil(t, err)
		assert.Equal(t, expectedPosts, posts)
	})

	mt.Run("correct query by category", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		byCategoryResponse := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			bson.E{Key: "_id", Value: firstPost.ID},
			bson.E{Key: "title", Value: firstPost.Title},
			bson.E{Key: "score", Value: firstPost.Score},
			bson.E{Key: "views", Value: firstPost.Views},
			bson.E{Key: "type", Value: firstPost.Type},
			bson.E{Key: "author", Value: firstPost.Author},
			bson.E{Key: "category", Value: firstPost.Category},
			bson.E{Key: "text", Value: firstPost.Text},
			bson.E{Key: "url", Value: firstPost.URL},
			bson.E{Key: "created", Value: firstPost.Created},
			bson.E{Key: "upvotePercentage", Value: firstPost.UpvotePercentage},
			bson.E{Key: "votes", Value: firstPost.Votes},
		})
		killCursor := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		mt.AddMockResponses(byCategoryResponse, killCursor)

		postsByCategory, err := repo.GetAllPosts(firstPost.Category, "")
		assert.Nil(t, err)
		assert.Equal(t, []*models.Post{&firstPost}, postsByCategory)
	})

	mt.Run("correct query by username", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		byUsernameResponse := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			bson.E{Key: "_id", Value: secondPost.ID},
			bson.E{Key: "title", Value: secondPost.Title},
			bson.E{Key: "score", Value: secondPost.Score},
			bson.E{Key: "views", Value: secondPost.Views},
			bson.E{Key: "type", Value: secondPost.Type},
			bson.E{Key: "author", Value: secondPost.Author},
			bson.E{Key: "category", Value: secondPost.Category},
			bson.E{Key: "text", Value: secondPost.Text},
			bson.E{Key: "url", Value: secondPost.URL},
			bson.E{Key: "created", Value: secondPost.Created},
			bson.E{Key: "upvotePercentage", Value: secondPost.UpvotePercentage},
			bson.E{Key: "votes", Value: secondPost.Votes},
		})
		killCursor := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		mt.AddMockResponses(byUsernameResponse, killCursor)

		postsByUser, err := repo.GetAllPosts("", postAuthor.Login)
		assert.Nil(t, err)
		assert.Equal(t, []*models.Post{&secondPost}, postsByUser)
	})

	mt.Run("ErrIncorrectPostCategory", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		incorrectCategoryName := "incorrect"
		_, err := repo.GetAllPosts(incorrectCategoryName, "")
		assert.Equal(t, models.ErrIncorrectPostCategory, err)
	})

	mt.Run("incorrect cursor reading", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		errorResponse := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			bson.E{Key: "_id", Value: primitive.NewObjectID().Hex()},
			bson.E{Key: "unexpectedField", Value: "unexpectedFieldValue"},
		})
		mt.AddMockResponses(errorResponse)

		_, err := repo.GetAllPosts("", "")
		assert.NotNil(t, err)
	})

	mt.Run("error due finding posts", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{bson.E{Key: "ok", Value: 0}})

		_, err := repo.GetAllPosts("", "")
		assert.NotNil(t, err)
	})
}

func TestCreateNewPost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var createdPost = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            1,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		insertedPost, err := repo.CreateNewPost(createdPost.Category,
			createdPost.Title,
			createdPost.Type,
			"",
			createdPost.Text,
			&createdPost.Author,
		)

		assert.Nil(t, err)

		createdPost.ID = insertedPost.ID
		createdPost.Created = insertedPost.Created
		assert.Equal(t, &createdPost, insertedPost)
	})

	mt.Run("error due inserting post", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{bson.E{Key: "ok", Value: 0}})

		_, err := repo.CreateNewPost(createdPost.Category,
			createdPost.Title,
			createdPost.Type,
			"",
			createdPost.Text,
			&createdPost.Author,
		)

		assert.NotNil(t, err)
	})
}

func TestUpvotePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            0,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		})

		err := repo.UpvotePost(&postAuthor, &post, 1)
		assert.Nil(t, err)
	})

	mt.Run("correct query with negative post score", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		})

		err := repo.UpvotePost(&postAuthor, &post, -1)
		assert.Nil(t, err)
	})

	mt.Run("ErrUnrecognizedRate", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		incorrectRate := -2

		err := repo.UpvotePost(&postAuthor, &post, incorrectRate)
		assert.Equal(t, models.ErrUnrecognizedRate, err)
	})

	mt.Run("ErrUpdatePost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 0},
		})

		err := repo.UpvotePost(&postAuthor, &post, 1)
		assert.Equal(t, models.ErrUpdatePost, err)
	})

	mt.Run("ErrNoPost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
		})

		err := repo.UpvotePost(&postAuthor, &post, 1)
		assert.Equal(t, models.ErrNoPost, err)
	})
}

func TestDeletePostComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var comment = models.Comment{
		Created: time.Now(),
		Author:  &postAuthor,
		Text:    "comment text",
		ID:      primitive.NewObjectID(),
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            0,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
		Comments: []*models.Comment{
			&comment,
		},
	}

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		})

		err := repo.DeletePostComment(&post, &comment)
		assert.Nil(t, err)
	})

	mt.Run("ErrUpdatePost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 0},
		})

		err := repo.DeletePostComment(&post, &comment)
		assert.Equal(t, models.ErrUpdatePost, err)
	})

	mt.Run("ErrNoPost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
		})

		err := repo.DeletePostComment(&post, &comment)
		assert.Equal(t, models.ErrNoPost, err)
	})
}

func TestAddPostComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var comment = models.Comment{
		Created: time.Now(),
		Author:  &postAuthor,
		Text:    "comment text",
		ID:      primitive.NewObjectID(),
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            0,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	var updatedPostCommentCount = 1

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
			bson.E{Key: "value", Value: bson.D{
				bson.E{Key: "_id", Value: post.ID},
				bson.E{Key: "title", Value: post.Title},
				bson.E{Key: "score", Value: post.Score},
				bson.E{Key: "views", Value: post.Views},
				bson.E{Key: "type", Value: post.Type},
				bson.E{Key: "author", Value: post.Author},
				bson.E{Key: "category", Value: post.Category},
				bson.E{Key: "text", Value: post.Text},
				bson.E{Key: "url", Value: post.URL},
				bson.E{Key: "created", Value: post.Created},
				bson.E{Key: "upvotePercentage", Value: post.UpvotePercentage},
				bson.E{Key: "votes", Value: post.Votes},
				bson.E{Key: "comments", Value: []*models.Comment{
					&comment,
				}},
			}},
		})

		updatedPost, err := repo.AddPostComment(&post, &comment)

		assert.Nil(t, err)
		assert.Equal(t, updatedPostCommentCount, len(updatedPost.Comments))
	})

	mt.Run("ErrUpdatePost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 0},
		})

		_, err := repo.AddPostComment(&post, &comment)
		assert.Equal(t, models.ErrUpdatePost, err)
	})

	mt.Run("ErrNoPost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
		})

		_, err := repo.AddPostComment(&post, &comment)
		assert.Equal(t, models.ErrNoPost, err)
	})
}

func TestDeletePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            0,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         "news",
		Text:             "post content",
		Created:          createdTime,
		UpvotePercentage: 100,
		Votes: []*models.Vote{
			{
				Author:   postAuthor,
				AuthorID: postAuthor.ID,
				Vote:     1,
			},
		},
	}

	mt.Run("correct query", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "acknowledged", Value: true},
		})

		err := repo.DeletePost(&post)

		assert.Nil(t, err)
	})

	mt.Run("ErrDeletePost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 0},
		})

		err := repo.DeletePost(&post)

		assert.Equal(t, models.ErrDeletePost, err)
	})

	mt.Run("ErrNoPost", func(mt *mtest.T) {
		repo := PostMongoDBRepository{
			DB: mt.Coll,
		}

		mt.AddMockResponses(bson.D{
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "acknowledged", Value: true},
		})

		err := repo.DeletePost(&post)

		assert.Equal(t, models.ErrNoPost, err)
	})
}
