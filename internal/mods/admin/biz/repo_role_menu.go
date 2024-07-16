package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"time"
)

type (
	RoleMenus []*RoleMenu

	RoleMenu struct {
		ID        string    `json:"id" gorm:"size:20;primarykey"` // Unique ID
		RoleID    string    `json:"role_id" gorm:"size:20;index"` // From Role.ID
		MenuID    string    `json:"menu_id" gorm:"size:20;index"` // From Menu.ID
		CreatedAt time.Time `json:"created_at" gorm:"index;"`     // Create time
		UpdatedAt time.Time `json:"updated_at" gorm:"index;"`     // Update time
	}

	RoleMenuQueryParam struct {
		common.PaginationParam
		RoleID string `form:"-"` // From Role.ID
	}

	RoleMenuQueryOptions struct {
		common.QueryOptions
	}

	RoleMenuQueryResult struct {
		Data       RoleMenus
		PageResult *common.PaginationResult
	}

	RoleMenuForm struct {
	}

	RoleMenuRepository interface {
		Query(ctx context.Context, params RoleMenuQueryParam, opts ...RoleMenuQueryOptions) (*RoleMenuQueryResult, error)
		Get(ctx context.Context, id string, opts ...RoleMenuQueryOptions) (*RoleMenu, error)
		Exists(ctx context.Context, id string) (bool, error)
		Create(ctx context.Context, item *RoleMenu) error
		Update(ctx context.Context, item *RoleMenu) error
		Delete(ctx context.Context, id string) error
		DeleteByRoleID(ctx context.Context, roleID string) error
		DeleteByMenuID(ctx context.Context, menuID string) error
	}
)

func (a *RoleMenu) TableName() string {
	return "role_menu"
}

func (a *RoleMenuForm) Validate() error {
	return nil
}

func (a *RoleMenuForm) FillTo(roleMenu *RoleMenu) error {
	return nil
}
