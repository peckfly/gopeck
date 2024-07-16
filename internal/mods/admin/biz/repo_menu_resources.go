package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"time"
)

type (
	MenuResources []*MenuResource
	MenuResource  struct {
		ID        string    `json:"id" gorm:"size:20;primarykey"` // Unique ID
		MenuID    string    `json:"menu_id" gorm:"size:20;index"` // From Menu.ID
		Method    string    `json:"method" gorm:"size:20;"`       // HTTP method
		Path      string    `json:"path" gorm:"size:255;"`        // API request path (e.g. /api/v1/users/:id)
		CreatedAt time.Time `json:"created_at" gorm:"index;"`     // Create time
		UpdatedAt time.Time `json:"updated_at" gorm:"index;"`     // Update time
	}

	MenuResourceQueryParam struct {
		common.PaginationParam
		MenuID  string   `form:"-"` // From Menu.ID
		MenuIDs []string `form:"-"` // From Menu.ID
	}

	MenuResourceQueryOptions struct {
		common.QueryOptions
	}

	MenuResourceQueryResult struct {
		Data       MenuResources
		PageResult *common.PaginationResult
	}

	MenuResourceForm struct {
	}

	MenuResourceRepository interface {
		Query(ctx context.Context, params MenuResourceQueryParam, opts ...MenuResourceQueryOptions) (*MenuResourceQueryResult, error)
		Get(ctx context.Context, id string, opts ...MenuResourceQueryOptions) (*MenuResource, error)
		Exists(ctx context.Context, id string) (bool, error)
		ExistsMethodPathByMenuID(ctx context.Context, method, path, menuID string) (bool, error)
		Create(ctx context.Context, item *MenuResource) error
		Update(ctx context.Context, item *MenuResource) error
		Delete(ctx context.Context, id string) error
		DeleteByMenuID(ctx context.Context, menuID string) error
	}
)

func (a *MenuResource) TableName() string {
	return "menu_resource"
}

func (a *MenuResourceForm) Validate() error {
	return nil
}

func (a *MenuResourceForm) FillTo(menuResource *MenuResource) error {
	return nil
}
