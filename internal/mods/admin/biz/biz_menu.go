package biz

import (
	"context"
	"fmt"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/pkg/cachex"
	"sort"
	"strings"
	"time"
)

type MenuUsecase struct {
	cache                  cachex.Cache
	trans                  *common.Trans
	menuRepository         MenuRepository
	menuResourceRepository MenuResourceRepository
	roleMenuRepository     RoleMenuRepository
	conf                   *conf.RbacConf
}

// NewMenuUsecase initializes a new MenuUsecase with the provided dependencies.
//
// cache: Cache interface for caching.
// trans: Trans pointer for translations.
// menuRepository: MenuRepository interface for menu operations.
// menuResourceRepository: MenuResourceRepository interface for menu resource operations.
// roleMenuRepository: RoleMenuRepository interface for role menu operations.
// conf: ServerConf pointer for server configuration.
// Returns a pointer to the initialized MenuUsecase.
func NewMenuUsecase(
	cache cachex.Cache,
	trans *common.Trans,
	menuRepository MenuRepository,
	menuResourceRepository MenuResourceRepository,
	roleMenuRepository RoleMenuRepository,
	conf *conf.ServerConf,
) *MenuUsecase {
	return &MenuUsecase{
		cache:                  cache,
		trans:                  trans,
		menuRepository:         menuRepository,
		menuResourceRepository: menuResourceRepository,
		roleMenuRepository:     roleMenuRepository,
		conf:                   &conf.Rbac,
	}
}

// CreateInBatchByParent creates multiple menu items in batch under a parent menu.
//
// ctx: the context for the operation.
// items: the list of menu items to create.
// parent: the parent menu under which the items will be created.
// error: an error if any occurred during the operation.
func (a *MenuUsecase) CreateInBatchByParent(ctx context.Context, items Menus, parent *Menu) error {
	total := len(items)
	for i, item := range items {
		var parentID string
		if parent != nil {
			parentID = parent.ID
		}

		exist := false

		if item.ID != "" {
			exists, err := a.menuRepository.Exists(ctx, item.ID)
			if err != nil {
				return err
			} else if exists {
				exist = true
			}
		} else if item.Code != "" {
			exists, err := a.menuRepository.ExistsCodeByParentID(ctx, item.Code, parentID)
			if err != nil {
				return err
			} else if exists {
				exist = true
				existItem, err := a.menuRepository.GetByCodeAndParentID(ctx, item.Code, parentID)
				if err != nil {
					return err
				}
				if existItem != nil {
					item.ID = existItem.ID
				}
			}
		} else if item.Name != "" {
			exists, err := a.menuRepository.ExistsNameByParentID(ctx, item.Name, parentID)
			if err != nil {
				return err
			} else if exists {
				exist = true
				existItem, err := a.menuRepository.GetByNameAndParentID(ctx, item.Name, parentID)
				if err != nil {
					return err
				}
				if existItem != nil {
					item.ID = existItem.ID
				}
			}
		}

		if !exist {
			if item.ID == "" {
				item.ID = common.NewXID()
			}
			if item.Status == "" {
				item.Status = MenuStatusEnabled
			}
			if item.Sequence == 0 {
				item.Sequence = total - i
			}

			item.ParentID = parentID
			if parent != nil {
				item.ParentPath = parent.ParentPath + parentID + common.TreePathDelimiter
			}
			item.CreatedAt = time.Now()

			if err := a.menuRepository.Create(ctx, item); err != nil {
				return err
			}
		}

		for _, res := range item.Resources {
			if res.ID != "" {
				exists, err := a.menuResourceRepository.Exists(ctx, res.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}

			if res.Path != "" {
				exists, err := a.menuResourceRepository.ExistsMethodPathByMenuID(ctx, res.Method, res.Path, item.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}

			if res.ID == "" {
				res.ID = common.NewXID()
			}
			res.MenuID = item.ID
			res.CreatedAt = time.Now()
			if err := a.menuResourceRepository.Create(ctx, res); err != nil {
				return err
			}
		}

		if item.Children != nil {
			if err := a.CreateInBatchByParent(ctx, *item.Children, item); err != nil {
				return err
			}
		}
	}
	return nil
}

// Query menus from the data access object based on the provided parameters and options.
func (a *MenuUsecase) Query(ctx context.Context, params MenuQueryParam) (*MenuQueryResult, error) {
	params.Pagination = false

	if err := a.fillQueryParam(ctx, &params); err != nil {
		return nil, err
	}

	result, err := a.menuRepository.Query(ctx, params, MenuQueryOptions{
		QueryOptions: common.QueryOptions{
			OrderFields: MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	}

	if params.LikeName != "" || params.CodePath != "" {
		result.Data, err = a.appendChildren(ctx, result.Data)
		if err != nil {
			return nil, err
		}
	}

	if params.IncludeResources {
		for i, item := range result.Data {
			resResult, err := a.menuResourceRepository.Query(ctx, MenuResourceQueryParam{
				MenuID: item.ID,
			})
			if err != nil {
				return nil, err
			}
			result.Data[i].Resources = resResult.Data
		}
	}

	result.Data = result.Data.ToTree()
	return result, nil
}

// fillQueryParam fills the query parameters for the MenuUsecase.
//
// ctx: the context.Context for the operation.
// params: the MenuQueryParam struct containing the query parameters.
// error: an error if any occurred during the operation.
func (a *MenuUsecase) fillQueryParam(ctx context.Context, params *MenuQueryParam) error {
	if params.CodePath != "" {
		var (
			codes    []string
			lastMenu Menu
		)
		for _, code := range strings.Split(params.CodePath, common.TreePathDelimiter) {
			if code == "" {
				continue
			}
			codes = append(codes, code)
			menu, err := a.menuRepository.GetByCodeAndParentID(ctx, code, lastMenu.ParentID, MenuQueryOptions{
				QueryOptions: common.QueryOptions{
					SelectFields: []string{"id", "parent_id", "parent_path"},
				},
			})
			if err != nil {
				return err
			} else if menu == nil {
				return errors.NotFound("", "Menu not found by code '%s'", strings.Join(codes, common.TreePathDelimiter))
			}
			lastMenu = *menu
		}
		params.ParentPathPrefix = lastMenu.ParentPath + lastMenu.ID + common.TreePathDelimiter
	}
	return nil
}

// appendChildren appends children to the Menus slice in the MenuUsecase struct.
//
// It takes a context.Context and a Menus slice as input parameters.
// It returns a Menus slice and an error.
func (a *MenuUsecase) appendChildren(ctx context.Context, data Menus) (Menus, error) {
	if len(data) == 0 {
		return data, nil
	}

	existsInData := func(id string) bool {
		for _, item := range data {
			if item.ID == id {
				return true
			}
		}
		return false
	}

	for _, item := range data {
		childResult, err := a.menuRepository.Query(ctx, MenuQueryParam{
			ParentPathPrefix: item.ParentPath + item.ID + common.TreePathDelimiter,
		})
		if err != nil {
			return nil, err
		}
		for _, child := range childResult.Data {
			if existsInData(child.ID) {
				continue
			}
			data = append(data, child)
		}
	}

	if parentIDs := data.SplitParentIDs(); len(parentIDs) > 0 {
		parentResult, err := a.menuRepository.Query(ctx, MenuQueryParam{
			InIDs: parentIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range parentResult.Data {
			if existsInData(p.ID) {
				continue
			}
			data = append(data, p)
		}
	}
	sort.Sort(data)

	return data, nil
}

// Get the specified menu from the data access object.
func (a *MenuUsecase) Get(ctx context.Context, id string) (*Menu, error) {
	menu, err := a.menuRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if menu == nil {
		return nil, errors.NotFound("", "Menu not found")
	}

	menuResResult, err := a.menuResourceRepository.Query(ctx, MenuResourceQueryParam{
		MenuID: menu.ID,
	})
	if err != nil {
		return nil, err
	}
	menu.Resources = menuResResult.Data

	return menu, nil
}

// Create a new menu in the data access object.
func (a *MenuUsecase) Create(ctx context.Context, formItem *MenuForm) (*Menu, error) {
	menu := &Menu{
		ID:        common.NewXID(),
		CreatedAt: time.Now(),
	}

	if parentID := formItem.ParentID; parentID != "" {
		parent, err := a.menuRepository.Get(ctx, parentID)
		if err != nil {
			return nil, err
		} else if parent == nil {
			return nil, errors.NotFound("", "Parent not found")
		}
		menu.ParentPath = parent.ParentPath + parent.ID + common.TreePathDelimiter
	}

	if exists, err := a.menuRepository.ExistsCodeByParentID(ctx, formItem.Code, formItem.ParentID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Menu code already exists at the same level")
	}

	if err := formItem.FillTo(menu); err != nil {
		return nil, err
	}

	err := a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.menuRepository.Create(ctx, menu); err != nil {
			return err
		}

		for _, res := range formItem.Resources {
			res.ID = common.NewXID()
			res.MenuID = menu.ID
			res.CreatedAt = time.Now()
			if err := a.menuResourceRepository.Create(ctx, res); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return menu, nil
}

// Update the specified menu in the data access object.
func (a *MenuUsecase) Update(ctx context.Context, id string, formItem *MenuForm) error {
	menu, err := a.menuRepository.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	oldParentPath := menu.ParentPath
	oldStatus := menu.Status
	var childData Menus
	if menu.ParentID != formItem.ParentID {
		if parentID := formItem.ParentID; parentID != "" {
			parent, err := a.menuRepository.Get(ctx, parentID)
			if err != nil {
				return err
			} else if parent == nil {
				return errors.NotFound("", "Parent not found")
			}
			menu.ParentPath = parent.ParentPath + parent.ID + common.TreePathDelimiter
		} else {
			menu.ParentPath = ""
		}

		childResult, err := a.menuRepository.Query(ctx, MenuQueryParam{
			ParentPathPrefix: oldParentPath + menu.ID + common.TreePathDelimiter,
		}, MenuQueryOptions{
			QueryOptions: common.QueryOptions{
				SelectFields: []string{"id", "parent_path"},
			},
		})
		if err != nil {
			return err
		}
		childData = childResult.Data
	}

	if menu.Code != formItem.Code {
		if exists, err := a.menuRepository.ExistsCodeByParentID(ctx, formItem.Code, formItem.ParentID); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Menu code already exists at the same level")
		}
	}

	if err := formItem.FillTo(menu); err != nil {
		return err
	}

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if oldStatus != formItem.Status {
			oldPath := oldParentPath + menu.ID + common.TreePathDelimiter
			if err := a.menuRepository.UpdateStatusByParentPath(ctx, oldPath, formItem.Status); err != nil {
				return err
			}
		}

		for _, child := range childData {
			oldPath := oldParentPath + menu.ID + common.TreePathDelimiter
			newPath := menu.ParentPath + menu.ID + common.TreePathDelimiter
			err := a.menuRepository.UpdateParentPath(ctx, child.ID, strings.Replace(child.ParentPath, oldPath, newPath, 1))
			if err != nil {
				return err
			}
		}

		if err := a.menuRepository.Update(ctx, menu); err != nil {
			return err
		}

		if err := a.menuResourceRepository.DeleteByMenuID(ctx, id); err != nil {
			return err
		}
		for _, res := range formItem.Resources {
			if res.ID == "" {
				res.ID = common.NewXID()
			}
			res.MenuID = id
			if res.CreatedAt.IsZero() {
				res.CreatedAt = time.Now()
			}
			res.UpdatedAt = time.Now()
			if err := a.menuResourceRepository.Create(ctx, res); err != nil {
				return err
			}
		}

		return a.syncToCasbin(ctx)
	})
}

// Delete the specified menu from the data access object.
func (a *MenuUsecase) Delete(ctx context.Context, id string) error {
	if a.conf.DenyDeleteMenu {
		return errors.BadRequest("", "Menu deletion is not allowed")
	}

	menu, err := a.menuRepository.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	childResult, err := a.menuRepository.Query(ctx, MenuQueryParam{
		ParentPathPrefix: menu.ParentPath + menu.ID + common.TreePathDelimiter,
	}, MenuQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"id"},
		},
	})
	if err != nil {
		return err
	}

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.delete(ctx, id); err != nil {
			return err
		}

		for _, child := range childResult.Data {
			if err := a.delete(ctx, child.ID); err != nil {
				return err
			}
		}
		return a.syncToCasbin(ctx)
	})
}

func (a *MenuUsecase) delete(ctx context.Context, id string) error {
	if err := a.menuRepository.Delete(ctx, id); err != nil {
		return err
	}
	if err := a.menuResourceRepository.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	if err := a.roleMenuRepository.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return nil
}

// syncToCasbin synchronizes the menus to the casbin rules.
// fix the menu update or delete did not update casbin problem
func (a *MenuUsecase) syncToCasbin(ctx context.Context) error {
	return a.cache.Set(ctx, CacheNSForRole, CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
