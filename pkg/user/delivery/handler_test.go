package delivery

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/models"
	sessionMock "redditclone/pkg/session/repository/mock_repository"
	userMock "redditclone/pkg/user/repository/mock_repository"
	"redditclone/tools"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("reader error")
}

type errWriter struct {
	http.ResponseWriter
}

func (w errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("writer error")
}

func TestUserHandlerSignup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockSessionRepo := sessionMock.NewMockSessionManager(ctrl)

	userHandler := &UserHandler{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
	}

	tools.Init()

	var authForm = &AuthForm{
		Login:    "alex12345",
		Password: "alex12345",
	}

	var user = &models.User{
		ID:       1,
		Login:    authForm.Login,
		Password: authForm.Password,
	}

	t.Run("correct signup", func(t *testing.T) {
		mockUserRepo.EXPECT().CreateUser(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Create(gomock.Any(), user.ID).Return(nil)

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Signup(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}

		err = json.NewDecoder(resp.Body).Decode(&response)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, response["token"])
	})

	t.Run("error reading body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/register", errReader(0))
		w := httptest.NewRecorder()

		userHandler.Signup(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("error writing response", func(t *testing.T) {
		mockUserRepo.EXPECT().CreateUser(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Create(gomock.Any(), user.ID).Return(nil)

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(reqBody))
		errorWriter := errWriter{}

		userHandler.Signup(errorWriter, req)
	})

	t.Run("bad request", func(t *testing.T) {
		badReqBody := []byte("{")
		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(badReqBody))
		w := httptest.NewRecorder()

		userHandler.Signup(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("UserRepo.CreateUser error", func(t *testing.T) {
		mockUserRepo.EXPECT().CreateUser(authForm.Login, authForm.Password).Return(nil, errors.New("mock error"))

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Signup(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		errorWrapper := "couldnt create user:"
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, response["error"], errorWrapper)
	})

	t.Run("SessionRepo.Create error", func(t *testing.T) {
		mockUserRepo.EXPECT().CreateUser(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Create(gomock.Any(), user.ID).Return(errors.New("mock error"))

		reqBody, _ := json.Marshal(authForm)
		req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Signup(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestUserHandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockUserRepo(ctrl)
	mockSessionRepo := sessionMock.NewMockSessionManager(ctrl)

	userHandler := &UserHandler{
		UserRepo:    mockUserRepo,
		SessionRepo: mockSessionRepo,
	}

	tools.Init()

	var authForm = &AuthForm{
		Login:    "alex12345",
		Password: "alex12345",
	}

	var user = &models.User{
		ID:       1,
		Login:    authForm.Login,
		Password: authForm.Password,
	}

	var session = &models.Session{
		ID:        1,
		JWT:       "some jwt token",
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 4),
	}

	t.Run("correct login", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Check(user.ID).Return(session, nil)

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}

		err = json.NewDecoder(resp.Body).Decode(&response)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, response["token"])
	})

	t.Run("error reading body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/login", errReader(0))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("error writing response", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Check(user.ID).Return(session, nil)

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		errorWriter := errWriter{}

		userHandler.Login(errorWriter, req)
	})

	t.Run("bad request", func(t *testing.T) {
		badReqBody := []byte("{")
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(badReqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UserRepo.GetUserFromRepo error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(nil, errors.New("mock error"))

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("SessionRepo.Check unexpected error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Check(user.ID).Return(nil, errors.New("mock error"))

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("SessionRepo.Check ErrNoSession - creating user session", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Check(user.ID).Return(nil, models.ErrNoSession)
		mockSessionRepo.EXPECT().Create(gomock.Any(), user.ID).Return(nil)

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotNil(t, response["token"])
	})

	t.Run("SessionRepo.Check ErrNoSession -> SessionRepo.Create error", func(t *testing.T) {
		mockUserRepo.EXPECT().GetUserFromRepo(authForm.Login, authForm.Password).Return(user, nil)
		mockSessionRepo.EXPECT().Check(user.ID).Return(nil, models.ErrNoSession)
		mockSessionRepo.EXPECT().Create(gomock.Any(), user.ID).Return(errors.New("mock error"))

		reqBody, err := json.Marshal(authForm)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		userHandler.Login(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
