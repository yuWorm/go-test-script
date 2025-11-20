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
	// ExecutionModeExecutionFlow 执行流模式（类似UE5，只执行被执行引脚连接的节点）
	ExecutionModeExecutionFlow
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
	switch e.options.Mode {
	case ExecutionModeParallel:
		err = e.executeParallel(ctx, bp)
	case ExecutionModeExecutionFlow:
		err = e.executeWithExecutionFlow(ctx, bp)
	default:
		err = e.executeSequential(ctx, bp)
	}

	// 构建执行结果
	result := &ExecutionResult{
		Success:   !ctx.HasErrors() && err == nil,
		Errors:    ctx.GetErrors(),
		Duration:  ctx.GetExecutionDuration(),
		Variables: ctx.GetAllVariables(),
		Outputs:   e.collectOutputs(bp),
		Nodes:     e.collectNodeInfo(ctx, bp),
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
		// 记录节点错误
		ctx.AddNodeError(node.ID, err)
		node.SetError(err)
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
		// 收集所有节点的输出缓存
		node.mu.RLock()
		for pinName, value := range node.outputCache {
			key := fmt.Sprintf("%s.%s", node.ID, pinName)
			outputs[key] = value
		}
		node.mu.RUnlock()
	}

	return outputs
}

// NodeExecutionInfo 节点执行信息
type NodeExecutionInfo struct {
	NodeID   string                 `json:"node_id"`  // 节点ID
	Status   string                 `json:"status"`   // 状态：success, error, skipped, idle
	Error    string                 `json:"error"`    // 错误信息（如果有）
	Outputs  map[string]interface{} `json:"outputs"`  // 节点输出
	Duration time.Duration          `json:"duration"` // 执行时长
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success   bool                         // 是否成功
	Errors    []error                      // 错误列表
	Duration  time.Duration                // 执行时长
	Variables map[string]interface{}       // 最终变量状态
	Outputs   map[string]interface{}       // 输出值（格式：nodeID.pinName -> value）
	Nodes     map[string]*NodeExecutionInfo // 每个节点的执行信息
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

// executeWithExecutionFlow 使用执行流模式执行蓝图
// 从Start节点开始，只执行可达节点（类似图形化脚本语言）
func (e *Executor) executeWithExecutionFlow(ctx *ExecutionContext, bp *Blueprint) error {
	// 1. 找到Start节点
	var startNode *Node
	for _, node := range bp.Nodes {
		if node.Type == NodeTypeStart {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return fmt.Errorf("no start node found in blueprint")
	}

	// 2. 构建连接图（包括执行连接和数据连接）
	execFlowMap := make(map[string]map[string][]*execFlowTarget)    // 执行流
	dataFlowMap := make(map[string]map[string][]string)              // 数据流：nodeID -> pinName -> [targetNodeIDs]
	incomingExecMap := make(map[string]bool)                         // 有执行输入的节点

	for _, conn := range bp.Connections {
		sourceNode := bp.nodeMap[conn.SourceNode]
		if sourceNode == nil {
			continue
		}

		// 查找源引脚
		var sourcePin *Pin
		for i := range sourceNode.OutputPins {
			if sourceNode.OutputPins[i].Name == conn.SourcePin {
				sourcePin = &sourceNode.OutputPins[i]
				break
			}
		}

		if sourcePin == nil {
			continue
		}

		sourcePinKind := sourcePin.Kind
		if sourcePinKind == "" {
			sourcePinKind = PinKindData
		}

		if sourcePinKind == PinKindExecution {
			// 执行引脚连接
			if execFlowMap[conn.SourceNode] == nil {
				execFlowMap[conn.SourceNode] = make(map[string][]*execFlowTarget)
			}
			execFlowMap[conn.SourceNode][conn.SourcePin] = append(
				execFlowMap[conn.SourceNode][conn.SourcePin],
				&execFlowTarget{
					nodeID:      conn.TargetNode,
					execPinName: conn.TargetPin,
				},
			)
			incomingExecMap[conn.TargetNode] = true
		} else {
			// 数据引脚连接
			if dataFlowMap[conn.SourceNode] == nil {
				dataFlowMap[conn.SourceNode] = make(map[string][]string)
			}
			dataFlowMap[conn.SourceNode][conn.SourcePin] = append(
				dataFlowMap[conn.SourceNode][conn.SourcePin],
				conn.TargetNode,
			)
		}
	}

	// 3. 检测哪些节点需要执行引脚
	// 规则：如果节点有执行输入引脚定义，则必须通过执行流激活
	needsExecActivation := make(map[string]bool)
	for _, node := range bp.Nodes {
		for _, pin := range node.InputPins {
			pinKind := pin.Kind
			if pinKind == "" {
				pinKind = PinKindData
			}
			if pinKind == PinKindExecution {
				needsExecActivation[node.ID] = true
				break
			}
		}
	}

	// 4. 使用广度优先搜索从Start节点开始执行
	executed := make(map[string]bool)
	queue := []*Node{startNode}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		// 检查是否已执行
		if executed[node.ID] {
			continue
		}

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

		executed[node.ID] = true

		// 5. 确定下一步要执行的节点
		nextNodes := make(map[string]bool)

		// 5a. 通过执行引脚连接的节点
		if execOutputs, exists := execFlowMap[node.ID]; exists {
			for execPinName, targets := range execOutputs {
				// 检查执行引脚是否应该激活
				if e.shouldActivateExecPin(node, execPinName) {
					for _, target := range targets {
						if !executed[target.nodeID] {
							nextNodes[target.nodeID] = true
						}
					}
				}
			}
		}

		// 5b. 通过数据引脚连接的节点（仅当目标节点不需要执行引脚激活时）
		if dataOutputs, exists := dataFlowMap[node.ID]; exists {
			for _, targetNodeIDs := range dataOutputs {
				for _, targetNodeID := range targetNodeIDs {
					// 如果目标节点需要执行引脚激活，则跳过
					if needsExecActivation[targetNodeID] {
						continue
					}
					// 如果目标节点已经有执行输入连接，也跳过
					if incomingExecMap[targetNodeID] {
						continue
					}
					if !executed[targetNodeID] {
						nextNodes[targetNodeID] = true
					}
				}
			}
		}

		// 5c. 将下一步节点加入队列
		for nodeID := range nextNodes {
			if targetNode := bp.nodeMap[nodeID]; targetNode != nil {
				queue = append(queue, targetNode)
			}
		}
	}

	return nil
}

// execFlowTarget 执行流目标
type execFlowTarget struct {
	nodeID      string
	execPinName string
}

// shouldActivateExecPin 判断执行引脚是否应该激活
func (e *Executor) shouldActivateExecPin(node *Node, execPinName string) bool {
	// 对于Branch节点，检查输出来决定激活哪个分支
	if node.Operation == "branch" || node.Operation == "if_else" {
		// 检查节点的输出
		if execOutput, ok := node.GetOutputValue(execPinName); ok {
			// 如果输出是布尔值且为true，则激活
			if boolVal, isBool := execOutput.(bool); isBool {
				return boolVal
			}
		}
		return false
	}

	// 默认情况下，激活所有执行输出
	return true
}

// collectNodeInfo 收集每个节点的执行信息
func (e *Executor) collectNodeInfo(ctx *ExecutionContext, bp *Blueprint) map[string]*NodeExecutionInfo {
	nodeInfos := make(map[string]*NodeExecutionInfo)

	for _, node := range bp.Nodes {
		info := &NodeExecutionInfo{
			NodeID:  node.ID,
			Status:  "idle",
			Outputs: make(map[string]interface{}),
		}

		// 检查节点是否有输出（说明已执行）
		node.mu.RLock()
		if node.outputCache != nil && len(node.outputCache) > 0 {
			info.Status = "success"
			// 复制输出
			for k, v := range node.outputCache {
				info.Outputs[k] = v
			}
		} else if node.lastError != nil {
			info.Status = "error"
			info.Error = node.lastError.Error()
		}
		node.mu.RUnlock()

		// 如果上下文中有该节点的错误信息
		nodeErrors := ctx.GetNodeErrors(node.ID)
		if len(nodeErrors) > 0 {
			info.Status = "error"
			info.Error = nodeErrors[0].Error()
		}

		nodeInfos[node.ID] = info
	}

	return nodeInfos
}
