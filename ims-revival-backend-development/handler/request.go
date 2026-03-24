package handler

import (
	"plirevival/core/port"
)

type DetailTableGetParams struct {
	EmpID  int64  `form:"emp-id" validate:"required"`
	PostID string `form:"post-id" validate:"omitempty"`
	port.MetadataRequest
}

type MakerGetParams struct {
	EmpID  int64  `form:"emp-id" validate:"required"`
	AOID   int64  `form:"admin-office-id" validate:"omitempty,office_id"`
	PostID string `form:"post-id" validate:"omitempty"`
	Status string `form:"status" validate:"omitempty,oneof=Pending Forwarded Approved Rejected Cancelled UnApproved"`
	port.MetadataRequest
}

type ApproveAwardsParams struct {
	AwardIDs        []int  `json:"award_ids" validate:"required"`
	ApprovedBy      string `json:"approved_by" validate:"required"`
	ApproveStatus   string `json:"approve_status" validate:"required"`
	ApproverRemarks string `json:"approve_auth_remarks" validate:"required,remarks"`
}

// CreateUserRequest represents the payload to create a user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
}

// UpdateUserRequest represents the payload to update a user (all fields optional)
type UpdateUserRequest struct {
	ID    int64  `json:"id" validate:"required"`
	Name  string `json:"name" validate:"omitempty,min=1,max=100"`
	Email string `json:"email" validate:"omitempty,email,max=255"`
}

// Uri struct for id
type UserIDUri struct {
	ID int64 `uri:"id" validate:"required"`
}

type ListUsersParams struct {
	port.MetadataRequest
}

func (p *ListUsersParams) Validate() error {
	return nil
}
