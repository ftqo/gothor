package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/discord-gophers/goapi-gen/types"
	"github.com/ftqo/gothor/db"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) *Response {
	body := &PostAuthLoginJSONRequestBody{}
	err := render.Bind(r, body)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to parse login request body")
		return &Response{
			Code: http.StatusBadRequest,
		}
	}

	user, err := s.DB.GetUserByEmail(r.Context(), string(body.Email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Log.Error().Err(err).Msg("User not found with the provided email")
			return &Response{
				Code: http.StatusUnauthorized,
			}
		}
		s.Log.Error().Err(err).Msg("Unable to retrieve user with the provided email")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	saltedPassword := make([]byte, len(body.Password)+len(user.Salt))
	copy(saltedPassword, body.Password)
	copy(saltedPassword[len(body.Password):], user.Salt)
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, saltedPassword)

	if err != nil {
		s.Log.Error().Err(err).Msg("Invalid email and password combination")
		return &Response{
			Code: http.StatusUnauthorized,
		}
	}

	err = s.Sessions.RenewToken(r.Context())
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to renew session token before login")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	s.Sessions.Put(r.Context(), "userID", user.ID.String())

	return PostAuthLoginJSON200Response(User{
		Username: user.Username,
		Email:    types.Email(user.Email),
		ID:       user.ID,
	})
}

func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) *Response {
	err := s.Sessions.RenewToken(r.Context())
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to renew token")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	err = s.Sessions.Destroy(r.Context())
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to destroy token")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}
	return &Response{
		Code: http.StatusNoContent,
	}
}

func (s *Server) PostAuthSignup(w http.ResponseWriter, r *http.Request) *Response {
	body := &PostAuthSignupJSONRequestBody{}
	err := render.Bind(r, body)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to parse signup request body")
		return &Response{
			Code: http.StatusBadRequest,
		}
	}

	_, err = s.DB.GetUserByEmail(r.Context(), string(body.Email))
	if err == nil {
		s.Log.Error().Err(err).Msg("Email already in use")
		return &Response{
			Code: http.StatusConflict,
		}
	} else if err != sql.ErrNoRows {
		s.Log.Error().Err(err).Msg("Unable to check if the email is already in use")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	salt, err := generateSalt()
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to generate salt for user signup")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	hashSaltPass, err := bcrypt.GenerateFromPassword(append([]byte(body.Password), salt...), 10)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to hash password and salt for user signup")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	user, err := s.DB.CreateUser(r.Context(), db.CreateUserParams{
		Email:          string(body.Email),
		HashedPassword: hashSaltPass,
		Salt:           salt,
	})
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to create user with the provided email")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	return PostAuthSignupJSON201Response(User{
		Username: user.Username,
		Email:    types.Email(user.Email),
		ID:       user.ID,
	})
}

func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request) *Response {
	posts, err := s.DB.GetAllPosts(r.Context())
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to retrieve all posts")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	var ps []Post
	for _, post := range posts {
		ps = append(ps, Post{
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			ID:        post.ID,
			Title:     post.Title,
			UpdatedAt: post.UpdatedAt,
		})
	}

	return GetPostsJSON200Response(ps)
}

func (s *Server) PostPosts(w http.ResponseWriter, r *http.Request) *Response {
	body := &PostPostsJSONRequestBody{}
	err := render.Bind(r, body)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to parse post creation request body")
		return &Response{
			Code: http.StatusBadRequest,
		}
	}

	userIDStr := s.Sessions.GetString(r.Context(), "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to parse user id from context")
		return &Response{
			Code: http.StatusUnauthorized,
		}
	}

	post, err := s.DB.CreatePost(r.Context(), db.CreatePostParams{
		UserID:  userID,
		Title:   body.Title,
		Content: body.Content,
	})
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to create new post")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	return PostPostsJSON201Response(Post{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

func (s *Server) DeletePostsPostID(w http.ResponseWriter, r *http.Request, postID uuid.UUID) *Response {
	err := s.DB.DeletePost(r.Context(), postID)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to delete post with the provided ID")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	return &Response{
		Code: http.StatusNoContent,
	}
}

func (s *Server) GetPostsPostID(w http.ResponseWriter, r *http.Request, postID uuid.UUID) *Response {
	post, err := s.DB.GetPostByID(r.Context(), postID)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to retrieve post with the provided ID")
		return &Response{
			Code: http.StatusNotFound,
		}
	}

	return GetPostsPostIDJSON200Response(Post{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}

func (s *Server) PutPostsPostID(w http.ResponseWriter, r *http.Request, postID uuid.UUID) *Response {
	body := &PutPostsPostIDJSONRequestBody{}
	err := render.Bind(r, body)
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to parse post update request body")
		return &Response{
			Code: http.StatusBadRequest,
		}
	}

	post, err := s.DB.UpdatePost(r.Context(), db.UpdatePostParams{
		ID:      body.ID,
		Title:   body.Title,
		Content: body.Content,
	})
	if err != nil {
		s.Log.Error().Err(err).Msg("Failed to update post with the provided ID")
		return &Response{
			Code: http.StatusInternalServerError,
		}
	}

	return PutPostsPostIDJSON200Response(Post{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	})
}
