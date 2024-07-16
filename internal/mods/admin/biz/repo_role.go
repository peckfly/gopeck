package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"time"
)

const (
	RoleStatusEnabled    = "enabled"  // Enabled
	RoleStatusDisabled   = "disabled" // Disabled
	RoleResultTypeSelect = "select"   // Select
)

type (
	Roles []*Role
	Role  struct {
		ID          string    `json:"id" gorm:"size:20;primarykey;"` // Unique ID
		Code        string    `json:"code" gorm:"size:32;index;"`    // Code of role (unique)
		Name        string    `json:"name" gorm:"size:128;index"`    // Display name of role
		Description string    `json:"description" gorm:"size:1024"`  // Details about role
		Sequence    int       `json:"sequence" gorm:"index"`         // Sequence for sorting
		Status      string    `json:"status" gorm:"size:20;index"`   // Status of role (disabled, enabled)
		CreatedAt   time.Time `json:"created_at" gorm:"index;"`      // Create time
		UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`      // Update time
		Menus       RoleMenus `json:"menus" gorm:"-"`                // Role menu list
	}

	RoleQueryParam struct {
		common.PaginationParam
		LikeName    string     `form:"name"`                                       // Display name of role
		Status      string     `form:"status" binding:"oneof=disabled enabled ''"` // Status of role (disabled, enabled)
		ResultType  string     `form:"resultType"`                                 // Result type (options: select)
		InIDs       []string   `form:"-"`                                          // ID list
		GtUpdatedAt *time.Time `form:"-"`                                          // Update time is greater than
	}
	RoleQueryOptions struct {
		common.QueryOptions
	}
	RoleQueryResult struct {
		Data       Roles
		PageResult *common.PaginationResult
	}
	RoleForm struct {
		Code        string    `json:"code" binding:"required,max=32"`                   // Code of role (unique)
		Name        string    `json:"name" binding:"required,max=128"`                  // Display name of role
		Description string    `json:"description"`                                      // Details about role
		Sequence    int       `json:"sequence"`                                         // Sequence for sorting
		Status      string    `json:"status" binding:"required,oneof=disabled enabled"` // Status of role (enabled, disabled)
		Menus       RoleMenus `json:"menus"`                                            // Role menu list
	}

	RoleRepository interface {
		Query(ctx context.Context, params RoleQueryParam, opts ...RoleQueryOptions) (*RoleQueryResult, error)
		Get(ctx context.Context, id string, opts ...RoleQueryOptions) (*Role, error)
		Exists(ctx context.Context, id string) (bool, error)
		ExistsCode(ctx context.Context, code string) (bool, error)
		Create(ctx context.Context, item *Role) error
		Update(ctx context.Context, item *Role) error
		Delete(ctx context.Context, id string) error
	}
)

func (a *Role) TableName() string {
	return "role"
}

func (a *RoleForm) Validate() error {
	return nil
}

func (a *RoleForm) FillTo(role *Role) error {
	role.Code = a.Code
	role.Name = a.Name
	role.Description = a.Description
	role.Sequence = a.Sequence
	role.Status = a.Status
	return nil
}
