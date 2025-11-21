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
	// 安全限制
	MaxNodes       int           // 最大节点数，0 表示无限制（默认 1000）
	MaxDepth       int           // 最大递归深度，0 表示无限制（默认 100）
	NodeTimeout    time.Duration // 单节点执行超时，0 表示无超时（默认 30s）
	MaxIterations  int           // 循环最大迭代次数（默认 10000）
}

// DefaultExecutionOptions 返回默认执行选项
func DefaultExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		Mode:           ExecutionModeSequential,
		Timeout:        0,
		MaxConcurrency: 0,
		StopOnError:    true,
		MaxNodes:       1000,
		MaxDepth:       100,
		NodeTimeout:    30 * time.Second,
		MaxIterations:  10000,
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

	// 安全检查：最大节点数
	if e.options.MaxNodes > 0 && len(bp.Nodes) > e.options.MaxNodes {
		return nil, fmt.Errorf("blueprint has %d nodes, exceeds maximum allowed %d", len(bp.Nodes), e.options.MaxNodes)
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

	// 2. 构建连接图
	flowInfo := e.buildFlowInfo(bp)

	// 3. 从Start节点开始递归执行
	executed := &sync.Map{}
	return e.executeNodeRecursive(ctx, bp, startNode, flowInfo, executed, 0)
}

// flowInfo 执行流信息
type flowInfo struct {
	execFlowMap         map[string]map[string][]*execFlowTarget // 执行流
	dataFlowMap         map[string]map[string][]string          // 数据流
	incomingExecMap     map[string]bool                         // 有执行输入的节点
	needsExecActivation map[string]bool                         // 需要执行引脚激活的节点
}

// buildFlowInfo 构建执行流信息
func (e *Executor) buildFlowInfo(bp *Blueprint) *flowInfo {
	info := &flowInfo{
		execFlowMap:         make(map[string]map[string][]*execFlowTarget),
		dataFlowMap:         make(map[string]map[string][]string),
		incomingExecMap:     make(map[string]bool),
		needsExecActivation: make(map[string]bool),
	}

	for _, conn := range bp.Connections {
		sourceNode := bp.nodeMap[conn.SourceNode]
		if sourceNode == nil {
			continue
		}

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
			if info.execFlowMap[conn.SourceNode] == nil {
				info.execFlowMap[conn.SourceNode] = make(map[string][]*execFlowTarget)
			}
			info.execFlowMap[conn.SourceNode][conn.SourcePin] = append(
				info.execFlowMap[conn.SourceNode][conn.SourcePin],
				&execFlowTarget{nodeID: conn.TargetNode, execPinName: conn.TargetPin},
			)
			info.incomingExecMap[conn.TargetNode] = true
		} else {
			if info.dataFlowMap[conn.SourceNode] == nil {
				info.dataFlowMap[conn.SourceNode] = make(map[string][]string)
			}
			info.dataFlowMap[conn.SourceNode][conn.SourcePin] = append(
				info.dataFlowMap[conn.SourceNode][conn.SourcePin],
				conn.TargetNode,
			)
		}
	}

	for _, node := range bp.Nodes {
		for _, pin := range node.InputPins {
			pinKind := pin.Kind
			if pinKind == "" {
				pinKind = PinKindData
			}
			if pinKind == PinKindExecution {
				info.needsExecActivation[node.ID] = true
				break
			}
		}
	}

	return info
}

// executeNodeRecursive 递归执行节点
func (e *Executor) executeNodeRecursive(ctx *ExecutionContext, bp *Blueprint, node *Node, info *flowInfo, executed *sync.Map, depth int) error {
	// 检查递归深度
	if e.options.MaxDepth > 0 && depth > e.options.MaxDepth {
		return fmt.Errorf("execution depth %d exceeds maximum allowed %d", depth, e.options.MaxDepth)
	}

	// 检查是否已执行
	if _, loaded := executed.LoadOrStore(node.ID, true); loaded {
		return nil
	}

	// 检查是否取消
	if ctx.IsCancelled() {
		return fmt.Errorf("execution cancelled")
	}

	// 先执行纯数据节点依赖（无执行引脚的节点）
	e.executeDataDependencies(ctx, bp, node, info, executed)

	// 执行当前节点（带超时）
	if err := e.executeNodeWithTimeout(ctx, bp, node); err != nil {
		ctx.AddError(fmt.Errorf("node %s execution failed: %w", node.ID, err))
		if e.options.StopOnError {
			return err
		}
	}

	// End节点：执行完成即结束，不继续传播
	if node.Type == NodeTypeEnd {
		return nil
	}

	// ForLoop 节点特殊处理：真正执行循环
	if node.Type == NodeTypeFlowControl && node.Operation == "for_loop" {
		return e.executeForLoop(ctx, bp, node, info, executed, depth)
	}

	// 判断是否是序列节点（并行执行）
	isSequence := node.Type == NodeTypeFlowControl && node.Operation == "sequence"

	// 收集要执行的下一批节点
	var nextNodes []*Node
	if execOutputs, exists := info.execFlowMap[node.ID]; exists {
		for execPinName, targets := range execOutputs {
			if e.shouldActivateExecPin(node, execPinName) {
				for _, target := range targets {
					if targetNode := bp.nodeMap[target.nodeID]; targetNode != nil {
						nextNodes = append(nextNodes, targetNode)
					}
				}
			}
		}
	}

	// 序列节点：并行执行所有分支
	if isSequence && len(nextNodes) > 1 {
		var wg sync.WaitGroup
		errChan := make(chan error, len(nextNodes))

		for _, nextNode := range nextNodes {
			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				if err := e.executeNodeRecursive(ctx, bp, n, info, executed, depth+1); err != nil {
					errChan <- err
				}
			}(nextNode)
		}

		wg.Wait()
		close(errChan)

		// 返回第一个错误
		for err := range errChan {
			if err != nil {
				return err
			}
		}
	} else {
		// 顺序执行
		for _, nextNode := range nextNodes {
			if err := e.executeNodeRecursive(ctx, bp, nextNode, info, executed, depth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeForLoop 执行 ForLoop 节点的真正循环
func (e *Executor) executeForLoop(ctx *ExecutionContext, bp *Blueprint, node *Node, info *flowInfo, executed *sync.Map, depth int) error {
	// 获取循环参数
	start := 0.0
	end := 10.0
	step := 1.0

	if v, ok := node.GetOutputValue("start"); ok {
		if f, ok := v.(float64); ok {
			start = f
		}
	} else {
		for _, pin := range node.InputPins {
			if pin.Name == "start" && pin.Value != nil {
				if f, ok := pin.Value.(float64); ok {
					start = f
				}
			}
		}
	}

	if v, ok := node.GetOutputValue("end"); ok {
		if f, ok := v.(float64); ok {
			end = f
		}
	} else {
		for _, pin := range node.InputPins {
			if pin.Name == "end" && pin.Value != nil {
				if f, ok := pin.Value.(float64); ok {
					end = f
				}
			}
		}
	}

	if v, ok := node.GetOutputValue("step"); ok {
		if f, ok := v.(float64); ok {
			step = f
		}
	} else {
		for _, pin := range node.InputPins {
			if pin.Name == "step" && pin.Value != nil {
				if f, ok := pin.Value.(float64); ok {
					step = f
				}
			}
		}
	}

	if step == 0 {
		return fmt.Errorf("for_loop step cannot be zero")
	}

	// 获取 loop_body 连接的节点
	var loopBodyNodes []*Node
	var completedNodes []*Node

	if execOutputs, exists := info.execFlowMap[node.ID]; exists {
		for pinName, targets := range execOutputs {
			for _, target := range targets {
				if targetNode := bp.nodeMap[target.nodeID]; targetNode != nil {
					if pinName == "loop_body" {
						loopBodyNodes = append(loopBodyNodes, targetNode)
					} else if pinName == "completed" {
						completedNodes = append(completedNodes, targetNode)
					}
				}
			}
		}
	}

	// 执行循环
	maxIterations := e.options.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10000
	}

	count := 0
	if step > 0 {
		for i := start; i < end; i += step {
			if ctx.IsCancelled() {
				return fmt.Errorf("execution cancelled")
			}
			if count >= maxIterations {
				return fmt.Errorf("for_loop exceeded maximum iterations (%d)", maxIterations)
			}

			// 设置当前索引输出
			node.SetOutputValue("index", i)
			node.SetOutputValue("count", float64(count))

			// 执行 loop_body 分支（需要重置已执行状态以允许重复执行）
			loopExecuted := &sync.Map{}
			for _, bodyNode := range loopBodyNodes {
				if err := e.executeNodeRecursive(ctx, bp, bodyNode, info, loopExecuted, depth+1); err != nil {
					return err
				}
			}

			count++
		}
	} else {
		for i := start; i > end; i += step {
			if ctx.IsCancelled() {
				return fmt.Errorf("execution cancelled")
			}
			if count >= maxIterations {
				return fmt.Errorf("for_loop exceeded maximum iterations (%d)", maxIterations)
			}

			node.SetOutputValue("index", i)
			node.SetOutputValue("count", float64(count))

			loopExecuted := &sync.Map{}
			for _, bodyNode := range loopBodyNodes {
				if err := e.executeNodeRecursive(ctx, bp, bodyNode, info, loopExecuted, depth+1); err != nil {
					return err
				}
			}

			count++
		}
	}

	// 设置最终输出
	node.SetOutputValue("count", float64(count))

	// 执行 completed 分支
	for _, completedNode := range completedNodes {
		if err := e.executeNodeRecursive(ctx, bp, completedNode, info, executed, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// executeNodeWithTimeout 带超时的节点执行
func (e *Executor) executeNodeWithTimeout(ctx *ExecutionContext, bp *Blueprint, node *Node) error {
	if e.options.NodeTimeout <= 0 {
		return e.executeNode(ctx, bp, node)
	}

	done := make(chan error, 1)
	go func() {
		done <- e.executeNode(ctx, bp, node)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(e.options.NodeTimeout):
		return fmt.Errorf("node %s execution timeout after %v", node.ID, e.options.NodeTimeout)
	case <-ctx.Context().Done():
		return fmt.Errorf("execution cancelled")
	}
}

// executeDataDependencies 执行纯数据节点依赖
func (e *Executor) executeDataDependencies(ctx *ExecutionContext, bp *Blueprint, node *Node, info *flowInfo, executed *sync.Map) {
	// 找到连接到当前节点数据引脚的所有源节点
	for _, conn := range bp.Connections {
		if conn.TargetNode != node.ID {
			continue
		}

		sourceNode := bp.nodeMap[conn.SourceNode]
		if sourceNode == nil {
			continue
		}

		// 如果源节点需要执行引脚激活，跳过（会通过执行流执行）
		if info.needsExecActivation[sourceNode.ID] {
			continue
		}

		// 递归执行纯数据节点
		if _, loaded := executed.LoadOrStore(sourceNode.ID, true); !loaded {
			// 先执行它的依赖
			e.executeDataDependencies(ctx, bp, sourceNode, info, executed)
			// 执行节点
			e.executeNode(ctx, bp, sourceNode)
		}
	}
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
