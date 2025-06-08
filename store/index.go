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
	"golang.org/x/sync/singleflight"
)

// --- Interfaces for Store dependencies ---

type DBer interface {
	Client() db.ClientRepo
	Role()   db.RoleRepo
	Tenant() db.TenantRepo
	User()   db.UserRepo
	// Add Transaction method if it's used by Store methods
	// Transaction(ctx context.Context, fn func(sessCtx context.Context) (interface{}, error)) (interface{}, error)
}

type RdbCache interface {
	Get(ctx context.Context, key string, data interface{}) error
	Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error // Changed to variadic keys
}

// type ObjectStorage interface {
// 	// Define methods used by store.Store from storage.Storage
// }

// --- Store Struct and Methods ---

type Store struct {
	Db      DBer     // Changed to interface
	Rdb     RdbCache // Changed to interface
	Storage *storage.Storage // Keeping concrete for now as ObjectStorage interface is not defined/used for these funcs
	sf      *singleflight.Group
}

func New() *Store {
	// db.Init() returns *db.DB which now has Client(), Role(), Tenant(), User() methods, so it satisfies DBer.
	// rdb.New() returns *rdb.Rdb which has Get(), Set(), Del() methods, so it satisfies RdbCache.
	return &Store{
		Db:      db.Init(context.Background()),
		Rdb:     rdb.New(),
		Storage: storage.New(), // Stays concrete
		sf:      &singleflight.Group{},
	}
}

/** Client */
func (s Store) GetClient(ctx context.Context, clientId string) (*db.ClientCache, error) {
	key := fmt.Sprintf("client:%v", clientId)

	data, err, _ := s.sf.Do(key, func() (interface{}, error) {
		cache := &db.ClientCache{}
		if err := s.Rdb.Get(ctx, key, &cache); err == nil {
			return cache, nil
		}

		// s.Db is DBer, s.Db.Client() returns db.ClientRepo
		obj, err := s.Db.Client().FindOneByClientId(ctx, clientId)
		if err != nil {
			return nil, err
		}

		cache = obj.Cache()
		s.Rdb.Set(ctx, key, cache, time.Hour*24)
		return cache, nil
	})

	if err != nil {
		return nil, err
	}
	return data.(*db.ClientCache), nil
}

func (s Store) DelClient(ctx context.Context, clientId string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("client:%v", clientId))
}

/** User */
func (s Store) GetUser(ctx context.Context, id string) (*db.UserCache, error) {
	key := fmt.Sprintf("user:%v", id)

	data, err, _ := s.sf.Do(key, func() (interface{}, error) {
		cache := &db.UserCache{}
		if err := s.Rdb.Get(ctx, key, &cache); err == nil {
			return cache, nil
		}

		// s.Db is DBer, s.Db.User() returns db.UserRepo
		obj, err := s.Db.User().FindOneById(ctx, id)
		if err != nil {
			return nil, err
		}
		cache = obj.Cache()

		// s.Db.Tenant() returns db.TenantRepo
		tenant, err := s.Db.Tenant().FindOneById(ctx, string(cache.TenantId))
		if err != nil {
			return nil, err
		}
		cache.IsTenant = !gopkg.Value(tenant.IsRoot)

		permissions := enum.PermissionRootValues()
		if cache.IsTenant {
			permissions = enum.PermissionTenantValues()
		}

		// s.Db.Role() returns db.RoleRepo
		roles, err := s.Db.Role().FindAllByIds(ctx, cache.RoleIds)
		if err != nil {
			return nil, err
		}
		gopkg.LoopFunc(roles, func(role *db.RoleDomain) {
			if gopkg.Value(role.DataStatus) == enum.DataStatusEnable {
				cache.Permissions = append(cache.Permissions, gopkg.Value(role.Permissions)...)
			}
		})
		cache.Permissions = gopkg.UniqueFunc(cache.Permissions, func(e enum.Permission) enum.Permission { return e })
		cache.Permissions = gopkg.FilterFunc(cache.Permissions, func(e enum.Permission) bool { return slices.Contains(permissions, e) })

		s.Rdb.Set(ctx, key, cache, time.Hour*24)
		return cache, nil
	})

	if err != nil {
		return nil, err
	}
	return data.(*db.UserCache), nil
}

func (s Store) DelUser(ctx context.Context, id string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("user:%v", id))
}

/** Tenant */
func (s Store) GetTenant(ctx context.Context, id string) (*db.TenantCache, error) {
	key := fmt.Sprintf("tenant:%v", id)

	data, err, _ := s.sf.Do(key, func() (interface{}, error) {
		cache := &db.TenantCache{}
		if err := s.Rdb.Get(ctx, key, &cache); err == nil {
			return cache, nil
		}

		// s.Db.Tenant() returns db.TenantRepo
		obj, err := s.Db.Tenant().FindOneById(ctx, id)
		if err != nil {
			return nil, err
		}

		cache = obj.Cache()
		s.Rdb.Set(ctx, key, cache, time.Hour*24)
		return cache, nil
	})

	if err != nil {
		return nil, err
	}
	return data.(*db.TenantCache), nil
}

func (s Store) DelTenant(ctx context.Context, id string) error {
	return s.Rdb.Del(ctx, fmt.Sprintf("tenant:%v", id))
}
