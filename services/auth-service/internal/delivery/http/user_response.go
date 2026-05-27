package http

import (
	"time"

	"github.com/TheDeutsch13/b2-project/services/auth-service/internal/domain"
)

type userResponse struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Nickname  string  `json:"nickname"`
	BirthDate *string `json:"birth_date,omitempty"`
	Gender    string  `json:"gender"`
	Phone     string  `json:"phone"`
	AvatarURL string  `json:"avatar_url"`
	CreatedAt string  `json:"created_at"`
}

type publicUserResponse struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

func toUserResponse(user *domain.User) userResponse {
	response := userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Gender:    user.Gender,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}

	if user.BirthDate != nil {
		formatted := user.BirthDate.Format("2006-01-02")
		response.BirthDate = &formatted
	}

	return response
}

func toUserResponses(users []domain.User) []userResponse {
	response := make([]userResponse, 0, len(users))
	for index := range users {
		response = append(response, toUserResponse(&users[index]))
	}

	return response
}

func toPublicUserResponses(users []domain.User) []publicUserResponse {
	response := make([]publicUserResponse, 0, len(users))
	for index := range users {
		user := users[index]
		response = append(response, publicUserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
		})
	}

	return response
}
