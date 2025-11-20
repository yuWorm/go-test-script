package blueprint

import (
	"fmt"
	"sync"
	"time"
)

// ExecutionMode 定义执行模式
type ExecutionMode int

const (
	// ExecutionModeSequential 顺序执行（按拓扑顺序）
	ExecutionModeSequential ExecutionMode = iota
	// ExecutionModeParallel 并行执行（在可能的情况下并行执行无依赖的节点）
	ExecutionModeParallel
)

// ExecutionOptions 执行选项
type ExecutionOptions struct {
	Mode           ExecutionMode // 执行模式
	Timeout        time.Duration // 超时时间，0 表示无超时
	MaxConcurrency int           // 最大并发数（仅用于并行模式），0 表示无限制
	StopOnError    bool          // 遇到错误时是否停止执行
}

// DefaultExecutionOptions 返回默认执行选项
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		Mode:           ExecutionModeSequential,
		Timeout:        0,
		MaxConcurrency: 0,
		StopOnError:    true,
	}
}

// Executor 蓝图执行器
type Executor struct {
	options *ExecutionOptions
	mu      sync.RWMutex
}

// NewExecutor 创建一个新的执行器
func NewExecutor(options *ExecutionOptions) *Executor {
	if options == nil {
		options = DefaultExecutionOptions()
	}
	return &Executor{
		options: options,
	}
}

// Execute 执行蓝图
func (e *Executor) Execute(bp *Blueprint, inputs map[string]interface{}) (*ExecutionResult, error) {
	// 检查蓝图是否已编译
	if !bp.IsCompiled() {
		return nil, fmt.Errorf("blueprint is not compiled, please compile it first")
	}

	// 创建执行上下文
	var ctx *ExecutionContext
	if e.options.Timeout > 0 {
		ctx = NewExecutionContextWithTimeout(bp, e.options.Timeout)
	} else {
		ctx = NewExecutionContext(bp)
	}
	defer ctx.Cancel()

	// 设置输入变量
	if inputs != nil {
		for k, v := range inputs {
			ctx.SetVariable(k, v)
		}
	}

	// 重置所有节点的缓存
	for _, node := range bp.Nodes {
		node.ResetCache()
	}

	// 根据执行模式执行
	var err error
	if e.options.Mode == ExecutionModeParallel {
		err = e.executeParallel(ctx, bp)
	} else {
		err = e.executeSequential(ctx, bp)
	}

	// 构建执行结果
	result := &ExecutionResult{
		Success:   !ctx.HasErrors() && err == nil,
		Errors:    ctx.GetErrors(),
		Duration:  ctx.GetExecutionDuration(),
		Variables: ctx.GetAllVariables(),
		Outputs:   e.collectOutputs(bp),
	}

	if err != nil {
		result.Errors = append(result.Errors, err)
		result.Success = false
	}

	return result, nil
}

// executeSequential 顺序执行蓝图
func (e *Executor) executeSequential(ctx *ExecutionContext, bp *Blueprint) error {
	executionOrder, _ := bp.GetExecutionOrder()

	for _, node := range executionOrder {
		// 检查是否取消
		if ctx.IsCancelled() {
			return fmt.Errorf("execution cancelled")
		}

		// 执行节点
		if err := e.executeNode(ctx, bp, node); err != nil {
			ctx.AddError(fmt.Errorf("node %s execution failed: %w", node.ID, err))
			if e.options.StopOnError {
				return err
			}
		}
	}

	return nil
}

// executeParallel 并行执行蓝图
func (e *Executor) executeParallel(ctx *ExecutionContext, bp *Blueprint) error {
	executionOrder, _ := bp.GetExecutionOrder()

	// 构建依赖关系
	dependencies := make(map[string][]string) // nodeID -> dependent node IDs
	for _, conn := range bp.Connections {
		dependencies[conn.TargetNode] = append(dependencies[conn.TargetNode], conn.SourceNode)
	}

	// 已完成的节点
	completed := make(map[string]bool)
	completedMu := sync.RWMutex{}

	// 节点就绪通道和完成计数
	ready := make(chan *Node, len(executionOrder))
	completedCount := 0
	var countMu sync.Mutex
	var firstError error
	var errorMu sync.Mutex

	// 工作协程池
	var wg sync.WaitGroup
	maxWorkers := e.options.MaxConcurrency
	if maxWorkers <= 0 {
		maxWorkers = len(executionOrder) // 无限制
	}

	// 条件变量用于通知新的节点完成
	cond := sync.NewCond(&completedMu)

	// 启动工作协程
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for node := range ready {
				// 检查是否取消
				if ctx.IsCancelled() {
					errorMu.Lock()
					if firstError == nil {
						firstError = fmt.Errorf("execution cancelled")
					}
					errorMu.Unlock()
					return
				}

				// 执行节点
				if err := e.executeNode(ctx, bp, node); err != nil {
					ctx.AddError(fmt.Errorf("node %s execution failed: %w", node.ID, err))
					errorMu.Lock()
					if firstError == nil {
						firstError = err
					}
					errorMu.Unlock()
					if e.options.StopOnError {
						return
					}
				}

				// 标记为已完成
				completedMu.Lock()
				completed[node.ID] = true
				completedMu.Unlock()

				// 增加完成计数
				countMu.Lock()
				completedCount++
				countMu.Unlock()

				// 通知调度器
				cond.Broadcast()
			}
		}()
	}

	// 调度协程：检查并发送就绪的节点
	go func() {
		defer close(ready)
		scheduled := make(map[string]bool)

		for {
			completedMu.Lock()

			// 找到所有就绪的节点
			foundReady := false
			for _, node := range executionOrder {
				if scheduled[node.ID] {
					continue
				}

				// 检查依赖是否都已完成
				deps := dependencies[node.ID]
				allDepsCompleted := true
				for _, depID := range deps {
					if !completed[depID] {
						allDepsCompleted = false
						break
					}
				}

				if allDepsCompleted {
					scheduled[node.ID] = true
					foundReady = true
					completedMu.Unlock()

					select {
					case ready <- node:
					case <-ctx.Context().Done():
						return
					}

					completedMu.Lock()
				}
			}

			// 检查是否所有节点都已调度
			if len(scheduled) >= len(executionOrder) {
				completedMu.Unlock()
				return
			}

			// 如果没有找到就绪的节点，等待通知
			if !foundReady {
				cond.Wait()
			}

			completedMu.Unlock()
		}
	}()

	// 等待所有工作协程完成
	wg.Wait()

	return firstError
}

// executeNode 执行单个节点
func (e *Executor) executeNode(ctx *ExecutionContext, bp *Blueprint, node *Node) error {
	// 收集输入数据
	inputs := make(map[string]interface{})

	// 从连接获取输入
	if connections, exists := bp.connectionMap[node.ID]; exists {
		for _, conn := range connections {
			sourceNode := bp.nodeMap[conn.SourceNode]
			if value, ok := sourceNode.GetOutputValue(conn.SourcePin); ok {
				inputs[conn.TargetPin] = value
			}
		}
	}

	// 从节点默认值获取未连接的输入
	for _, pin := range node.InputPins {
		if _, exists := inputs[pin.Name]; !exists && pin.Value != nil {
			inputs[pin.Name] = pin.Value
		}
	}

	// 执行节点
	if node.executor == nil {
		return fmt.Errorf("node %s has no executor", node.ID)
	}

	outputs, err := node.executor.Execute(ctx, inputs)
	if err != nil {
		return err
	}

	// 保存输出
	for pinName, value := range outputs {
		node.SetOutputValue(pinName, value)
	}

	// 特殊处理：Start 节点的输出来自其输出引脚的默认值
	if node.Type == NodeTypeStart {
		for _, pin := range node.OutputPins {
			if pin.Value != nil {
				node.SetOutputValue(pin.Name, pin.Value)
			}
		}
	}

	return nil
}

// collectOutputs 收集所有输出节点的输出
func (e *Executor) collectOutputs(bp *Blueprint) map[string]interface{} {
	outputs := make(map[string]interface{})

	for _, node := range bp.Nodes {
		// 收集所有节点的输出（可以根据需要过滤特定类型的节点）
		for _, pin := range node.OutputPins {
			if value, ok := node.GetOutputValue(pin.Name); ok {
				key := fmt.Sprintf("%s.%s", node.ID, pin.Name)
				outputs[key] = value
			}
		}
	}

	return outputs
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success   bool                   // 是否成功
	Errors    []error                // 错误列表
	Duration  time.Duration          // 执行时长
	Variables map[string]interface{} // 最终变量状态
	Outputs   map[string]interface{} // 输出值（格式：nodeID.pinName -> value）
}

// GetOutput 获取特定节点的输出值
func (r *ExecutionResult) GetOutput(nodeID, pinName string) (interface{}, bool) {
	key := fmt.Sprintf("%s.%s", nodeID, pinName)
	val, ok := r.Outputs[key]
	return val, ok
}

// GetNodeOutputs 获取特定节点的所有输出
func (r *ExecutionResult) GetNodeOutputs(nodeID string) map[string]interface{} {
	prefix := nodeID + "."
	outputs := make(map[string]interface{})

	for key, value := range r.Outputs {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			pinName := key[len(prefix):]
			outputs[pinName] = value
		}
	}

	return outputs
}
