package biz

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/pkg/crypto/hash"
	"time"
)

const (
	UserStatusActivated = "activated"
	UserStatusFreezed   = "freezed"
)

type (
	Users []*User
	User  struct {
		ID        string    `json:"id" gorm:"size:20;primarykey;"` // Unique ID
		Username  string    `json:"username" gorm:"size:64;index"` // Username for login
		Name      string    `json:"name" gorm:"size:64;index"`     // Name of user
		Password  string    `json:"-" gorm:"size:64;"`             // Password for login (encrypted)
		Phone     string    `json:"phone" gorm:"size:32;"`         // Phone number of user
		Email     string    `json:"email" gorm:"size:128;"`        // Email of user
		Remark    string    `json:"remark" gorm:"size:1024;"`      // Remark of user
		Status    string    `json:"status" gorm:"size:20;index"`   // Status of user (activated, freezed)
		CreatedAt time.Time `json:"created_at" gorm:"index;"`      // Create time
		UpdatedAt time.Time `json:"updated_at" gorm:"index;"`      // Update time
		Roles     UserRoles `json:"roles" gorm:"-"`                // Roles of user
	}

	UserQueryOptions struct {
		common.QueryOptions
	}

	UserQueryParam struct {
		common.PaginationParam
		LikeUsername string `form:"username"`                                    // Username for login
		LikeName     string `form:"name"`                                        // Name of user
		Status       string `form:"status" binding:"oneof=activated freezed ''"` // Status of user (activated, freezed)
	}

	UserQueryResult struct {
		Data       Users
		PageResult *common.PaginationResult
	}

	UserForm struct {
		Username string    `json:"username" binding:"required,max=64"`                // Username for login
		Name     string    `json:"name" binding:"required,max=64"`                    // Name of user
		Password string    `json:"password" binding:"max=64"`                         // Password for login (md5 hash)
		Phone    string    `json:"phone" binding:"max=32"`                            // Phone number of user
		Email    string    `json:"email" binding:"max=128"`                           // Email of user
		Remark   string    `json:"remark" binding:"max=1024"`                         // Remark of user
		Status   string    `json:"status" binding:"required,oneof=activated freezed"` // Status of user (activated, freezed)
		Roles    UserRoles `json:"roles" binding:"required"`                          // Roles of user
	}

	UserRepository interface {
		Query(ctx context.Context, params UserQueryParam, opts ...UserQueryOptions) (*UserQueryResult, error)
		Get(ctx context.Context, id string, opts ...UserQueryOptions) (*User, error)
		GetByUsername(ctx context.Context, username string, opts ...UserQueryOptions) (*User, error)
		Exists(ctx context.Context, id string) (bool, error)
		ExistsUsername(ctx context.Context, username string) (bool, error)
		Create(ctx context.Context, item *User) error
		Update(ctx context.Context, item *User, selectFields ...string) error
		Delete(ctx context.Context, id string) error
		UpdatePasswordByID(ctx context.Context, id string, password string) error
	}
)

func (a *User) TableName() string {
	return "user"
}

func (a Users) ToIDs() []string {
	var ids []string
	for _, item := range a {
		ids = append(ids, item.ID)
	}
	return ids
}

func (a *UserForm) Validate() error {
	if a.Email != "" && validator.New().Var(a.Email, "email") != nil {
		return errors.BadRequest("", "Invalid email address")
	}
	return nil
}

func (a *UserForm) FillTo(user *User) error {
	user.Username = a.Username
	user.Name = a.Name
	user.Phone = a.Phone
	user.Email = a.Email
	user.Remark = a.Remark
	user.Status = a.Status

	if pass := a.Password; pass != "" {
		hashPass, err := hash.GeneratePassword(pass)
		if err != nil {
			return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
		}
		user.Password = hashPass
	}

	return nil
}
