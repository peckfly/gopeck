package data

import (
	"context"
	"fmt"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"gorm.io/gorm"
)

type (
	userRoleRepository struct {
		*gorm.DB
	}
)

func NewUserRoleRepository(db *gorm.DB) biz.UserRoleRepository {
	return &userRoleRepository{db}
}

func GetUserRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(biz.UserRole))
}

// Query user roles from the database based on the provided parameters and options.
func (a *userRoleRepository) Query(ctx context.Context, params biz.UserRoleQueryParam, opts ...biz.UserRoleQueryOptions) (*biz.UserRoleQueryResult, error) {
	var opt biz.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := a.DB.Table(fmt.Sprintf("%s AS a", new(biz.UserRole).TableName()))
	if opt.JoinRole {
		db = db.Joins(fmt.Sprintf("left join %s b on a.role_id=b.id", new(biz.Role).TableName()))
		db = db.Select("a.*,b.name as role_name")
	}

	if v := params.InUserIDs; len(v) > 0 {
		db = db.Where("a.user_id IN (?)", v)
	}
	if v := params.UserID; len(v) > 0 {
		db = db.Where("a.user_id = ?", v)
	}
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("a.role_id = ?", v)
	}

	var list biz.UserRoles
	pageResult, err := common.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &biz.UserRoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get the specified user role from the database.
func (a *userRoleRepository) Get(ctx context.Context, id string, opts ...biz.UserRoleQueryOptions) (*biz.UserRole, error) {
	var opt biz.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(biz.UserRole)
	ok, err := common.FindOne(ctx, GetUserRoleDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists Exist checks if the specified user role exists in the database.
func (a *userRoleRepository) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := common.Exists(ctx, GetUserRoleDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create a new user role.
func (a *userRoleRepository) Create(ctx context.Context, item *biz.UserRole) error {
	result := GetUserRoleDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update the specified user role in the database.
func (a *userRoleRepository) Update(ctx context.Context, item *biz.UserRole) error {
	result := GetUserRoleDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete the specified user role from the database.
func (a *userRoleRepository) Delete(ctx context.Context, id string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("id=?", id).Delete(new(biz.UserRole))
	return errors.WithStack(result.Error)
}

func (a *userRoleRepository) DeleteByUserID(ctx context.Context, userID string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("user_id=?", userID).Delete(new(biz.UserRole))
	return errors.WithStack(result.Error)
}

func (a *userRoleRepository) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("role_id=?", roleID).Delete(new(biz.UserRole))
	return errors.WithStack(result.Error)
}
