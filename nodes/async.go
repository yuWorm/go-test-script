package nodes

import (
	"fmt"
	"time"

	"github.com/go-blueprint-engine/blueprint"
)

// AsyncStartExecutor 异步开始执行器
type AsyncStartExecutor struct{}

// NewAsyncStartExecutor 创建异步开始执行器
func NewAsyncStartExecutor() *AsyncStartExecutor {
	return &AsyncStartExecutor{}
}

// Execute 执行异步开始
func (e *AsyncStartExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 自动生成唯一的异步任务 ID
	taskID := fmt.Sprintf("async_task_%d", time.Now().UnixNano())

	// 创建异步任务
	asyncManager := ctx.GetAsyncManager()
	task := asyncManager.CreateTask(taskID)

	// 输出任务 ID 和任务句柄
	outputs := map[string]interface{}{
		"task_id":  taskID,
		"started":  true,
		"task_ptr": task, // 传递任务指针供后续节点使用
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *AsyncStartExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// AsyncEndExecutor 异步结束执行器
type AsyncEndExecutor struct{}

// NewAsyncEndExecutor 创建异步结束执行器
func NewAsyncEndExecutor() *AsyncEndExecutor {
	return &AsyncEndExecutor{}
}

// Execute 执行异步结束
func (e *AsyncEndExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取任务 ID
	taskID, ok := inputs["task_id"].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	// 获取结果值
	result := inputs["result"]

	// 获取异步任务
	asyncManager := ctx.GetAsyncManager()
	task, err := asyncManager.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 设置任务结果
	task.SetResult(result)

	outputs := map[string]interface{}{
		"task_id":   taskID,
		"completed": true,
		"result":    result,
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *AsyncEndExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// AwaitExecutor 等待异步结果执行器
type AwaitExecutor struct{}

// NewAwaitExecutor 创建等待执行器
func NewAwaitExecutor() *AwaitExecutor {
	return &AwaitExecutor{}
}

// Execute 执行等待
func (e *AwaitExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取任务 ID
	taskID, ok := inputs["task_id"].(string)
	if !ok || taskID == "" {
		return nil, fmt.Errorf("task_id is required for await")
	}

	// 获取超时时间（秒）
	timeoutSec := 30.0
	if t, ok := inputs["timeout"]; ok {
		if timeout, err := toFloat64(t); err == nil && timeout > 0 {
			timeoutSec = timeout
		}
	}

	// 获取异步任务
	asyncManager := ctx.GetAsyncManager()
	task, err := asyncManager.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 等待任务完成
	timeout := time.Duration(timeoutSec * float64(time.Second))
	result, err := task.Wait(timeout)
	if err != nil {
		return nil, fmt.Errorf("await failed: %w", err)
	}

	outputs := map[string]interface{}{
		"result":    result,
		"task_id":   taskID,
		"completed": true,
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *AwaitExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// SleepExecutor 休眠执行器（用于模拟耗时操作）
type SleepExecutor struct{}

// NewSleepExecutor 创建休眠执行器
func NewSleepExecutor() *SleepExecutor {
	return &SleepExecutor{}
}

// Execute 执行休眠
func (e *SleepExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取休眠时间（毫秒）
	durationMs := 1000.0
	if d, ok := inputs["duration"]; ok {
		if duration, err := toFloat64(d); err == nil && duration > 0 {
			durationMs = duration
		}
	}

	// 执行休眠
	duration := time.Duration(durationMs * float64(time.Millisecond))
	time.Sleep(duration)

	outputs := map[string]interface{}{
		"duration": durationMs,
		"done":     true,
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *SleepExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ParallelExecutor 并行执行多个任务
type ParallelExecutor struct{}

// NewParallelExecutor 创建并行执行器
func NewParallelExecutor() *ParallelExecutor {
	return &ParallelExecutor{}
}

// Execute 执行并行任务
func (e *ParallelExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取任务 ID 列表
	taskIDs := make([]string, 0)

	// 从输入中提取所有 task_id
	for _, value := range inputs {
		if strVal, ok := value.(string); ok && len(strVal) > 0 {
			taskIDs = append(taskIDs, strVal)
		}
	}

	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("no task IDs provided for parallel execution")
	}

	// 等待所有任务完成
	asyncManager := ctx.GetAsyncManager()
	results := make(map[string]interface{})

	for i, taskID := range taskIDs {
		task, err := asyncManager.GetTask(taskID)
		if err != nil {
			return nil, fmt.Errorf("task %s not found: %w", taskID, err)
		}

		// 等待任务（30秒超时）
		result, err := task.Wait(30 * time.Second)
		if err != nil {
			return nil, fmt.Errorf("task %s failed: %w", taskID, err)
		}

		resultKey := fmt.Sprintf("result_%d", i)
		results[resultKey] = result
	}

	results["count"] = float64(len(taskIDs))
	results["all_completed"] = true

	return results, nil
}

// Validate 验证节点配置
func (e *ParallelExecutor) Validate(node *blueprint.Node) error {
	return nil
}
