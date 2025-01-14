// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matrixorigin/matrixone/pkg/bootstrap/versions"
	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/common/runtime"
	"github.com/matrixorigin/matrixone/pkg/common/stopper"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	mock_frontend "github.com/matrixorigin/matrixone/pkg/frontend/test"
	"github.com/matrixorigin/matrixone/pkg/pb/lock"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/txn/clock"
	"github.com/matrixorigin/matrixone/pkg/txn/rpc"
	"github.com/matrixorigin/matrixone/pkg/util/executor"
)

var _ client.TxnOperator = new(testTxnOperator)

type testTxnOperator struct {
}

func (tTxnOp *testTxnOperator) GetOverview() client.TxnOverview {
	return client.TxnOverview{}
}

func (tTxnOp *testTxnOperator) CloneSnapshotOp(snapshot timestamp.Timestamp) client.TxnOperator {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) IsSnapOp() bool {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Txn() txn.TxnMeta {
	return txn.TxnMeta{}
}

func (tTxnOp *testTxnOperator) TxnOptions() txn.TxnOptions {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) TxnRef() *txn.TxnMeta {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Snapshot() (txn.CNTxnSnapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) UpdateSnapshot(ctx context.Context, ts timestamp.Timestamp) error {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) SnapshotTS() timestamp.Timestamp {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) CreateTS() timestamp.Timestamp {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Status() txn.TxnStatus {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) ApplySnapshot(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Read(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Write(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) WriteAndCommit(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Commit(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Rollback(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) AddLockTable(locktable lock.LockTable) error {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) HasLockTable(table uint64) bool {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) AddWaitLock(tableID uint64, rows [][]byte, opt lock.LockOptions) uint64 {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) RemoveWaitLock(key uint64) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) LockTableCount() int32 {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) LockSkipped(tableID uint64, mode lock.LockMode) bool {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) GetWaitActiveCost() time.Duration {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) AddWorkspace(workspace client.Workspace) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) GetWorkspace() client.Workspace {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) AppendEventCallback(event client.EventType, callbacks ...func(client.TxnEvent)) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) Debug(ctx context.Context, ops []txn.TxnRequest) (*rpc.SendResult, error) {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) NextSequence() uint64 {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) EnterRunSql() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) ExitRunSql() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) EnterIncrStmt() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) ExitIncrStmt() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) EnterRollbackStmt() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) ExitRollbackStmt() {
	//TODO implement me
	panic("implement me")
}

func (tTxnOp *testTxnOperator) SetFootPrints(id int, enter bool) {
	//TODO implement me
	panic("implement me")
}

func TestBootstrapAlreadyBootstrapped(t *testing.T) {
	sid := ""
	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			n := 0
			exec := executor.NewMemExecutor(func(sql string) (executor.Result, error) {
				if sql == "show databases" {
					n++
					memRes := executor.NewMemResult(
						[]types.Type{types.New(types.T_varchar, 2, 0)},
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{bootstrappedCheckerDB})
					return memRes.GetResult(), nil
				}
				return executor.Result{}, nil
			})

			b := NewService(
				sid,
				&memLocker{},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				exec,
			)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			require.NoError(t, b.Bootstrap(ctx))
			assert.Equal(t, 1, n)
		},
	)
}

func TestBootstrapWithWait(t *testing.T) {
	sid := ""
	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			var n atomic.Uint32
			exec := executor.NewMemExecutor(func(sql string) (executor.Result, error) {
				if sql == "show databases" && n.Load() == 1 {
					memRes := executor.NewMemResult(
						[]types.Type{types.New(types.T_varchar, 2, 0)},
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{bootstrappedCheckerDB})
					return memRes.GetResult(), nil
				}
				n.Add(1)
				return executor.Result{}, nil
			})

			b := NewService(
				sid,
				&memLocker{ids: map[string]uint64{
					bootstrapKey: 1,
				}},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				exec,
			)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			require.NoError(t, b.Bootstrap(ctx))
			assert.True(t, n.Load() > 0)
		},
	)
}

type memLocker struct {
	sync.Mutex
	ids map[string]uint64
}

func (l *memLocker) Get(
	ctx context.Context,
	key string) (bool, error) {
	l.Lock()
	defer l.Unlock()
	if l.ids == nil {
		l.ids = make(map[string]uint64)
	}

	l.ids[key]++
	return l.ids[key] == 1, nil
}

// tolerance test
func TestDoCheckUpgrade(t *testing.T) {
	sid := ""
	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			defer func() {
				if r := recover(); r != nil {
					assert.Equal(t, "cannot upgrade to version 2.0.0, because version 1.2.3 is in upgrading", r)
				} else {
					t.Errorf("Expected panic but did not occur")
				}
			}()

			txnOp := &testTxnOperator{}
			exec := executor.NewMemExecutor2(func(sql string) (executor.Result, error) {
				if strings.HasPrefix(sql, "SELECT reldatabase, relname, account_id FROM mo_catalog.mo_tables") {
					memRes := executor.NewMemResult(
						[]types.Type{types.New(types.T_varchar, 2, 0)},
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{bootstrappedCheckerDB})
					return memRes.GetResult(), nil
				}

				if strings.EqualFold(sql, "select version, version_offset, state from mo_version order by create_at desc limit 1") {
					typs := []types.Type{
						types.New(types.T_varchar, 50, 0),
						types.New(types.T_uint32, 32, 0),
						types.New(types.T_int32, 32, 0),
					}

					memRes := executor.NewMemResult(
						typs,
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{"1.2.3"})
					executor.AppendFixedRows(memRes, 1, []uint32{10})
					executor.AppendFixedRows(memRes, 2, []int32{0})
					return memRes.GetResult(), nil
				}
				return executor.Result{}, nil
			}, txnOp)

			b := newServiceForTest(
				sid,
				&memLocker{},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				exec,
				func(s *service) {
					h1 := newTestVersionHandler("1.2.0", "1.1.0", versions.Yes, versions.No, 10)
					h2 := newTestVersionHandler("2.0.0", "1.2.0", versions.Yes, versions.No, 2)
					s.handles = append(s.handles, h1)
					s.handles = append(s.handles, h2)
				},
			)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			b.doCheckUpgrade(ctx)
		},
	)

	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			defer func() {
				if r := recover(); r != nil {
					assert.Equal(t, "cannot upgrade to version 2.0.0 with versionOffset[2], because version 2.0.0 with versionOffset[1] is in upgrading", r)
				} else {
					t.Errorf("Expected panic but did not occur")
				}
			}()

			txnOp := &testTxnOperator{}
			exec := executor.NewMemExecutor2(func(sql string) (executor.Result, error) {
				if strings.HasPrefix(sql, "SELECT reldatabase, relname, account_id FROM mo_catalog.mo_tables") {
					memRes := executor.NewMemResult(
						[]types.Type{types.New(types.T_varchar, 2, 0)},
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{bootstrappedCheckerDB})
					return memRes.GetResult(), nil
				}

				if strings.EqualFold(sql, "select version, version_offset, state from mo_version order by create_at desc limit 1") {
					typs := []types.Type{
						types.New(types.T_varchar, 50, 0),
						types.New(types.T_uint32, 32, 0),
						types.New(types.T_int32, 32, 0),
					}

					memRes := executor.NewMemResult(
						typs,
						mpool.MustNewZero())
					memRes.NewBatch()
					executor.AppendStringRows(memRes, 0, []string{"2.0.0"})
					executor.AppendFixedRows(memRes, 1, []uint32{1})
					executor.AppendFixedRows(memRes, 2, []int32{0})
					return memRes.GetResult(), nil
				}
				return executor.Result{}, nil
			}, txnOp)

			b := newServiceForTest(
				sid,
				&memLocker{},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				exec,
				func(s *service) {
					h1 := newTestVersionHandler("1.2.0", "1.1.0", versions.Yes, versions.No, 10)
					h2 := newTestVersionHandler("2.0.0", "1.2.0", versions.Yes, versions.No, 2)
					s.handles = append(s.handles, h1)
					s.handles = append(s.handles, h2)
				},
			)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			b.doCheckUpgrade(ctx)
		},
	)
}

func TestDoUpgrade(t *testing.T) {
	sid := ""
	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Expected no panic")
				}
			}()

			sqlExecutor := executor.NewMemExecutor(func(sql string) (executor.Result, error) {
				return executor.Result{}, nil
			})

			b := newServiceForTest(
				sid,
				&memLocker{},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				sqlExecutor,
				func(s *service) {
					h1 := newTestVersionHandler("1.2.0", "1.1.0", versions.Yes, versions.No, 10)
					h2 := newTestVersionHandler("2.0.0", "1.2.0", versions.Yes, versions.No, 2)
					s.handles = append(s.handles, h1)
					s.handles = append(s.handles, h2)
				},
			)

			txnOperator := mock_frontend.NewMockTxnOperator(gomock.NewController(t))
			txnOperator.EXPECT().TxnOptions().Return(txn.TxnOptions{CN: sid}).AnyTimes()

			txnExecutor := executor.NewMemTxnExecutor(func(sql string) (executor.Result, error) {
				return executor.Result{}, nil
			}, txnOperator)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			versionUpg := versions.VersionUpgrade{
				FromVersion:        "1.2.0",
				ToVersion:          "2.0.0",
				FinalVersion:       "2.0.0",
				FinalVersionOffset: 10,
				State:              versions.StateCreated,
				UpgradeOrder:       1,
				UpgradeCluster:     versions.Yes,
				UpgradeTenant:      versions.No,
			}

			_, err := b.doUpgrade(ctx, versionUpg, txnExecutor)
			assert.NoError(t, err)
		},
	)
}

// tolerance test
func TestUpgradeTenant(t *testing.T) {
	sid := ""
	runtime.RunTest(
		sid,
		func(rt runtime.Runtime) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Expected no panic")
				}
			}()

			txnOp := &testTxnOperator{}
			sqlExecutor := executor.NewMemExecutor2(func(sql string) (executor.Result, error) {
				target := fmt.Sprintf(`select version, version_offset, state from %s order by create_at desc limit 1`, catalog.MOVersionTable)
				if sql == target {
					return executor.Result{}, moerr.NewInternalErrorNoCtx("test err")
				}
				return executor.Result{}, nil
			}, txnOp)

			b := newServiceForTest(
				sid,
				&memLocker{},
				clock.NewHLCClock(func() int64 { return 0 }, 0),
				nil,
				sqlExecutor,
				func(s *service) {
					h1 := newTestVersionHandler("1.2.0", "1.1.0", versions.Yes, versions.No, 10)
					h2 := newTestVersionHandler("2.0.0", "1.2.0", versions.Yes, versions.No, 2)
					s.handles = append(s.handles, h1)
					s.handles = append(s.handles, h2)
				},
			)

			txnOperator := mock_frontend.NewMockTxnOperator(gomock.NewController(t))
			txnOperator.EXPECT().TxnOptions().Return(txn.TxnOptions{CN: sid}).AnyTimes()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			_, err := b.UpgradeTenant(ctx, "sys", 0, true)
			assert.Error(t, err)
		},
	)
}

func newServiceForTest(
	sid string,
	lock Locker,
	clock clock.Clock,
	client client.TxnClient,
	exec executor.SQLExecutor,
	initUpgrade func(s *service),
	opts ...Option,
) *service {
	s := &service{
		sid:     sid,
		clock:   clock,
		exec:    exec,
		lock:    lock,
		client:  client,
		logger:  getLogger(sid).Named("upgrade-framework"),
		stopper: stopper.NewStopper("upgrade", stopper.WithLogger(getLogger(sid).RawLogger())),
	}
	s.mu.tenants = make(map[int32]bool)
	initUpgrade(s)
	//s.handles = append(s.handles, v2_0_0.Handler)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func newTestVersionHandler(
	version, minVersion string,
	upgradeCluster, upgradeTenant int32, versionOffset uint32) *testVersionHandle {
	return &testVersionHandle{
		metadata: versions.Version{
			Version:           version,
			MinUpgradeVersion: minVersion,
			UpgradeCluster:    upgradeCluster,
			UpgradeTenant:     upgradeTenant,
			VersionOffset:     versionOffset,
		},
	}
}

type testVersionHandle struct {
	metadata                 versions.Version
	callHandleClusterUpgrade atomic.Uint64
	callHandleTenantUpgrade  atomic.Uint64
}

func (h *testVersionHandle) Metadata() versions.Version {
	return h.metadata
}
func (h *testVersionHandle) Prepare(ctx context.Context, txn executor.TxnExecutor, final bool) error {
	return nil
}
func (h *testVersionHandle) HandleClusterUpgrade(ctx context.Context, txn executor.TxnExecutor) error {
	h.callHandleClusterUpgrade.Add(1)
	return nil
}
func (h *testVersionHandle) HandleTenantUpgrade(ctx context.Context, tenantID int32, txn executor.TxnExecutor) error {
	h.callHandleTenantUpgrade.Add(1)
	return nil
}

func (h *testVersionHandle) HandleCreateFrameworkDeps(txn executor.TxnExecutor) error {
	return nil
}
