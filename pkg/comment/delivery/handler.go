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

	"github.com/gorilla/mux"
)

type CommentHandler struct {
	PostRepo    postRepository.PostRepo
	CommentRepo commentRepository.CommentRepo
	UserRepo    userRepository.UserRepo
}

type CommentForm struct {
	Text string `json:"comment"`
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	defer r.Body.Close()
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentHandler.Create")
		return
	}

	commentForm := &CommentForm{}
	err = json.Unmarshal(body, commentForm)
	if err != nil {
		tools.JSONError(w, http.StatusUnauthorized, "couldnt umarshall comment", "CommentHandler.Create")
		return
	}

	vars := mux.Vars(r)
	postID := vars["postID"]
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetPostByID")
		return
	}

	comment, err := h.CommentRepo.CreateComment(post, user, commentForm.Text)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentRepo.CreateComment")
		return
	}

	post, err = h.PostRepo.AddPostComment(post, comment)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.AddPostComment")
		return
	}

	jsonPost, err := json.Marshal(post)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentHandler.Create")
		return
	}

	_, err = w.Write(jsonPost)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentHandler.Create")
		return
	}
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		tools.JSONError(w, http.StatusUnauthorized, "you should authorize first", "PostHandler.Create")
		return
	}

	vars := mux.Vars(r)
	commentID := vars["commentID"]
	comment, err := h.CommentRepo.GetCommentByID(commentID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentRepo.GetCommentByID")
		return
	}

	if userID != comment.Author.ID {
		tools.JSONError(w, http.StatusForbidden, "you are not allowed to delete this comment", "CommentHandler.Delete")
		return
	}

	postID := vars["postID"]
	post, err := h.PostRepo.GetPostByID(postID)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.GetPostByID")
		return
	}

	err = h.CommentRepo.DeleteComment(comment)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentRepo.DeleteComment")
		return
	}

	err = h.PostRepo.DeletePostComment(post, comment)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "PostRepo.DeletePostComment")
		return
	}

	jsonPost, err := json.Marshal(post)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentHandler.Create")
		return
	}

	_, err = w.Write(jsonPost)
	if err != nil {
		tools.JSONError(w, http.StatusInternalServerError, err.Error(), "CommentHandler.Create")
		return
	}
}
