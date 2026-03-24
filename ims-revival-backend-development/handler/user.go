package handler

import (
	"math"
	"plirevival/core/port"
	resp "plirevival/handler/response"
	repo "plirevival/repo/postgres"

	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

type UserHandler struct {
	*serverHandler.Base
	svc *repo.UserRepository
}

func NewUserHandler(svc *repo.UserRepository) *UserHandler {
	base := serverHandler.New("Users").SetPrefix("/v1").AddPrefix("")
	return &UserHandler{Base: base, svc: svc}
}

func (h *UserHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/users", h.CreateUser).Name("Create User"),
		serverRoute.GET("/users", h.ListUsers).Name("List Users"),
		serverRoute.GET("/users/:id", h.GetUserByID).Name("Get User By ID"),
		serverRoute.PUT("/users/:id", h.UpdateUserByID).Name("Update User By ID"),
	}
}

func (h *UserHandler) CreateUser(sctx *serverRoute.Context, req CreateUserRequest) (*resp.UserCreateResponse, error) {
	u, err := h.svc.CreateUser(sctx.Ctx, req.Name, req.Email)
	if err != nil {
		return nil, err
	}
	r := &resp.UserCreateResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 resp.NewUserResponse(u),
	}
	return r, nil
}

func (h *UserHandler) ListUsers(sctx *serverRoute.Context, req ListUsersParams) (*resp.UsersListResponse, error) {
	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}
	users, err := h.svc.GetAllUsers(sctx.Ctx, req.Skip, req.Limit)
	if err != nil {
		return nil, err
	}
	data := resp.NewUsersResponse(users)
	md := port.NewMetaDataResponse(req.Skip, req.Limit, uint64(len(data)))
	r := &resp.UsersListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     md,
		Data:                 data,
	}
	return r, nil
}

func (h *UserHandler) GetUserByID(sctx *serverRoute.Context, req UserIDUri) (*resp.UserFetchResponse, error) {
	u, err := h.svc.GetUserByID(sctx.Ctx, req.ID)
	if err != nil {
		return nil, err
	}
	r := &resp.UserFetchResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 resp.NewUserResponse(u),
	}
	return r, nil
}

func (h *UserHandler) UpdateUserByID(sctx *serverRoute.Context, req UpdateUserRequest) (*resp.UserFetchResponse, error) {
	var namePtr, emailPtr *string
	if req.Name != "" {
		namePtr = &req.Name
	}
	if req.Email != "" {
		emailPtr = &req.Email
	}
	u, err := h.svc.UpdateUserByID(sctx.Ctx, req.ID, namePtr, emailPtr)
	if err != nil {
		return nil, err
	}
	r := &resp.UserFetchResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data:                 resp.NewUserResponse(u),
	}
	return r, nil
}
