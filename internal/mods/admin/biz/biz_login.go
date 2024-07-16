package biz

import (
	"context"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/internal/pkg/jwtx"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/crypto/hash"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"time"
)

const (
	captchaLength = 4
	captchaWidth  = 400
	captchaHeight = 160
)

type LoginUsecase struct {
	cache              cachex.Cache
	conf               *conf.RbacConf
	userRepository     UserRepository
	auth               jwtx.Auther
	userUsecase        *UserUsecase
	userRoleRepository UserRoleRepository
	menuRepository     MenuRepository
}

func NewLoginUsecase(
	cache cachex.Cache,
	conf *conf.ServerConf,
	userRepository UserRepository,
	auth jwtx.Auther,
	userUsecase *UserUsecase,
	userRoleRepository UserRoleRepository,
	menuRepository MenuRepository,
) *LoginUsecase {
	return &LoginUsecase{
		cache:              cache,
		conf:               &conf.Rbac,
		userRepository:     userRepository,
		auth:               auth,
		userUsecase:        userUsecase,
		userRoleRepository: userRoleRepository,
		menuRepository:     menuRepository,
	}
}

// ParseUserID parse user id for middleware
func (u *LoginUsecase) ParseUserID(c *gin.Context) (string, error) {
	rootID := u.conf.RootId
	if u.conf.AuthDisable {
		return rootID, nil
	}

	invalidToken := errors.Unauthorized(ErrInvalidTokenID, "Invalid access token")
	token := common.GetToken(c)
	if token == "" {
		return "", invalidToken
	}

	ctx := c.Request.Context()
	ctx = common.NewUserToken(ctx, token)

	userID, err := u.auth.ParseSubject(ctx, token)
	if err != nil {
		if errors.Is(err, jwtx.ErrInvalidToken) {
			return "", invalidToken
		}
		return "", err
	} else if userID == rootID {
		c.Request = c.Request.WithContext(common.NewIsRootUser(ctx))
		return userID, nil
	}

	userCacheVal, ok, err := u.cache.Get(ctx, CacheNSForUser, userID)
	if err != nil {
		return "", err
	} else if ok {
		userCache := common.ParseUserCache(userCacheVal)
		c.Request = c.Request.WithContext(common.NewUserCache(ctx, userCache))
		return userID, nil
	}

	// Check user status, if not activated, force to logout
	user, err := u.userRepository.Get(ctx, userID, UserQueryOptions{
		QueryOptions: common.QueryOptions{SelectFields: []string{"status"}},
	})
	if err != nil {
		return "", err
	} else if user == nil || user.Status != UserStatusActivated {
		return "", invalidToken
	}

	roleIDs, err := u.userUsecase.GetRoleIDs(ctx, userID)
	if err != nil {
		return "", err
	}

	userCache := common.UserCache{
		RoleIDs: roleIDs,
	}
	err = u.cache.Set(ctx, CacheNSForUser, userID, userCache.String())
	if err != nil {
		return "", err
	}

	c.Request = c.Request.WithContext(common.NewUserCache(ctx, userCache))
	return userID, nil
}

func (u *LoginUsecase) GetCaptcha(ctx context.Context) (CaptchaResult, error) {
	return CaptchaResult{
		CaptchaID: captcha.NewLen(captchaLength),
	}, nil
}

// ResponseCaptcha Response captcha image
func (u *LoginUsecase) ResponseCaptcha(ctx context.Context, w http.ResponseWriter, id string, reload bool) error {
	if reload && !captcha.Reload(id) {
		return errors.NotFound("", "Captcha id not found")
	}
	err := captcha.WriteImage(w, id, captchaWidth, captchaHeight)
	if err != nil {
		if errors.Is(err, captcha.ErrNotFound) {
			return errors.NotFound("", "Captcha id not found")
		}
		return err
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "image/png")
	return nil
}

func (u *LoginUsecase) genUserToken(ctx context.Context, userID string) (*LoginResult, error) {
	token, err := u.auth.GenerateToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	tokenBuf, err := token.EncodeToJSON()
	if err != nil {
		return nil, err
	}
	log.Context(ctx).Info("Generate user token", zap.Any("token", string(tokenBuf)))
	return &LoginResult{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
		ExpiresAt:   token.GetExpiresAt(),
	}, nil
}

func (u *LoginUsecase) Login(ctx context.Context, formItem *LoginForm) (*LoginResult, error) {
	// verify captcha
	if !captcha.VerifyString(formItem.CaptchaID, formItem.CaptchaCode) {
		return nil, errors.BadRequest(ErrInvalidCaptchaID, "Incorrect captcha")
	}
	ctx = log.NewTag(ctx, log.TagKeyLogin)
	// login by root
	if formItem.Username == u.conf.RootUsername {
		if formItem.Password != u.conf.RootPassword {
			return nil, errors.BadRequest(ErrInvalidUsernameOrPassword, "Incorrect username or password")
		}
		userID := u.conf.RootId
		ctx = log.NewUserID(ctx, userID)
		log.Context(ctx).Info("Login by root")
		return u.genUserToken(ctx, userID)
	}
	// get user info
	user, err := u.userRepository.GetByUsername(ctx, formItem.Username, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"id", "password", "status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.BadRequest(ErrInvalidUsernameOrPassword, "Incorrect username or password")
	} else if user.Status != UserStatusActivated {
		return nil, errors.BadRequest("", "User status is not activated, please contact the administrator")
	}

	// check password
	if err := hash.CompareHashAndPassword(user.Password, formItem.Password); err != nil {
		return nil, errors.BadRequest(ErrInvalidUsernameOrPassword, "Incorrect username or password")
	}

	userID := user.ID
	ctx = log.NewUserID(ctx, userID)

	// set user cache with role ids
	roleIDs, err := u.userUsecase.GetRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	userCache := common.UserCache{RoleIDs: roleIDs}
	err = u.cache.Set(ctx, CacheNSForUser, userID, userCache.String(),
		time.Duration(u.conf.UserCacheExp)*time.Hour)
	if err != nil {
		log.Context(ctx).Error("Failed to set cache", zap.Error(err))
	}
	log.Context(ctx).Info("Login success", zap.String("username", formItem.Username))
	// generate token
	return u.genUserToken(ctx, userID)
}

func (u *LoginUsecase) RefreshToken(ctx context.Context) (*LoginResult, error) {
	userID := common.FromUserID(ctx)
	user, err := u.userRepository.Get(ctx, userID, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.BadRequest("", "Incorrect user")
	} else if user.Status != UserStatusActivated {
		return nil, errors.BadRequest("", "User status is not activated, please contact the administrator")
	}
	return u.genUserToken(ctx, userID)
}

func (u *LoginUsecase) Logout(ctx context.Context) error {
	userToken := common.FromUserToken(ctx)
	if userToken == "" {
		return nil
	}

	ctx = log.NewTag(ctx, log.TagKeyLogout)
	if err := u.auth.DestroyToken(ctx, userToken); err != nil {
		return err
	}

	userID := common.FromUserID(ctx)
	err := u.cache.Delete(ctx, CacheNSForUser, userID)
	if err != nil {
		log.Context(ctx).Error("Failed to delete user cache", zap.Error(err))
	}
	log.Context(ctx).Info("Logout success")
	return nil
}

func (u *LoginUsecase) GetUserInfo(ctx context.Context) (*User, error) {
	if common.FromIsRootUser(ctx) {
		return &User{
			ID:       u.conf.RootId,
			Username: u.conf.RootUsername,
			Name:     u.conf.RootName,
			Status:   UserStatusActivated,
		}, nil
	}
	userID := common.FromUserID(ctx)
	user, err := u.userRepository.Get(ctx, userID, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	userRoleResult, err := u.userRoleRepository.Query(ctx, UserRoleQueryParam{
		UserID: userID,
	}, UserRoleQueryOptions{
		JoinRole: true,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

func (u *LoginUsecase) UpdatePassword(ctx context.Context, updateItem *UpdateLoginPassword) error {
	if common.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot change password")
	}

	userID := common.FromUserID(ctx)
	user, err := u.userRepository.Get(ctx, userID, UserQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"password"},
		},
	})
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// check old password
	if err := hash.CompareHashAndPassword(user.Password, updateItem.OldPassword); err != nil {
		return errors.BadRequest("", "Incorrect old password")
	}

	// update password
	newPassword, err := hash.GeneratePassword(updateItem.NewPassword)
	if err != nil {
		return err
	}
	return u.userRepository.UpdatePasswordByID(ctx, userID, newPassword)
}

func (u *LoginUsecase) QueryMenus(ctx context.Context) (Menus, error) {
	menuQueryParams := MenuQueryParam{
		Status: MenuStatusEnabled,
	}

	isRoot := common.FromIsRootUser(ctx)
	if !isRoot {
		menuQueryParams.UserID = common.FromUserID(ctx)
	}
	menuResult, err := u.menuRepository.Query(ctx, menuQueryParams, MenuQueryOptions{
		QueryOptions: common.QueryOptions{
			OrderFields: MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	} else if isRoot {
		return menuResult.Data.ToTree(), nil
	}

	// fill parent menus
	if parentIDs := menuResult.Data.SplitParentIDs(); len(parentIDs) > 0 {
		var missMenusIDs []string
		menuIDMapper := menuResult.Data.ToMap()
		for _, parentID := range parentIDs {
			if _, ok := menuIDMapper[parentID]; !ok {
				missMenusIDs = append(missMenusIDs, parentID)
			}
		}
		if len(missMenusIDs) > 0 {
			parentResult, err := u.menuRepository.Query(ctx, MenuQueryParam{
				InIDs: missMenusIDs,
			})
			if err != nil {
				return nil, err
			}
			menuResult.Data = append(menuResult.Data, parentResult.Data...)
			sort.Sort(menuResult.Data)
		}
	}
	return menuResult.Data.ToTree(), nil
}

func (u *LoginUsecase) UpdateUser(ctx context.Context, updateItem *UpdateCurrentUser) error {
	if common.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot update")
	}

	userID := common.FromUserID(ctx)
	user, err := u.userRepository.Get(ctx, userID)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	user.Name = updateItem.Name
	user.Phone = updateItem.Phone
	user.Email = updateItem.Email
	user.Remark = updateItem.Remark
	return u.userRepository.Update(ctx, user, "name", "phone", "email", "remark")
}
