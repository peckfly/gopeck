package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"time"
)

type (
	UserRoles []*UserRole
	UserRole  struct {
		ID        string    `json:"id" gorm:"size:20;primarykey"`           // Unique ID
		UserID    string    `json:"user_id" gorm:"size:20;index"`           // From User.ID
		RoleID    string    `json:"role_id" gorm:"size:20;index"`           // From Role.ID
		CreatedAt time.Time `json:"created_at" gorm:"index;"`               // Create time
		UpdatedAt time.Time `json:"updated_at" gorm:"index;"`               // Update time
		RoleName  string    `json:"role_name" gorm:"<-:false;-:migration;"` // From Role.Name
	}
	UserRoleQueryParam struct {
		common.PaginationParam
		InUserIDs []string `form:"-"` // From User.ID
		UserID    string   `form:"-"` // From User.ID
		RoleID    string   `form:"-"` // From Role.ID
	}
	UserRoleQueryOptions struct {
		common.QueryOptions
		JoinRole bool // Join role table
	}
	UserRoleQueryResult struct {
		Data       UserRoles
		PageResult *common.PaginationResult
	}
	UserRoleForm struct {
	}
	UserRoleRepository interface {
		Query(ctx context.Context, param UserRoleQueryParam, options ...UserRoleQueryOptions) (*UserRoleQueryResult, error)
		Get(ctx context.Context, id string, opts ...UserRoleQueryOptions) (*UserRole, error)
		Exists(ctx context.Context, id string) (bool, error)
		Create(ctx context.Context, item *UserRole) error
		Update(ctx context.Context, item *UserRole) error
		Delete(ctx context.Context, id string) error
		DeleteByUserID(ctx context.Context, userID string) error
		DeleteByRoleID(ctx context.Context, roleID string) error
	}
)

func (a *UserRole) TableName() string {
	return "user_role"
}

func (a UserRoles) ToRoleIDs() []string {
	var ids []string
	for _, item := range a {
		ids = append(ids, item.RoleID)
	}
	return ids
}

func (a UserRoles) ToUserIDMap() map[string]UserRoles {
	m := make(map[string]UserRoles)
	for _, userRole := range a {
		m[userRole.UserID] = append(m[userRole.UserID], userRole)
	}
	return m
}

func (a *UserRoleForm) Validate() error {
	return nil
}

func (a *UserRoleForm) FillTo(userRole *UserRole) error {
	return nil
}
