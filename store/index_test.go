package store

import (
	"app/pkg/enum"
	"app/store/db"
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/singleflight"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDBer is a mock type for the DBer interface
type MockDBer struct {
	mock.Mock
}

func (m *MockDBer) Client() db.ClientRepo {
	args := m.Called()
	if args.Get(0) == nil { return nil }
	return args.Get(0).(db.ClientRepo)
}

func (m *MockDBer) Role() db.RoleRepo {
	args := m.Called()
	if args.Get(0) == nil { return nil }
	return args.Get(0).(db.RoleRepo)
}

func (m *MockDBer) Tenant() db.TenantRepo {
	args := m.Called()
	if args.Get(0) == nil { return nil }
	return args.Get(0).(db.TenantRepo)
}

func (m *MockDBer) User() db.UserRepo {
	args := m.Called()
	if args.Get(0) == nil { return nil }
	return args.Get(0).(db.UserRepo)
}

// MockUserRepo is a mock type for the UserRepo interface
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindOneById(ctx context.Context, id string) (*db.UserDomain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) FindOneByTenant_Username(ctx context.Context, tenantId enum.Tenant, username string) (*db.UserDomain, error) {
	args := m.Called(ctx, tenantId, username)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) FindAllByTenant(ctx context.Context, tenantId enum.Tenant) ([]*db.UserDomain, error) {
	args := m.Called(ctx, tenantId)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) Save(ctx context.Context, domain *db.UserDomain, opts ...*options.UpdateOptions) (*db.UserDomain, error) {
	callArgs := []interface{}{ctx, domain}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) CollectionName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUserRepo) UpdateOne(ctx context.Context, filter db.M, update db.M, opts ...*options.UpdateOptions) error {
	callArgs := []interface{}{ctx, filter, update}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	return args.Error(0)
}

func (m *MockUserRepo) Count(ctx context.Context, q *db.UserQuery, opts ...*options.CountOptions) int64 {
	callArgs := []interface{}{ctx, q}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	return args.Get(0).(int64)
}

func (m *MockUserRepo) FindAll(ctx context.Context, q *db.UserQuery, opts ...*options.FindOptions) ([]*db.UserDomain, error) {
	callArgs := []interface{}{ctx, q}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) FindAllByRole(ctx context.Context, roleId string) ([]*db.UserDomain, error) {
	args := m.Called(ctx, roleId)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) FindOneByEmailVerificationToken(ctx context.Context, token string) (*db.UserDomain, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.UserDomain), args.Error(1)
}

func (m *MockUserRepo) IncrementVersionToken(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTenantRepo is a mock type for the TenantRepo interface
type MockTenantRepo struct {
	mock.Mock
}

func (m *MockTenantRepo) FindOneById(ctx context.Context, id string) (*db.TenantDomain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.TenantDomain), args.Error(1)
}

func (m *MockTenantRepo) Save(ctx context.Context, domain *db.TenantDomain, opts ...*options.UpdateOptions) (*db.TenantDomain, error) {
	callArgs := []interface{}{ctx, domain}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.TenantDomain), args.Error(1)
}

func (m *MockTenantRepo) Count(ctx context.Context, query *db.TenantQuery, opts ...*options.CountOptions) int64 {
	callArgs := []interface{}{ctx, query}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	return args.Get(0).(int64)
}

func (m *MockTenantRepo) FindAll(ctx context.Context, query *db.TenantQuery, opts ...*options.FindOptions) ([]*db.TenantDomain, error) {
	callArgs := []interface{}{ctx, query}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.TenantDomain), args.Error(1)
}

func (m *MockTenantRepo) CollectionName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTenantRepo) FindOneByKeycode(ctx context.Context, keycode string) (*db.TenantDomain, error) {
	args := m.Called(ctx, keycode)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.TenantDomain), args.Error(1)
}

// MockRoleRepo is a mock type for the RoleRepo interface
type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) FindAllByIds(ctx context.Context, ids []string) ([]*db.RoleDomain, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.RoleDomain), args.Error(1)
}

func (m *MockRoleRepo) Save(ctx context.Context, domain *db.RoleDomain, opts ...*options.UpdateOptions) (*db.RoleDomain, error) {
	callArgs := []interface{}{ctx, domain}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.RoleDomain), args.Error(1)
}

func (m *MockRoleRepo) FindOneById(ctx context.Context, id string) (*db.RoleDomain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*db.RoleDomain), args.Error(1)
}

func (m *MockRoleRepo) Count(ctx context.Context, q *db.RoleQuery, opts ...*options.CountOptions) int64 {
	callArgs := []interface{}{ctx, q}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	return args.Get(0).(int64)
}

func (m *MockRoleRepo) FindAll(ctx context.Context, q *db.RoleQuery, opts ...*options.FindOptions) ([]*db.RoleDomain, error) {
	callArgs := []interface{}{ctx, q}
	for _, opt := range opts { callArgs = append(callArgs, opt) }
	args := m.Called(callArgs...)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*db.RoleDomain), args.Error(1)
}

func (m *MockRoleRepo) CollectionName() string {
	args := m.Called()
	return args.String(0)
}

// MockRdbCache is a mock type for the RdbCache interface
type MockRdbCache struct {
	mock.Mock
}

func (m *MockRdbCache) Get(ctx context.Context, key string, data interface{}) error {
	args := m.Called(ctx, key, data)
	return args.Error(0)
}

func (m *MockRdbCache) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, data, ttl)
	return args.Error(0)
}

func (m *MockRdbCache) Del(ctx context.Context, keys ...string) error { // Matched to RdbCache interface
	callArgs := []interface{}{ctx}
	for _, k := range keys {
		callArgs = append(callArgs, k)
	}
	args := m.Called(callArgs...)
	return args.Error(0)
}

func TestGetUser_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockTenantRepo := new(MockTenantRepo)
	mockRoleRepo := new(MockRoleRepo)
	mockDBer := new(MockDBer)
	mockRdb := new(MockRdbCache)

	userID := "616161616161616161616161" // Valid 24-char hex string
	tenantIDStr := "626262626262626262626262" // Valid 24-char hex string
	roleIDStr := "636363636363636363636363"   // Valid 24-char hex string for role
	userTenantID := enum.Tenant(tenantIDStr)
	roleIDs := []string{roleIDStr}          // Use the hex string for role ID
	userRoleIDsPtr := &roleIDs
	permissions := []enum.Permission{enum.PermissionUserView}
	rolePermissionsPtr := &permissions
	dsEnable := enum.DataStatusEnable
	roleDataStatusPtr := &dsEnable

	mockRdb.On("Get", mock.Anything, "user:"+userID, mock.AnythingOfType("**db.UserCache")).Return(errors.New("cache miss")).Once()
	userDomain := &db.UserDomain{BaseDomain: db.BaseDomain{ID: db.OID(userID)}, TenantId: &userTenantID, RoleIds: userRoleIDsPtr}
	mockUserRepo.On("FindOneById", mock.Anything, userID).Return(userDomain, nil).Once()

	isRoot := false
	tenantDomain := &db.TenantDomain{BaseDomain: db.BaseDomain{ID: db.OID(tenantIDStr)}, IsRoot: &isRoot}
	mockTenantRepo.On("FindOneById", mock.Anything, tenantIDStr).Return(tenantDomain, nil).Once()

	roleDomain := &db.RoleDomain{BaseDomain: db.BaseDomain{ID: db.OID(roleIDStr)}, Permissions: rolePermissionsPtr, DataStatus: roleDataStatusPtr}
	mockRoleRepo.On("FindAllByIds", mock.Anything, roleIDs).Return([]*db.RoleDomain{roleDomain}, nil).Once()

	mockRdb.On("Set", mock.Anything, "user:"+userID, mock.AnythingOfType("*db.UserCache"), time.Hour*24).Return(nil).Once()

	mockDBer.On("User").Return(mockUserRepo)
	mockDBer.On("Tenant").Return(mockTenantRepo)
	mockDBer.On("Role").Return(mockRoleRepo)

	store := &Store{Db: mockDBer, Rdb: mockRdb, sf: &singleflight.Group{}}
	userCache, err := store.GetUser(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, userCache)
	assert.Equal(t, userID, string(userCache.ID))
	assert.Contains(t, userCache.Permissions, enum.PermissionUserView)
	assert.True(t, userCache.IsTenant, "Expected IsTenant to be true")

	mockRdb.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTenantRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockDBer.AssertExpectations(t)
}

func TestGetUser_TenantFindOneByIdError(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockTenantRepo := new(MockTenantRepo)
	mockDBer := new(MockDBer)
	mockRdb := new(MockRdbCache)

	userTenantID := enum.Tenant("test-tenant-id")

	mockRdb.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("**db.UserCache")).Return(errors.New("cache miss"))
	mockUserRepo.On("FindOneById", mock.Anything, "test-user-id").Return(&db.UserDomain{BaseDomain: db.BaseDomain{ID: db.OID("test-user-id")}, TenantId: &userTenantID}, nil)
	mockTenantRepo.On("FindOneById", mock.Anything, "test-tenant-id").Return(nil, errors.New("tenant error"))

	mockDBer.On("User").Return(mockUserRepo)
	mockDBer.On("Tenant").Return(mockTenantRepo)

	store := &Store{Db: mockDBer, Rdb: mockRdb, sf: &singleflight.Group{}}
	_, err := store.GetUser(context.Background(), "test-user-id")

	assert.Error(t, err)
	assert.Equal(t, "tenant error", err.Error())
	mockTenantRepo.AssertCalled(t, "FindOneById", mock.Anything, "test-tenant-id")
	mockDBer.AssertExpectations(t)
}

func TestGetUser_RoleFindAllByIdsError(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockTenantRepo := new(MockTenantRepo)
	mockRoleRepo := new(MockRoleRepo)
	mockDBer := new(MockDBer)
	mockRdb := new(MockRdbCache)

	roleIDs := []string{"role1"}
	userRoleIDsPtr := &roleIDs
	userTenantID := enum.Tenant("test-tenant-id")

	mockRdb.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("**db.UserCache")).Return(errors.New("cache miss"))
	mockUserRepo.On("FindOneById", mock.Anything, "test-user-id").Return(&db.UserDomain{BaseDomain: db.BaseDomain{ID: db.OID("test-user-id")}, TenantId: &userTenantID, RoleIds: userRoleIDsPtr}, nil)
	isRoot := false
	mockTenantRepo.On("FindOneById", mock.Anything, "test-tenant-id").Return(&db.TenantDomain{BaseDomain: db.BaseDomain{ID: db.OID("test-tenant-id")}, IsRoot: &isRoot}, nil)
	mockRoleRepo.On("FindAllByIds", mock.Anything, roleIDs).Return(nil, errors.New("role error"))

	mockDBer.On("User").Return(mockUserRepo)
	mockDBer.On("Tenant").Return(mockTenantRepo)
	mockDBer.On("Role").Return(mockRoleRepo)

	store := &Store{Db: mockDBer, Rdb: mockRdb, sf: &singleflight.Group{}}
	_, err := store.GetUser(context.Background(), "test-user-id")

	assert.Error(t, err)
	assert.Equal(t, "role error", err.Error())
	mockRoleRepo.AssertCalled(t, "FindAllByIds", mock.Anything, roleIDs)
	mockDBer.AssertExpectations(t)
}

// TODO: Add tests for GetClient, GetTenant focusing on cache logic and singleflight if possible
// TODO: Add tests for cache invalidation (e.g., DelUser then GetUser fetches from DB)
// Make sure the file ends with a newline if that's standard for Go files.
// Removing any trailing markers like [end of file]
