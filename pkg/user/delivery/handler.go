package delivery

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"redditclone/pkg/models"
	sessionRepository "redditclone/pkg/session/repository"
	userRepository "redditclone/pkg/user/repository"
	"strconv"
	"time"

	"redditclone/tools"

	"github.com/dgrijalva/jwt-go"
)

type AuthForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type UserHandler struct {
	UserRepo    userRepository.UserRepo
	SessionRepo sessionRepository.SessionManager
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Signup")
		return
	}

	authForm := &AuthForm{}
	err = json.Unmarshal(body, authForm)
	if err != nil {
		tools.JSONError(w, http.StatusUnauthorized, "bad login or pass", "UserHandler.Signup")
		return
	}

	user, err := h.UserRepo.CreateUser(authForm.Login, authForm.Password)
	if err != nil {
		tools.JSONError(w, http.StatusUnauthorized, "couldnt create user:"+err.Error(), "UserRepo.CreateUser")
		return
	}

	tokenString, err := createUserJWT(user)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "createUserJWT")
		return
	}

	err = h.SessionRepo.Create(tokenString, user.ID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "SessionRepo.Create")
		return
	}

	response, err := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Signup")
		return
	}

	_, err = w.Write(response)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Signup")
		return
	}
}

func createUserJWT(user *models.User) (string, error) {
	tokenKey := []byte(os.Getenv("TOKEN_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]string{
			"username": user.Login,
			"id":       strconv.Itoa(user.ID),
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().AddDate(0, 0, 4).Unix(),
	})
	tokenString, err := token.SignedString(tokenKey)

	return tokenString, err
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Login")
		return
	}

	authForm := &AuthForm{}
	err = json.Unmarshal(body, authForm)
	if err != nil {
		tools.JSONError(w, http.StatusBadRequest, "cant unpack payload", "UserHandler.Login")
		return
	}

	user, err := h.UserRepo.GetUserFromRepo(authForm.Login, authForm.Password)
	if err != nil {
		tools.JSONError(w, http.StatusBadRequest, err.Error(), "UserRepo.GetUserFromRepo")
		return
	}

	tokenString, err := createUserJWT(user)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "createUserJWT")
		return
	}

	_, err = h.SessionRepo.Check(user.ID)
	if err == models.ErrNoSession {
		if createSessionErr := h.SessionRepo.Create(tokenString, user.ID); createSessionErr != nil {
			tools.JSONError(w, http.StatusInternalServerError, err.Error(), "SessionRepo.Create")
			return
		}
	} else if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "SessionRepo.Check")
		return
	}

	response, err := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Login")
		return
	}

	_, err = w.Write(response)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "UserHandler.Login")
		return
	}
}
