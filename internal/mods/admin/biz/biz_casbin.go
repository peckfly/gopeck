package biz

import (
	"bytes"
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CasbinUsecase struct {
	conf                   *conf.CasbinConf
	cache                  cachex.Cache
	menuRepository         MenuRepository
	roleRepository         RoleRepository
	menuResourceRepository MenuResourceRepository

	Enforcer *atomic.Value `wire:"-"`
	ticker   *time.Ticker  `wire:"-"`
}

// NewCasbinUsecase initializes a new CasbinUsecase with the given parameters.
//
// cache: cachex.Cache
// menuRepository: MenuRepository
// roleRepository: RoleRepository
// menuResourceRepository: MenuResourceRepository
// conf: *conf.ServerConf
// Returns *CasbinUsecase
func NewCasbinUsecase(
	cache cachex.Cache,
	menuRepository MenuRepository,
	roleRepository RoleRepository,
	menuResourceRepository MenuResourceRepository,
	conf *conf.ServerConf,
) *CasbinUsecase {
	return &CasbinUsecase{
		cache:                  cache,
		conf:                   &conf.Casbin,
		menuRepository:         menuRepository,
		roleRepository:         roleRepository,
		menuResourceRepository: menuResourceRepository,
	}
}

func (a *CasbinUsecase) GetEnforcer() *casbin.Enforcer {
	if v := a.Enforcer.Load(); v != nil {
		return v.(*casbin.Enforcer)
	}
	return nil
}

type policyQueueItem struct {
	RoleID    string
	Resources MenuResources
}

func (a *CasbinUsecase) Load(ctx context.Context) error {
	if a.conf.Disable {
		return nil
	}

	a.Enforcer = new(atomic.Value)
	if err := a.load(ctx); err != nil {
		return err
	}

	go a.autoLoad(ctx)
	return nil
}

func (a *CasbinUsecase) load(ctx context.Context) error {
	start := time.Now()
	roleResult, err := a.roleRepository.Query(ctx, RoleQueryParam{
		Status: RoleStatusEnabled,
	}, RoleQueryOptions{
		QueryOptions: common.QueryOptions{SelectFields: []string{"id"}},
	})
	if err != nil {
		return err
	} else if len(roleResult.Data) == 0 {
		return nil
	}

	var resCount int32
	queue := make(chan *policyQueueItem, len(roleResult.Data))
	threadNum := a.conf.LoadThread
	lock := new(sync.Mutex)
	buf := new(bytes.Buffer)

	wg := new(sync.WaitGroup)
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go func() {
			defer wg.Done()
			buf0 := new(bytes.Buffer)
			for item := range queue {
				for _, res := range item.Resources {
					_, _ = buf0.WriteString(fmt.Sprintf("p, %s, %s, %s \n", item.RoleID, res.Path, res.Method))
				}
			}
			lock.Lock()
			_, _ = buf.Write(buf0.Bytes())
			lock.Unlock()
		}()
	}

	for _, item := range roleResult.Data {
		resources, err := a.queryRoleResources(ctx, item.ID)
		if err != nil {
			log.Context(ctx).Error("Failed to query role resources", zap.Error(err))
			continue
		}
		atomic.AddInt32(&resCount, int32(len(resources)))
		queue <- &policyQueueItem{
			RoleID:    item.ID,
			Resources: resources,
		}
	}
	close(queue)
	wg.Wait()

	if buf.Len() > 0 {
		policyFile := filepath.Join(a.conf.WorkDir, a.conf.GenPolicyFile)
		_ = os.Rename(policyFile, policyFile+".bak")
		_ = os.MkdirAll(filepath.Dir(policyFile), 0755)
		if err := os.WriteFile(policyFile, buf.Bytes(), 0666); err != nil {
			log.Context(ctx).Error("Failed to write policy file", zap.Error(err))
			return err
		}
		// set readonly
		_ = os.Chmod(policyFile, 0444)

		modelFile := filepath.Join(a.conf.WorkDir, a.conf.ModelFile)
		e, err := casbin.NewEnforcer(modelFile, policyFile)
		if err != nil {
			log.Context(ctx).Error("Failed to create casbin enforcer", zap.Error(err))
			return err
		}
		a.Enforcer.Store(e)
	}
	log.Context(ctx).Info("Casbin load policy",
		zap.Duration("cost", time.Since(start)),
		zap.Int("roles", len(roleResult.Data)),
		zap.Int32("resources", resCount),
		zap.Int("bytes", buf.Len()),
	)
	return nil
}

func (a *CasbinUsecase) queryRoleResources(ctx context.Context, roleID string) (MenuResources, error) {
	menuResult, err := a.menuRepository.Query(ctx, MenuQueryParam{
		RoleID: roleID,
		Status: MenuStatusEnabled,
	}, MenuQueryOptions{
		QueryOptions: common.QueryOptions{
			SelectFields: []string{"id", "parent_id", "parent_path"},
		},
	})
	if err != nil {
		return nil, err
	} else if len(menuResult.Data) == 0 {
		return nil, nil
	}

	menuIDs := make([]string, 0, len(menuResult.Data))
	menuIDMapper := make(map[string]struct{})
	for _, item := range menuResult.Data {
		if _, ok := menuIDMapper[item.ID]; ok {
			continue
		}
		menuIDs = append(menuIDs, item.ID)
		menuIDMapper[item.ID] = struct{}{}
		if pp := item.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, common.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := menuIDMapper[pid]; ok {
					continue
				}
				menuIDs = append(menuIDs, pid)
				menuIDMapper[pid] = struct{}{}
			}
		}
	}

	menuResourceResult, err := a.menuResourceRepository.Query(ctx, MenuResourceQueryParam{
		MenuIDs: menuIDs,
	})
	if err != nil {
		return nil, err
	}

	return menuResourceResult.Data, nil
}

func (a *CasbinUsecase) autoLoad(ctx context.Context) {
	var lastUpdated int64
	a.ticker = time.NewTicker(time.Duration(a.conf.AutoLoadInterval) * time.Second)
	for range a.ticker.C {
		val, ok, err := a.cache.Get(ctx, CacheNSForRole, CacheKeyForSyncToCasbin)
		if err != nil {
			log.Context(ctx).Error("Failed to get cache", zap.Error(err), zap.String("key", CacheKeyForSyncToCasbin))
			continue
		} else if !ok {
			continue
		}

		updated, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Context(ctx).Error("Failed to parse cache value", zap.Error(err), zap.String("val", val))
			continue
		}

		if lastUpdated < updated {
			if err := a.load(ctx); err != nil {
				log.Context(ctx).Error("Failed to load casbin policy", zap.Error(err))
			} else {
				lastUpdated = updated
			}
		}
	}
}

func (a *CasbinUsecase) Release(ctx context.Context) error {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	return nil
}
