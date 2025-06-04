package store

import (
	"app/pkg/enum"
	"app/store/db"
	"app/store/rdb"
	"app/store/storage"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/nhnghia272/gopkg"
)

type Store struct {
	Db      *db.DB
	Rdb     *rdb.Rdb
	Storage *storage.Storage
}

func New() *Store {
	return &Store{db.Init(context.Background()), rdb.New(), storage.New()}
}

/** Client */
func (s Store) GetClient(ctx context.Context, clientId string) (*db.ClientCache, error) {
	key := fmt.Sprintf("client:%v", clientId)
	cache := &db.ClientCache{}

	if err := s.Rdb.Get(ctx, key, &cache); err == nil {
		return cache, nil
	}
	obj, err := s.Db.Client.FindOneByClientId(ctx, clientId)
	if err == nil {
		cache = obj.Cache()
		s.Rdb.Set(ctx, key, cache, time.Hour*24)
	}

	return cache, err
}

func (s Store) DelClient(ctx context.Context, clientId string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("client:%v", clientId))
}

/** User */
func (s Store) GetUser(ctx context.Context, id string) (*db.UserCache, error) {
	key := fmt.Sprintf("user:%v", id)
	cache := &db.UserCache{}

	if err := s.Rdb.Get(ctx, key, &cache); err == nil {
		return cache, nil
	}
	obj, err := s.Db.User.FindOneById(ctx, id)
	if err == nil {
		cache = obj.Cache()

		tenant, _ := s.Db.Tenant.FindOneById(ctx, string(cache.TenantId))
		cache.IsTenant = !gopkg.Value(tenant.IsRoot)

		permissions := enum.PermissionRootValues()
		if cache.IsTenant {
			permissions = enum.PermissionTenantValues()
		}

		roles, _ := s.Db.Role.FindAllByIds(ctx, cache.RoleIds)
		gopkg.LoopFunc(roles, func(role *db.RoleDomain) {
			if gopkg.Value(role.DataStatus) == enum.DataStatusEnable {
				cache.Permissions = append(cache.Permissions, gopkg.Value(role.Permissions)...)
			}
		})
		cache.Permissions = gopkg.UniqueFunc(cache.Permissions, func(e enum.Permission) enum.Permission { return e })
		cache.Permissions = gopkg.FilterFunc(cache.Permissions, func(e enum.Permission) bool { return slices.Contains(permissions, e) })

		s.Rdb.Set(ctx, key, cache, time.Hour*24)
	}

	return cache, err
}

func (s Store) DelUser(ctx context.Context, id string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("user:%v", id))
}

/** Tenant */
func (s Store) GetTenant(ctx context.Context, id string) (*db.TenantCache, error) {
	key := fmt.Sprintf("tenant:%v", id)
	cache := &db.TenantCache{}

	if err := s.Rdb.Get(ctx, key, &cache); err == nil {
		return cache, nil
	}
	obj, err := s.Db.Tenant.FindOneById(ctx, id)
	if err == nil {
		cache = obj.Cache()
		s.Rdb.Set(ctx, key, cache, time.Hour*24)
	}

	return cache, err
}

func (s Store) DelTenant(ctx context.Context, id string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("tenant:%v", id))
}
