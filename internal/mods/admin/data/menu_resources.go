package data

import (
	"context"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"gorm.io/gorm"
)

type menuResourceRepository struct {
	*gorm.DB
}

func NewMenuResourceRepository(db *gorm.DB) biz.MenuResourceRepository {
	return &menuResourceRepository{db}
}

func GetMenuResourceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(biz.MenuResource))
}

// Query menu resources from the database based on the provided parameters and options.
func (a *menuResourceRepository) Query(ctx context.Context, params biz.MenuResourceQueryParam, opts ...biz.MenuResourceQueryOptions) (*biz.MenuResourceQueryResult, error) {
	var opt biz.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetMenuResourceDB(ctx, a.DB)
	if v := params.MenuID; len(v) > 0 {
		db = db.Where("menu_id = ?", v)
	}
	if v := params.MenuIDs; len(v) > 0 {
		db = db.Where("menu_id IN ?", v)
	}

	var list biz.MenuResources
	pageResult, err := common.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &biz.MenuResourceQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get the specified menu resource from the database.
func (a *menuResourceRepository) Get(ctx context.Context, id string, opts ...biz.MenuResourceQueryOptions) (*biz.MenuResource, error) {
	var opt biz.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(biz.MenuResource)
	ok, err := common.FindOne(ctx, GetMenuResourceDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists Exist checks if the specified menu resource exists in the database.
func (a *menuResourceRepository) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := common.Exists(ctx, GetMenuResourceDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsMethodPathByMenuID checks if the specified menu resource exists in the database.
func (a *menuResourceRepository) ExistsMethodPathByMenuID(ctx context.Context, method, path, menuID string) (bool, error) {
	ok, err := common.Exists(ctx, GetMenuResourceDB(ctx, a.DB).Where("method=? AND path=? AND menu_id=?", method, path, menuID))
	return ok, errors.WithStack(err)
}

// Create a new menu resource.
func (a *menuResourceRepository) Create(ctx context.Context, item *biz.MenuResource) error {
	result := GetMenuResourceDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update the specified menu resource in the database.
func (a *menuResourceRepository) Update(ctx context.Context, item *biz.MenuResource) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete the specified menu resource from the database.
func (a *menuResourceRepository) Delete(ctx context.Context, id string) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("id=?", id).Delete(new(biz.MenuResource))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID Deletes the menu resource by menu id.
func (a *menuResourceRepository) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("menu_id=?", menuID).Delete(new(biz.MenuResource))
	return errors.WithStack(result.Error)
}
