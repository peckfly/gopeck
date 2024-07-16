package data

import (
	"context"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"gorm.io/gorm"
)

type roleMenuRepository struct {
	*gorm.DB
}

func NewRoleMenuRepository(db *gorm.DB) biz.RoleMenuRepository {
	return &roleMenuRepository{db}
}

func GetRoleMenuDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(biz.RoleMenu))
}

// Query role menus from the database based on the provided parameters and options.
func (a *roleMenuRepository) Query(ctx context.Context, params biz.RoleMenuQueryParam, opts ...biz.RoleMenuQueryOptions) (*biz.RoleMenuQueryResult, error) {
	var opt biz.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetRoleMenuDB(ctx, a.DB)
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("role_id = ?", v)
	}

	var list biz.RoleMenus
	pageResult, err := common.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &biz.RoleMenuQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get the specified role menu from the database.
func (a *roleMenuRepository) Get(ctx context.Context, id string, opts ...biz.RoleMenuQueryOptions) (*biz.RoleMenu, error) {
	var opt biz.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(biz.RoleMenu)
	ok, err := common.FindOne(ctx, GetRoleMenuDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists Exist checks if the specified role menu exists in the database.
func (a *roleMenuRepository) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := common.Exists(ctx, GetRoleMenuDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create a new role menu.
func (a *roleMenuRepository) Create(ctx context.Context, item *biz.RoleMenu) error {
	result := GetRoleMenuDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update the specified role menu in the database.
func (a *roleMenuRepository) Update(ctx context.Context, item *biz.RoleMenu) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete the specified role menu from the database.
func (a *roleMenuRepository) Delete(ctx context.Context, id string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("id=?", id).Delete(new(biz.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByRoleID Deletes role menus by role id.
func (a *roleMenuRepository) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("role_id=?", roleID).Delete(new(biz.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID Deletes role menus by menu id.
func (a *roleMenuRepository) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("menu_id=?", menuID).Delete(new(biz.RoleMenu))
	return errors.WithStack(result.Error)
}
