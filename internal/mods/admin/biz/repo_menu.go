package biz

import (
	"context"
	"encoding/json"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"strings"
	"time"
)

const (
	MenuStatusDisabled = "disabled"
	MenuStatusEnabled  = "enabled"
)

var (
	MenusOrderParams = []common.OrderByParam{
		{Field: "sequence", Direction: common.DESC},
		{Field: "created_at", Direction: common.DESC},
	}
)

type (
	Menus []*Menu
	Menu  struct {
		ID          string        `json:"id" gorm:"size:20;primarykey;"`      // Unique ID
		Code        string        `json:"code" gorm:"size:32;index;"`         // Code of menu (unique for each level)
		Name        string        `json:"name" gorm:"size:128;index"`         // Display name of menu
		Description string        `json:"description" gorm:"size:1024"`       // Details about menu
		Sequence    int           `json:"sequence" gorm:"index;"`             // Sequence for sorting (Order by desc)
		Type        string        `json:"type" gorm:"size:20;index"`          // Type of menu (page, button)
		Path        string        `json:"path" gorm:"size:255;"`              // Access path of menu
		Properties  string        `json:"properties" gorm:"type:text;"`       // Properties of menu (JSON)
		Status      string        `json:"status" gorm:"size:20;index"`        // Status of menu (enabled, disabled)
		ParentID    string        `json:"parent_id" gorm:"size:20;index;"`    // Parent ID (From Menu.ID)
		ParentPath  string        `json:"parent_path" gorm:"size:255;index;"` // Parent path (split by .)
		Children    *Menus        `json:"children" gorm:"-"`                  // Child menus
		CreatedAt   time.Time     `json:"created_at" gorm:"index;"`           // Create time
		UpdatedAt   time.Time     `json:"updated_at" gorm:"index;"`           // Update time
		Resources   MenuResources `json:"resources" gorm:"-"`                 // Resources of menu
	}

	MenuQueryParam struct {
		common.PaginationParam
		CodePath         string   `form:"code"`             // Code path (like xxx.xxx.xxx)
		LikeName         string   `form:"name"`             // Display name of menu
		IncludeResources bool     `form:"includeResources"` // Include resources
		InIDs            []string `form:"-"`                // Include menu IDs
		Status           string   `form:"-"`                // Status of menu (disabled, enabled)
		ParentID         string   `form:"-"`                // Parent ID (From Menu.ID)
		ParentPathPrefix string   `form:"-"`                // Parent path (split by .)
		UserID           string   `form:"-"`                // User ID
		RoleID           string   `form:"-"`                // Role ID
	}

	MenuQueryOptions struct {
		common.QueryOptions
	}

	MenuQueryResult struct {
		Data       Menus
		PageResult *common.PaginationResult
	}

	MenuForm struct {
		Code        string        `json:"code" binding:"required,max=32"`                   // Code of menu (unique for each level)
		Name        string        `json:"name" binding:"required,max=128"`                  // Display name of menu
		Description string        `json:"description"`                                      // Details about menu
		Sequence    int           `json:"sequence"`                                         // Sequence for sorting (Order by desc)
		Type        string        `json:"type" binding:"required,oneof=page button"`        // Type of menu (page, button)
		Path        string        `json:"path"`                                             // Access path of menu
		Properties  string        `json:"properties"`                                       // Properties of menu (JSON)
		Status      string        `json:"status" binding:"required,oneof=disabled enabled"` // Status of menu (enabled, disabled)
		ParentID    string        `json:"parent_id"`                                        // Parent ID (From Menu.ID)
		Resources   MenuResources `json:"resources"`                                        // Resources of menu
	}

	MenuRepository interface {
		Query(ctx context.Context, params MenuQueryParam, opts ...MenuQueryOptions) (*MenuQueryResult, error)
		Get(ctx context.Context, id string, opts ...MenuQueryOptions) (*Menu, error)
		GetByCodeAndParentID(ctx context.Context, code, parentID string, opts ...MenuQueryOptions) (*Menu, error)
		GetByNameAndParentID(ctx context.Context, name, parentID string, opts ...MenuQueryOptions) (*Menu, error)
		Exists(ctx context.Context, id string) (bool, error)
		ExistsCodeByParentID(ctx context.Context, code, parentID string) (bool, error)
		ExistsNameByParentID(ctx context.Context, name, parentID string) (bool, error)
		Create(ctx context.Context, item *Menu) error
		Update(ctx context.Context, item *Menu) error
		Delete(ctx context.Context, id string) error
		UpdateParentPath(ctx context.Context, id, parentPath string) error
		UpdateStatusByParentPath(ctx context.Context, parentPath, status string) error
	}
)

func (a *Menu) TableName() string {
	return "menu"
}

func (a Menus) Len() int {
	return len(a)
}

func (a Menus) Less(i, j int) bool {
	if a[i].Sequence == a[j].Sequence {
		return a[i].CreatedAt.Unix() > a[j].CreatedAt.Unix()
	}
	return a[i].Sequence > a[j].Sequence
}

func (a Menus) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Menus) ToMap() map[string]*Menu {
	m := make(map[string]*Menu)
	for _, item := range a {
		m[item.ID] = item
	}
	return m
}

func (a Menus) SplitParentIDs() []string {
	parentIDs := make([]string, 0, len(a))
	idMapper := make(map[string]struct{})
	for _, item := range a {
		if _, ok := idMapper[item.ID]; ok {
			continue
		}
		idMapper[item.ID] = struct{}{}
		if pp := item.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, common.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := idMapper[pid]; ok {
					continue
				}
				parentIDs = append(parentIDs, pid)
				idMapper[pid] = struct{}{}
			}
		}
	}
	return parentIDs
}

func (a Menus) ToTree() Menus {
	var list Menus
	m := a.ToMap()
	for _, item := range a {
		if item.ParentID == "" {
			list = append(list, item)
			continue
		}
		if parent, ok := m[item.ParentID]; ok {
			if parent.Children == nil {
				children := Menus{item}
				parent.Children = &children
				continue
			}
			*parent.Children = append(*parent.Children, item)
		}
	}
	return list
}

// Validate A validation function for the `MenuForm` struct.
func (a *MenuForm) Validate() error {
	if v := a.Properties; v != "" {
		if !json.Valid([]byte(v)) {
			return errors.BadRequest("", "invalid properties")
		}
	}
	return nil
}

func (a *MenuForm) FillTo(menu *Menu) error {
	menu.Code = a.Code
	menu.Name = a.Name
	menu.Description = a.Description
	menu.Sequence = a.Sequence
	menu.Type = a.Type
	menu.Path = a.Path
	menu.Properties = a.Properties
	menu.Status = a.Status
	menu.ParentID = a.ParentID
	return nil
}
