package response

import (
	"plirevival/core/domain"
	"plirevival/core/port"
)

type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUserResponse(u domain.User) UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email}
}

func NewUsersResponse(users []domain.User) []UserResponse {
	res := make([]UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, NewUserResponse(u))
	}
	return res
}

type UserCreateResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      UserResponse `json:"data"`
}

type UserFetchResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      UserResponse `json:"data"`
}

type UsersListResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []UserResponse `json:"data"`
}
