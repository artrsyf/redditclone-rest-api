package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	commentMock "redditclone/pkg/comment/repository/mock_repository"
	"redditclone/pkg/middleware"
	"redditclone/pkg/models"
	postMock "redditclone/pkg/post/repository/mock_repository"
	userMock "redditclone/pkg/user/repository/mock_repository"
	"redditclone/tools"
)

func ComparePosts(firstPost, secondPost models.Post) bool {
	return firstPost.ID == secondPost.ID &&
		firstPost.Title == secondPost.Title &&
		firstPost.Text == secondPost.Text &&
		firstPost.URL == secondPost.URL &&
		firstPost.Category == secondPost.Category &&
		firstPost.Score == secondPost.Score &&
		firstPost.Views == secondPost.Views &&
		firstPost.Type == secondPost.Type &&
		firstPost.Author.ID == secondPost.Author.ID &&
		firstPost.Author.Login == secondPost.Author.Login &&
		firstPost.Created == secondPost.Created &&
		firstPost.UpvotePercentage == secondPost.UpvotePercentage
}

func TestIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
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

	t.Run("correct Index", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts("", "").Return([]*models.Post{&post}, nil)

		req := httptest.NewRequest("GET", "/api/posts/", nil)
		w := httptest.NewRecorder()

		postHandler.Index(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		posts := []*models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&posts)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, *posts[0]))
	})

	t.Run("PostRepo.GetAllPosts error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts("", "").Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "/api/posts/", nil)
		w := httptest.NewRecorder()

		postHandler.Index(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestIndexByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
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

	t.Run("correct IndexByUser", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts("", postAuthor.Login).Return([]*models.Post{&post}, nil)

		req := httptest.NewRequest("GET", "/api/user/"+postAuthor.Login, nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"username": postAuthor.Login,
		}
		req = mux.SetURLVars(req, vars)

		postHandler.IndexByUser(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		posts := []*models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&posts)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, *posts[0]))
	})

	t.Run("PostRepo.GetAllPosts error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts("", postAuthor.Login).Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "/api/user/"+postAuthor.Login, nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"username": postAuthor.Login,
		}
		req = mux.SetURLVars(req, vars)

		postHandler.IndexByUser(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestIndexByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}
	var category = "news"

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            1,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         category,
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

	t.Run("correct IndexByCategory", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts(category, "").Return([]*models.Post{&post}, nil)

		req := httptest.NewRequest("GET", "/api/posts/"+category, nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"category": category,
		}
		req = mux.SetURLVars(req, vars)

		postHandler.IndexByCategory(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		posts := []*models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&posts)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, *posts[0]))
	})

	t.Run("PostRepo.GetAllPosts error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetAllPosts(category, "").Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "/api/posts/"+category, nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"category": category,
		}
		req = mux.SetURLVars(req, vars)

		postHandler.IndexByCategory(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}
	var category = "news"

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            1,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         category,
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

	t.Run("correct GetPost", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)

		req := httptest.NewRequest("GET", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"id": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		postHandler.GetPost(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})

	t.Run("PostRepo.GetPostByID error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"id": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		postHandler.GetPost(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}
	var category = "news"

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
		ID:               primitive.NewObjectID(),
		Title:            "some title",
		Score:            1,
		Views:            1,
		Type:             "text",
		Author:           postAuthor,
		Category:         category,
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
			{
				ID:      primitive.NewObjectID(),
				Author:  &postAuthor,
				Created: createdTime,
				Text:    "comment body",
			},
		},
	}

	t.Run("correct Delete", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockCommentRepo.EXPECT().DeleteComment(post.Comments[0]).Return(nil)
		mockPostRepo.EXPECT().DeletePost(&post).Return(nil)

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response["message"], "success")
	})

	t.Run("PostRepo.GetPostByID error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("auth error, permission denied", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "you should authorize first", response["error"])
	})

	t.Run("permission denied - delete not owns post error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		otherUserID := postAuthor.ID + 2
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, otherUserID)
		req = req.WithContext(ctx)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "you are not allowed to delete this post", response["error"])
	})

	t.Run("CommentRepo.DeleteComment error", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockCommentRepo.EXPECT().DeleteComment(post.Comments[0]).Return(errors.New("mock error"))

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "cant delete post comment", response["error"])
	})

	t.Run("correct Delete", func(t *testing.T) {
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockCommentRepo.EXPECT().DeleteComment(post.Comments[0]).Return(nil)
		mockPostRepo.EXPECT().DeletePost(&post).Return(errors.New("mock error"))

		req := httptest.NewRequest("DELETE", "/api/post/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Delete(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "cant delete such post", response["error"])
	})
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
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
		Comments: []*models.Comment{
			{
				ID:      primitive.NewObjectID(),
				Author:  &postAuthor,
				Created: createdTime,
				Text:    "comment body",
			},
		},
	}

	var postForm = &PostForm{
		Category: post.Category,
		Title:    post.Title,
		Type:     post.Type,
		Text:     post.Text,
	}

	t.Run("correct Create", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().CreateNewPost(
			postForm.Category,
			postForm.Title,
			postForm.Type,
			postForm.URL,
			postForm.Text,
			&postAuthor,
		).Return(&post, nil)

		reqBody, err := json.Marshal(postForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Create(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err = json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})

	t.Run("auth error, permission denied", func(t *testing.T) {
		reqBody, err := json.Marshal(postForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		postHandler.Create(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "you should authorize first", response["error"])
	})

	t.Run("UserRepo.GetUserByID", func(t *testing.T) {
		unknownUserID := postAuthor.ID + 2

		mockUserRepo.EXPECT().GetUserByID(unknownUserID).Return(nil, errors.New("mock error"))
		reqBody, err := json.Marshal(postForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, unknownUserID)
		req = req.WithContext(ctx)

		postHandler.Create(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "you should authorize first", response["error"])
	})

	t.Run("govalidator.ValidateStruct error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)

		// Делаем форму некорректной для валидации.
		incorrectPostForm := *postForm
		incorrectPostForm.Category = "incorrect post category"

		reqBody, err := json.Marshal(&incorrectPostForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Create(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("PostRepo.CreateNewPost error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().CreateNewPost(
			postForm.Category,
			postForm.Title,
			postForm.Type,
			postForm.URL,
			postForm.Text,
			&postAuthor,
		).Return(nil, errors.New("mock error"))

		reqBody, err := json.Marshal(postForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/posts", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Create(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
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
		Comments: []*models.Comment{
			{
				ID:      primitive.NewObjectID(),
				Author:  &postAuthor,
				Created: createdTime,
				Text:    "comment body",
			},
		},
	}

	var voteRate = 1

	t.Run("correct Vote", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockPostRepo.EXPECT().UpvotePost(&postAuthor, &post, voteRate).Return(nil)

		// Данный URL не реализован, т.к. хендлеру Vote делегируется изменение поста.
		req := httptest.NewRequest("GET", "/post/upvote/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Vote(w, req, voteRate)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})

	t.Run("auth error, permission denied", func(t *testing.T) {
		// Данный URL не реализован, т.к. хендлеру Vote делегируется изменение поста.
		req := httptest.NewRequest("GET", "/post/upvote/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		postHandler.Vote(w, req, voteRate)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "you should authorize first", response["error"])
	})

	t.Run("UserRepo.GetUserByID", func(t *testing.T) {
		unknownUserID := postAuthor.ID + 2

		mockUserRepo.EXPECT().GetUserByID(unknownUserID).Return(nil, errors.New("mock error"))

		req := httptest.NewRequest("GET", "/post/upvote/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, unknownUserID)
		req = req.WithContext(ctx)

		postHandler.Vote(w, req, voteRate)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "you should authorize first", response["error"])
	})

	t.Run("PostRepo.GetPostByID error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(nil, errors.New("mock error"))

		// Данный URL не реализован, т.к. хендлеру Vote делегируется изменение поста.
		req := httptest.NewRequest("GET", "/post/upvote/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Vote(w, req, voteRate)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("PostRepo.UpvotePost error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockPostRepo.EXPECT().UpvotePost(&postAuthor, &post, voteRate).Return(errors.New("mock error"))

		// Данный URL не реализован, т.к. хендлеру Vote делегируется изменение поста.
		req := httptest.NewRequest("GET", "/post/upvote/"+post.ID.Hex(), nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Vote(w, req, voteRate)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestUpUnDownvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockPostRepo := postMock.NewMockPostRepo(ctrl)
	mockCommentRepo := commentMock.NewMockCommentRepo(ctrl)

	postHandler := &PostHandler{
		UserRepo:    mockUserRepo,
		PostRepo:    mockPostRepo,
		CommentRepo: mockCommentRepo,
	}

	tools.Init()

	var postAuthor = models.User{
		ID:    1,
		Login: "alex12345",
	}

	var createdTime = time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC)

	var post = models.Post{
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
		Comments: []*models.Comment{
			{
				ID:      primitive.NewObjectID(),
				Author:  &postAuthor,
				Created: createdTime,
				Text:    "comment body",
			},
		},
	}

	t.Run("correct Upvote", func(t *testing.T) {
		voteRate := 1

		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockPostRepo.EXPECT().UpvotePost(&postAuthor, &post, voteRate).Return(nil)

		req := httptest.NewRequest("GET", "/api/post/"+post.ID.Hex()+"/upvote", nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Upvote(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})

	t.Run("correct Unvote", func(t *testing.T) {
		voteRate := 0

		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockPostRepo.EXPECT().UpvotePost(&postAuthor, &post, voteRate).Return(nil)

		req := httptest.NewRequest("GET", "/api/post/"+post.ID.Hex()+"/unvote", nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Unvote(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})

	t.Run("correct Downvote", func(t *testing.T) {
		voteRate := -1

		mockUserRepo.EXPECT().GetUserByID(postAuthor.ID).Return(&postAuthor, nil)
		mockPostRepo.EXPECT().GetPostByID(post.ID.Hex()).Return(&post, nil)
		mockPostRepo.EXPECT().UpvotePost(&postAuthor, &post, voteRate).Return(nil)

		req := httptest.NewRequest("GET", "/api/post/"+post.ID.Hex()+"/downvote", nil)
		w := httptest.NewRecorder()

		vars := map[string]string{
			"postID": post.ID.Hex(),
		}
		req = mux.SetURLVars(req, vars)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, postAuthor.ID)
		req = req.WithContext(ctx)

		postHandler.Downvote(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		actualPost := models.Post{}
		err := json.NewDecoder(resp.Body).Decode(&actualPost)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, ComparePosts(post, actualPost))
	})
}
