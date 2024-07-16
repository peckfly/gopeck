package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/crypto/hash"
	"time"
)

type UserUsecase struct {
	conf               *conf.RbacConf
	cache              cachex.Cache
	trans              *common.Trans
	userRepository     UserRepository
	userRoleRepository UserRoleRepository
}

func NewUserUsecase(
	cache cachex.Cache,
	trans *common.Trans,
	userRepository UserRepository,
	userRoleRepository UserRoleRepository,
	conf *conf.ServerConf,
) *UserUsecase {
	return &UserUsecase{
		cache:              cache,
		trans:              trans,
		userRepository:     userRepository,
		userRoleRepository: userRoleRepository,
		conf:               &conf.Rbac,
	}
}

// Query users from the data access object based on the provided parameters and options.
func (a *UserUsecase) Query(ctx context.Context, params UserQueryParam) (*UserQueryResult, error) {
	params.Pagination = true

	result, err := a.userRepository.Query(ctx, params, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			OrderFields: []common.OrderByParam{
				{Field: "created_at", Direction: common.DESC},
			},
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	}

	if userIDs := result.Data.ToIDs(); len(userIDs) > 0 {
		userRoleResult, err := a.userRoleRepository.Query(ctx, UserRoleQueryParam{
			InUserIDs: userIDs,
		}, UserRoleQueryOptions{
			JoinRole: true,
		})
		if err != nil {
			return nil, err
		}
		userRolesMap := userRoleResult.Data.ToUserIDMap()
		for _, user := range result.Data {
			user.Roles = userRolesMap[user.ID]
		}
	}

	return result, nil
}

// Get the specified user from the data access object.
func (a *UserUsecase) Get(ctx context.Context, id string) (*User, error) {
	user, err := a.userRepository.Get(ctx, id, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	userRoleResult, err := a.userRoleRepository.Query(ctx, UserRoleQueryParam{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

// Create a new user in the data access object.
func (a *UserUsecase) Create(ctx context.Context, formItem *UserForm) (*User, error) {
	existsUsername, err := a.userRepository.ExistsUsername(ctx, formItem.Username)
	if err != nil {
		return nil, err
	} else if existsUsername {
		return nil, errors.BadRequest("", "Username already exists")
	}

	user := &User{
		ID:        common.NewXID(),
		CreatedAt: time.Now(),
	}

	if formItem.Password == "" {
		formItem.Password = a.conf.DefaultLoginPwd
	}

	if err := formItem.FillTo(user); err != nil {
		return nil, err
	}

	err = a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.userRepository.Create(ctx, user); err != nil {
			return err
		}

		for _, userRole := range formItem.Roles {
			userRole.ID = common.NewXID()
			userRole.UserID = user.ID
			userRole.CreatedAt = time.Now()
			if err := a.userRoleRepository.Create(ctx, userRole); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	user.Roles = formItem.Roles

	return user, nil
}

// Update the specified user in the data access object.
func (a *UserUsecase) Update(ctx context.Context, id string, formItem *UserForm) error {
	user, err := a.userRepository.Get(ctx, id)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	} else if user.Username != formItem.Username {
		existsUsername, err := a.userRepository.ExistsUsername(ctx, formItem.Username)
		if err != nil {
			return err
		} else if existsUsername {
			return errors.BadRequest("", "Username already exists")
		}
	}

	if err := formItem.FillTo(user); err != nil {
		return err
	}
	user.UpdatedAt = time.Now()

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.userRepository.Update(ctx, user); err != nil {
			return err
		}

		if err := a.userRoleRepository.DeleteByUserID(ctx, id); err != nil {
			return err
		}
		for _, userRole := range formItem.Roles {
			if userRole.ID == "" {
				userRole.ID = common.NewXID()
			}
			userRole.UserID = user.ID
			if userRole.CreatedAt.IsZero() {
				userRole.CreatedAt = time.Now()
			}
			userRole.UpdatedAt = time.Now()
			if err := a.userRoleRepository.Create(ctx, userRole); err != nil {
				return err
			}
		}

		return a.cache.Delete(ctx, CacheNSForUser, id)
	})
}

// Delete the specified user from the data access object.
func (a *UserUsecase) Delete(ctx context.Context, id string) error {
	exists, err := a.userRepository.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.userRepository.Delete(ctx, id); err != nil {
			return err
		}
		if err := a.userRoleRepository.DeleteByUserID(ctx, id); err != nil {
			return err
		}
		return a.cache.Delete(ctx, CacheNSForUser, id)
	})
}

func (a *UserUsecase) ResetPassword(ctx context.Context, id string) error {
	exists, err := a.userRepository.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	hashPass, err := hash.GeneratePassword(a.conf.DefaultLoginPwd)
	if err != nil {
		return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
	}

	return a.trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.userRepository.UpdatePasswordByID(ctx, id, hashPass); err != nil {
			return err
		}
		return nil
	})
}

func (a *UserUsecase) GetRoleIDs(ctx context.Context, id string) ([]string, error) {
	userRoleResult, err := a.userRoleRepository.Query(ctx, UserRoleQueryParam{
		UserID: id,
	}, UserRoleQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"role_id"},
		},
	})
	if err != nil {
		return nil, err
	}
	return userRoleResult.Data.ToRoleIDs(), nil
}
