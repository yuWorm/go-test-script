package blueprint

import (
	"context"
	"sync"
	"time"
)

// ExecutionContext 表示蓝图执行的上下文
type ExecutionContext struct {
	ctx          context.Context
	cancel       context.CancelFunc
	blueprint    *Blueprint
	variables    map[string]interface{} // 运行时变量
	mu           sync.RWMutex
	startTime    time.Time
	errors       []error
	nodeErrors   map[string][]error // 每个节点的错误
	errorsMu     sync.Mutex
	asyncManager *AsyncTaskManager // 异步任务管理器

	// 返回机制（类似 UE5 Return 节点）
	returned    bool                   // 是否已返回
	returnValue map[string]interface{} // 返回值
	returnMu    sync.RWMutex

	// 优化：快速变量访问（单线程模式下跳过锁）
	fastMode bool // 是否启用快速模式（单线程执行）
}

// EnableFastMode 启用快速模式（单线程执行，跳过锁）
func (ec *ExecutionContext) EnableFastMode() {
	ec.fastMode = true
}

// GetVariableFast 快速获取变量（快速模式下跳过锁）
func (ec *ExecutionContext) GetVariableFast(name string) (interface{}, bool) {
	if ec.fastMode {
		val, ok := ec.variables[name]
		return val, ok
	}
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	val, ok := ec.variables[name]
	return val, ok
}

// SetVariableFast 快速设置变量（快速模式下跳过锁）
func (ec *ExecutionContext) SetVariableFast(name string, value interface{}) {
	if ec.fastMode {
		ec.variables[name] = value
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.variables[name] = value
}

// NewExecutionContext 创建一个新的执行上下文
func NewExecutionContext(bp *Blueprint) *ExecutionContext {
	ctx, cancel := context.WithCancel(context.Background())

	execCtx := &ExecutionContext{
		ctx:          ctx,
		cancel:       cancel,
		blueprint:    bp,
		variables:    make(map[string]interface{}),
		startTime:    time.Now(),
		errors:       make([]error, 0),
		nodeErrors:   make(map[string][]error),
		asyncManager: NewAsyncTaskManager(),
	}

	// 复制蓝图全局变量到执行上下文
	if bp.Variables != nil {
		for k, v := range bp.Variables {
			execCtx.variables[k] = v
		}
	}

	return execCtx
}

// NewExecutionContextWithTimeout 创建一个带超时的执行上下文
func NewExecutionContextWithTimeout(bp *Blueprint, timeout time.Duration) *ExecutionContext {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	execCtx := &ExecutionContext{
		ctx:          ctx,
		cancel:       cancel,
		blueprint:    bp,
		variables:    make(map[string]interface{}),
		startTime:    time.Now(),
		errors:       make([]error, 0),
		nodeErrors:   make(map[string][]error),
		asyncManager: NewAsyncTaskManager(),
	}

	// 复制蓝图全局变量到执行上下文
	if bp.Variables != nil {
		for k, v := range bp.Variables {
			execCtx.variables[k] = v
		}
	}

	return execCtx
}

// Context 返回底层的 context.Context
func (ec *ExecutionContext) Context() context.Context {
	return ec.ctx
}

// Cancel 取消执行并清理所有异步任务
func (ec *ExecutionContext) Cancel() {
	ec.cancel()
	// 取消所有异步任务，防止 goroutine 泄漏
	if ec.asyncManager != nil {
		ec.asyncManager.CancelAll()
	}
}

// GetVariable 获取变量值（并发安全）
func (ec *ExecutionContext) GetVariable(name string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	val, ok := ec.variables[name]
	return val, ok
}

// SetVariable 设置变量值（并发安全）
func (ec *ExecutionContext) SetVariable(name string, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.variables[name] = value
}

// GetAllVariables 获取所有变量（并发安全）
func (ec *ExecutionContext) GetAllVariables() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	// 返回副本以避免外部修改
	vars := make(map[string]interface{}, len(ec.variables))
	for k, v := range ec.variables {
		vars[k] = v
	}
	return vars
}

// AddError 添加执行错误（并发安全）
func (ec *ExecutionContext) AddError(err error) {
	if err == nil {
		return
	}

	ec.errorsMu.Lock()
	defer ec.errorsMu.Unlock()

	ec.errors = append(ec.errors, err)
}

// GetErrors 获取所有执行错误
func (ec *ExecutionContext) GetErrors() []error {
	ec.errorsMu.Lock()
	defer ec.errorsMu.Unlock()

	// 返回副本
	errs := make([]error, len(ec.errors))
	copy(errs, ec.errors)
	return errs
}

// HasErrors 检查是否有错误
func (ec *ExecutionContext) HasErrors() bool {
	ec.errorsMu.Lock()
	defer ec.errorsMu.Unlock()

	return len(ec.errors) > 0
}

// AddNodeError 添加节点错误（并发安全）
func (ec *ExecutionContext) AddNodeError(nodeID string, err error) {
	if err == nil {
		return
	}

	ec.errorsMu.Lock()
	defer ec.errorsMu.Unlock()

	if ec.nodeErrors[nodeID] == nil {
		ec.nodeErrors[nodeID] = make([]error, 0)
	}
	ec.nodeErrors[nodeID] = append(ec.nodeErrors[nodeID], err)
	ec.errors = append(ec.errors, err)
}

// GetNodeErrors 获取特定节点的错误
func (ec *ExecutionContext) GetNodeErrors(nodeID string) []error {
	ec.errorsMu.Lock()
	defer ec.errorsMu.Unlock()

	if errs, ok := ec.nodeErrors[nodeID]; ok {
		// 返回副本
		result := make([]error, len(errs))
		copy(result, errs)
		return result
	}
	return nil
}

// GetExecutionDuration 获取执行时长
func (ec *ExecutionContext) GetExecutionDuration() time.Duration {
	return time.Since(ec.startTime)
}

// IsCancelled 检查执行是否被取消
func (ec *ExecutionContext) IsCancelled() bool {
	select {
	case <-ec.ctx.Done():
		return true
	default:
		return false
	}
}

// GetAsyncManager 获取异步任务管理器
func (ec *ExecutionContext) GetAsyncManager() *AsyncTaskManager {
	return ec.asyncManager
}

// SetReturned 设置返回状态（类似 return 语句，终止后续执行）
func (ec *ExecutionContext) SetReturned(outputs map[string]interface{}) {
	ec.returnMu.Lock()
	defer ec.returnMu.Unlock()
	ec.returned = true
	ec.returnValue = outputs
}

// IsReturned 检查是否已返回
func (ec *ExecutionContext) IsReturned() bool {
	ec.returnMu.RLock()
	defer ec.returnMu.RUnlock()
	return ec.returned
}

// GetReturnValue 获取返回值
func (ec *ExecutionContext) GetReturnValue() map[string]interface{} {
	ec.returnMu.RLock()
	defer ec.returnMu.RUnlock()
	return ec.returnValue
}
