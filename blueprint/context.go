package blueprint

import (
	"context"
	"sync"
	"time"
)

// ExecutionContext 表示蓝图执行的上下文
type ExecutionContext struct {
	ctx       context.Context
	cancel    context.CancelFunc
	blueprint *Blueprint
	variables map[string]interface{} // 运行时变量
	mu        sync.RWMutex
	startTime time.Time
	errors    []error
	errorsMu  sync.Mutex
}

// NewExecutionContext 创建一个新的执行上下文
func NewExecutionContext(bp *Blueprint) *ExecutionContext {
	ctx, cancel := context.WithCancel(context.Background())

	execCtx := &ExecutionContext{
		ctx:       ctx,
		cancel:    cancel,
		blueprint: bp,
		variables: make(map[string]interface{}),
		startTime: time.Now(),
		errors:    make([]error, 0),
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
		ctx:       ctx,
		cancel:    cancel,
		blueprint: bp,
		variables: make(map[string]interface{}),
		startTime: time.Now(),
		errors:    make([]error, 0),
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

// Cancel 取消执行
func (ec *ExecutionContext) Cancel() {
	ec.cancel()
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
