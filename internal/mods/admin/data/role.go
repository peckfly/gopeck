package data

import (
	"context"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"gorm.io/gorm"
)

type roleRepository struct {
	*gorm.DB
}

func NewRoleRepository(db *gorm.DB) biz.RoleRepository {
	return &roleRepository{db}
}

func GetRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(biz.Role))
}

// Query roles from the database based on the provided parameters and options.
func (a *roleRepository) Query(ctx context.Context, params biz.RoleQueryParam, opts ...biz.RoleQueryOptions) (*biz.RoleQueryResult, error) {
	var opt biz.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetRoleDB(ctx, a.DB)
	if v := params.InIDs; len(v) > 0 {
		db = db.Where("id IN (?)", v)
	}
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	if v := params.GtUpdatedAt; v != nil {
		db = db.Where("updated_at > ?", v)
	}

	var list biz.Roles
	pageResult, err := common.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &biz.RoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get the specified role from the database.
func (a *roleRepository) Get(ctx context.Context, id string, opts ...biz.RoleQueryOptions) (*biz.Role, error) {
	var opt biz.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(biz.Role)
	ok, err := common.FindOne(ctx, GetRoleDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists Exist checks if the specified role exists in the database.
func (a *roleRepository) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := common.Exists(ctx, GetRoleDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

func (a *roleRepository) ExistsCode(ctx context.Context, code string) (bool, error) {
	ok, err := common.Exists(ctx, GetRoleDB(ctx, a.DB).Where("code=?", code))
	return ok, errors.WithStack(err)
}

// Create a new role.
func (a *roleRepository) Create(ctx context.Context, item *biz.Role) error {
	result := GetRoleDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update the specified role in the database.
func (a *roleRepository) Update(ctx context.Context, item *biz.Role) error {
	result := GetRoleDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete the specified role from the database.
func (a *roleRepository) Delete(ctx context.Context, id string) error {
	result := GetRoleDB(ctx, a.DB).Where("id=?", id).Delete(new(biz.Role))
	return errors.WithStack(result.Error)
}
