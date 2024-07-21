package delivery

import (
	"encoding/json"
	"io"
	"net/http"
	commentRepository "redditclone/pkg/comment/repository"
	"redditclone/pkg/middleware"
	postRepository "redditclone/pkg/post/repository"
	userRepository "redditclone/pkg/user/repository"
	"redditclone/tools"

	"github.com/asaskevich/govalidator"

	"github.com/gorilla/mux"
)

type PostHandler struct {
	CommentRepo commentRepository.CommentRepo
	PostRepo    postRepository.PostRepo
	UserRepo    userRepository.UserRepo
}

func (h *PostHandler) Index(w http.ResponseWriter, r *http.Request) {
	posts, err := h.PostRepo.GetAllPosts("", "")
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetAllPosts")
	}

	jsonPosts, err := json.Marshal(posts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Index")
		return
	}

	_, err = w.Write(jsonPosts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Index")
		return
	}
}

func (h *PostHandler) IndexByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	posts, err := h.PostRepo.GetAllPosts("", username)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetAllPosts")
	}

	jsonPosts, err := json.Marshal(posts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.IndexByUser")
		return
	}

	_, err = w.Write(jsonPosts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.IndexByUser")
		return
	}
}

func (h *PostHandler) IndexByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]
	posts, err := h.PostRepo.GetAllPosts(category, "")
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetAllPosts")
	}
	jsonPosts, err := json.Marshal(posts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.IndexByCategory")
		return
	}

	_, err = w.Write(jsonPosts)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.IndexByCategory")
		return
	}
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["id"]
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetPostByID")
		return
	}

	jsonPost, err := json.Marshal(post)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.GetPost")
		return
	}

	_, err = w.Write(jsonPost)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.GetPost")
		return
	}
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postID"]
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Delete")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "PostHandler.Create")
		return
	}

	if userID != post.Author.ID {
		tools.JSONError(w, http.StatusForbidden, "you are not allowed to delete this post", "PostHandler.Delete")
		return
	}

	for _, comment := range post.Comments {
		err := h.CommentRepo.DeleteComment(comment)
		if err != nil {
			tools.JSONError(w, http.StatusInternalServerError, "cant delete post comment", "CommentRepo.DeleteComment")
			return
		}
	}

	err = h.PostRepo.DeletePost(post)
	if err != nil {
		tools.JSONError(w, http.StatusForbidden, "cant delete such post", "PostRepo.DeletePost")
		return
	}

	jsonOK, err := json.Marshal(map[string]string{"message": "success"})
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Delete")
		return
	}

	_, err = w.Write(jsonOK)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Delete")
		return
	}
}

type PostForm struct {
	Category string `json:"category" valid:"in(music|funny|videos|programming|news|fashion)"`
	Title    string `json:"title"`
	Type     string `json:"type" valid:"in(text|link)"`
	URL      string `json:"url" valid:"url"`
	Text     string `json:"text"`
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "PostHandler.Create")
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "UserRepo.GetUserByID")
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Create")
		return
	}

	postForm := &PostForm{}
	err = json.Unmarshal(body, postForm)
	if err != nil {
		tools.JSONError(w, http.StatusBadRequest, "cant unpack payload", "PostHandler.Create")
		return
	}

	_, err = govalidator.ValidateStruct(postForm)
	if err != nil {
		tools.ValidationError(w, err)
		return
	}

	newPost, err := h.PostRepo.CreateNewPost(postForm.Category, postForm.Title, postForm.Type, postForm.URL, postForm.Text, user)
	if err != nil {
		tools.JSONError(w, http.StatusConflict, err.Error(), "PostRepo.CreateNewPost")
		return
	}

	newPostJSON, err := json.Marshal(newPost)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Create")
		return
	}

	_, err = w.Write(newPostJSON)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Create")
		return
	}
}

func (h *PostHandler) Vote(w http.ResponseWriter, r *http.Request, rate int) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "PostHandler.Create")
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "UserRepo.GetUserByID")
		return
	}

	vars := mux.Vars(r)
	postID := vars["postID"]
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetPostByID")
		return
	}

	err = h.PostRepo.UpvotePost(user, post, rate)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.UpvotePost")
		return
	}

	postJSON, err := json.Marshal(post)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Upvote")
		return
	}

	_, err = w.Write(postJSON)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostHandler.Upvote")
		return
	}
}

func (h *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	h.Vote(w, r, 1)
}

func (h *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	h.Vote(w, r, 0)
}

func (h *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	h.Vote(w, r, -1)
}
