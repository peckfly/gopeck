package biz

import (
	"context"
	"fmt"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/pkg/cachex"
	"time"
)

type RoleUsecase struct {
	cache              cachex.Cache
	trans              *common.Trans
	roleRepository     RoleRepository
	roleMenuRepository RoleMenuRepository
	userRoleRepository UserRoleRepository
}

func NewRoleUsecase(
	cache cachex.Cache,
	trans *common.Trans,
	roleRepository RoleRepository,
	roleMenuRepository RoleMenuRepository,
	userRoleRepository UserRoleRepository,
) *RoleUsecase {
	return &RoleUsecase{
		cache:              cache,
		trans:              trans,
		roleRepository:     roleRepository,
		roleMenuRepository: roleMenuRepository,
		userRoleRepository: userRoleRepository,
	}
}

// Query roles from the data access object based on the provided parameters and options.
func (a *RoleUsecase) Query(ctx context.Context, params RoleQueryParam) (*RoleQueryResult, error) {
	params.Pagination = true

	var selectFields []string
	if params.ResultType == RoleResultTypeSelect {
		params.Pagination = false
		selectFields = []string{"id", "name"}
	}

	result, err := a.roleRepository.Query(ctx, params, RoleQueryOptions{
		QueryOptions: common.QueryOptions{
			OrderFields: []common.OrderByParam{
				{Field: "sequence", Direction: common.DESC},
				{Field: "created_at", Direction: common.DESC},
			},
			SelectFields: selectFields,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get the specified role from the data access object.
func (a *RoleUsecase) Get(ctx context.Context, id string) (*Role, error) {
	role, err := a.roleRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if role == nil {
		return nil, errors.NotFound("", "Role not found")
	}

	roleMenuResult, err := a.roleMenuRepository.Query(ctx, RoleMenuQueryParam{
		RoleID: id,
	})
	if err != nil {
		return nil, err
	}
	role.Menus = roleMenuResult.Data

	return role, nil
}

// Create a new role in the data access object.
func (a *RoleUsecase) Create(ctx context.Context, formItem *RoleForm) (*Role, error) {
	if exists, err := a.roleRepository.ExistsCode(ctx, formItem.Code); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Role code already exists")
	}

	role := &Role{
		ID:        common.NewXID(),
		CreatedAt: time.Now(),
	}
	if err := formItem.FillTo(role); err != nil {
		return nil, err
	}

	err := a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.roleRepository.Create(ctx, role); err != nil {
			return err
		}

		for _, roleMenu := range formItem.Menus {
			roleMenu.ID = common.NewXID()
			roleMenu.RoleID = role.ID
			roleMenu.CreatedAt = time.Now()
			if err := a.roleMenuRepository.Create(ctx, roleMenu); err != nil {
				return err
			}
		}
		return a.syncToCasbin(ctx)
	})
	if err != nil {
		return nil, err
	}
	role.Menus = formItem.Menus

	return role, nil
}

// Update the specified role in the data access object.
func (a *RoleUsecase) Update(ctx context.Context, id string, formItem *RoleForm) error {
	role, err := a.roleRepository.Get(ctx, id)
	if err != nil {
		return err
	} else if role == nil {
		return errors.NotFound("", "Role not found")
	} else if role.Code != formItem.Code {
		if exists, err := a.roleRepository.ExistsCode(ctx, formItem.Code); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Role code already exists")
		}
	}

	if err := formItem.FillTo(role); err != nil {
		return err
	}
	role.UpdatedAt = time.Now()

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.roleRepository.Update(ctx, role); err != nil {
			return err
		}
		if err := a.roleMenuRepository.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		for _, roleMenu := range formItem.Menus {
			if roleMenu.ID == "" {
				roleMenu.ID = common.NewXID()
			}
			roleMenu.RoleID = role.ID
			if roleMenu.CreatedAt.IsZero() {
				roleMenu.CreatedAt = time.Now()
			}
			roleMenu.UpdatedAt = time.Now()
			if err := a.roleMenuRepository.Create(ctx, roleMenu); err != nil {
				return err
			}
		}
		return a.syncToCasbin(ctx)
	})
}

// Delete the specified role from the data access object.
func (a *RoleUsecase) Delete(ctx context.Context, id string) error {
	exists, err := a.roleRepository.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Role not found")
	}

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.roleRepository.Delete(ctx, id); err != nil {
			return err
		}
		if err := a.roleMenuRepository.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		if err := a.userRoleRepository.DeleteByRoleID(ctx, id); err != nil {
			return err
		}

		return a.syncToCasbin(ctx)
	})
}

func (a *RoleUsecase) syncToCasbin(ctx context.Context) error {
	return a.cache.Set(ctx, CacheNSForRole, CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
