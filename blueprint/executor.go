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
		Mode:           ExecutionModeExecutionFlow, // 默认使用执行流模式（类似UE5）
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

	// 启用快速模式（ExecutionFlow模式默认单线程执行）
	if e.options.Mode == ExecutionModeExecutionFlow || e.options.Mode == ExecutionModeSequential {
		ctx.EnableFastMode()
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

	// 如果有返回值（通过 End 节点返回），添加到输出中
	if ctx.IsReturned() {
		returnValue := ctx.GetReturnValue()
		for k, v := range returnValue {
			result.Outputs["return."+k] = v
		}
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

		// 检查是否已经通过 End 节点返回（类似 return 语句）
		if ctx.IsReturned() {
			break
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

				// 检查是否已经通过 End 节点返回（类似 return 语句）
				if ctx.IsReturned() {
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

// 优化：复用 inputs/outputs map 池
var inputsPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 8)
	},
}

var outputsPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 8)
	},
}

// executeNode 执行单个节点（优化版）
func (e *Executor) executeNode(ctx *ExecutionContext, bp *Blueprint, node *Node) error {
	// 从池获取 inputs map 并清空复用
	inputs := inputsPool.Get().(map[string]interface{})
	for k := range inputs {
		delete(inputs, k)
	}
	defer inputsPool.Put(inputs)

	// 从连接获取输入（直接访问 outputCache 避免 mutex 开销）
	if connections, exists := bp.connectionMap[node.ID]; exists {
		for _, conn := range connections {
			sourceNode := bp.nodeMap[conn.SourceNode]
			if sourceNode.outputCache != nil {
				if value, ok := sourceNode.outputCache[conn.SourcePin]; ok {
					inputs[conn.TargetPin] = value
				}
			}
		}
	}

	// 从节点默认值获取未连接的输入
	for _, pin := range node.InputPins {
		if _, exists := inputs[pin.Name]; !exists && pin.Value != nil {
			inputs[pin.Name] = pin.Value
		}
	}

	// 特殊处理：Start 节点从变量中获取输入（用于函数参数传递）
	if node.Type == NodeTypeStart {
		for _, pin := range node.InputPins {
			// 如果输入还没有值，尝试从变量中获取
			if _, exists := inputs[pin.Name]; !exists {
				if varValue, exists := ctx.GetVariableFast(pin.Name); exists {
					inputs[pin.Name] = varValue
				}
			}
		}
	}

	// 执行节点
	if node.executor == nil {
		return fmt.Errorf("node %s has no executor", node.ID)
	}

	outputs, err := node.executor.Execute(ctx, inputs)
	if err != nil {
		ctx.AddNodeError(node.ID, err)
		node.SetError(err)
		return err
	}

	// 直接写入 outputCache（单线程热路径安全，跳过 mutex）
	if node.outputCache == nil {
		node.outputCache = make(map[string]interface{}, len(outputs))
	}
	for pinName, value := range outputs {
		node.outputCache[pinName] = value
	}

	// 特殊处理：Start 节点
	if node.Type == NodeTypeStart {
		for _, pin := range node.OutputPins {
			if pin.Value != nil {
				node.outputCache[pin.Name] = pin.Value
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

	// 2. 使用预编译的执行流信息（避免每次执行都重新构建）
	flowInfo := bp.flowInfo
	if flowInfo == nil {
		// 如果没有预编译，则构建（兼容性）
		flowInfo = BuildFlowInfo(bp)
	}

	// 3. 从Start节点开始递归执行
	executed := &sync.Map{}
	returned := &returnSignal{} // 用于标记是否已返回
	return e.executeNodeRecursive(ctx, bp, startNode, flowInfo, executed, returned, 0)
}

// returnSignal 返回信号，用于标记执行是否已通过End节点返回
type returnSignal struct {
	returned bool
	mu       sync.Mutex
}

func (r *returnSignal) SetReturned() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.returned {
		return false // 已经返回过
	}
	r.returned = true
	return true
}

func (r *returnSignal) IsReturned() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.returned
}

// FlowInfo 执行流信息（编译时构建，运行时使用）
type FlowInfo struct {
	ExecFlowMap         map[string]map[string][]*ExecFlowTarget // 执行流
	DataFlowMap         map[string]map[string][]string          // 数据流
	IncomingExecMap     map[string]bool                         // 有执行输入的节点
	NeedsExecActivation map[string]bool                         // 需要执行引脚激活的节点
	// 优化：预编译的数据依赖链（拓扑排序后的执行顺序）
	DataDependencyChain map[string][]*Node // nodeID -> 该节点的数据依赖链（按执行顺序）
}

// BuildFlowInfo 构建执行流信息
func BuildFlowInfo(bp *Blueprint) *FlowInfo {
	info := &FlowInfo{
		ExecFlowMap:         make(map[string]map[string][]*ExecFlowTarget),
		DataFlowMap:         make(map[string]map[string][]string),
		IncomingExecMap:     make(map[string]bool),
		NeedsExecActivation: make(map[string]bool),
		DataDependencyChain: make(map[string][]*Node),
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
			if info.ExecFlowMap[conn.SourceNode] == nil {
				info.ExecFlowMap[conn.SourceNode] = make(map[string][]*ExecFlowTarget)
			}
			info.ExecFlowMap[conn.SourceNode][conn.SourcePin] = append(
				info.ExecFlowMap[conn.SourceNode][conn.SourcePin],
				&ExecFlowTarget{NodeID: conn.TargetNode, ExecPinName: conn.TargetPin},
			)
			info.IncomingExecMap[conn.TargetNode] = true
		} else {
			if info.DataFlowMap[conn.SourceNode] == nil {
				info.DataFlowMap[conn.SourceNode] = make(map[string][]string)
			}
			info.DataFlowMap[conn.SourceNode][conn.SourcePin] = append(
				info.DataFlowMap[conn.SourceNode][conn.SourcePin],
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
				info.NeedsExecActivation[node.ID] = true
				break
			}
		}
	}

	// 预编译每个节点的数据依赖链
	buildDataDependencyChain(bp, info)

	return info
}

// buildDataDependencyChain 为每个节点构建数据依赖链（编译时优化）
func buildDataDependencyChain(bp *Blueprint, info *FlowInfo) {
	// 为每个节点构建其数据依赖的拓扑排序
	for _, node := range bp.Nodes {
		chain := make([]*Node, 0, 8)
		visited := make(map[string]bool, 16)
		collectDataDeps(bp, node, info, visited, &chain)
		if len(chain) > 0 {
			info.DataDependencyChain[node.ID] = chain
		}
	}
}

// collectDataDeps 递归收集数据依赖（深度优先，后序遍历确保依赖先执行）
func collectDataDeps(bp *Blueprint, node *Node, info *FlowInfo, visited map[string]bool, chain *[]*Node) {
	for _, conn := range bp.Connections {
		if conn.TargetNode != node.ID {
			continue
		}
		sourceNode := bp.nodeMap[conn.SourceNode]
		if sourceNode == nil || visited[sourceNode.ID] {
			continue
		}
		// 跳过需要执行引脚激活的节点
		if info.NeedsExecActivation[sourceNode.ID] {
			continue
		}
		visited[sourceNode.ID] = true
		// 先递归处理依赖的依赖
		collectDataDeps(bp, sourceNode, info, visited, chain)
		// 后序添加（依赖先执行）
		*chain = append(*chain, sourceNode)
	}
}

// executeNodeRecursive 递归执行节点
func (e *Executor) executeNodeRecursive(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 检查是否已经通过End节点返回
	if returned.IsReturned() {
		return nil
	}

	// 检查递归深度
	if e.options.MaxDepth > 0 && depth > e.options.MaxDepth {
		return fmt.Errorf("execution depth %d exceeds maximum allowed %d", depth, e.options.MaxDepth)
	}

	// 检查是否已执行（End节点除外，允许多个End节点）
	if node.Type != NodeTypeEnd {
		if _, loaded := executed.LoadOrStore(node.ID, true); loaded {
			return nil
		}
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

	// End节点：标记已返回，停止执行（类似 return 语句）
	if node.Type == NodeTypeEnd {
		returned.SetReturned()
		return nil
	}

	// ForLoop 节点特殊处理：真正执行循环
	if node.Type == NodeTypeFlowControl && node.Operation == "for_loop" {
		return e.executeForLoop(ctx, bp, node, info, executed, returned, depth)
	}

	// WhileLoop 节点特殊处理：真正执行循环
	if node.Type == NodeTypeFlowControl && node.Operation == "while_loop" {
		return e.executeWhileLoop(ctx, bp, node, info, executed, returned, depth)
	}

	// ForEach 节点特殊处理：遍历数组
	if node.Type == NodeTypeFlowControl && node.Operation == "foreach" {
		return e.executeForEach(ctx, bp, node, info, executed, returned, depth)
	}

	// Sequence 节点特殊处理：顺序执行所有输出分支
	if node.Type == NodeTypeFlowControl && node.Operation == "sequence" {
		return e.executeSequenceNode(ctx, bp, node, info, executed, returned, depth)
	}

	// Switch 节点特殊处理：根据值选择分支
	if node.Type == NodeTypeFlowControl && node.Operation == "switch" {
		return e.executeSwitchNode(ctx, bp, node, info, executed, returned, depth)
	}

	// Delay 节点特殊处理：异步延时，让出执行权
	if node.Type == NodeTypeAsync && node.Operation == "delay" {
		return e.executeDelay(ctx, bp, node, info, executed, returned, depth)
	}

	// Join 节点特殊处理：等待所有分支完成
	if node.Type == NodeTypeFlowControl && node.Operation == "join" {
		return e.executeJoin(ctx, bp, node, info, executed, returned, depth)
	}

	// 判断节点类型
	isFork := node.Type == NodeTypeFlowControl && node.Operation == "fork"

	// 收集要执行的下一批节点（按引脚名排序以保证顺序）
	var nextNodes []*Node
	var nextPinNames []string
	if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
		for execPinName, targets := range execOutputs {
			if e.shouldActivateExecPin(node, execPinName) {
				nextPinNames = append(nextPinNames, execPinName)
				for _, target := range targets {
					if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
						nextNodes = append(nextNodes, targetNode)
					}
				}
			}
		}
	}

	// Fork 节点：真正的并行执行（类似 UE5 的 Fork/Spawn）
	if isFork && len(nextNodes) > 0 {
		return e.executeFork(ctx, bp, node, nextNodes, info, executed, returned, depth)
	}

	// 顺序执行
	for _, nextNode := range nextNodes {
		// 检查是否已返回
		if returned.IsReturned() {
			return nil
		}
		if err := e.executeNodeRecursive(ctx, bp, nextNode, info, executed, returned, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// executeForLoop 执行 ForLoop 节点的真正循环
func (e *Executor) executeForLoop(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
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

	if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
		for pinName, targets := range execOutputs {
			for _, target := range targets {
				if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
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
			if returned.IsReturned() || ctx.IsCancelled() {
				break
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
				if err := e.executeNodeRecursive(ctx, bp, bodyNode, info, loopExecuted, returned, depth+1); err != nil {
					return err
				}
			}

			count++
		}
	} else {
		for i := start; i > end; i += step {
			if returned.IsReturned() || ctx.IsCancelled() {
				break
			}
			if count >= maxIterations {
				return fmt.Errorf("for_loop exceeded maximum iterations (%d)", maxIterations)
			}

			node.SetOutputValue("index", i)
			node.SetOutputValue("count", float64(count))

			loopExecuted := &sync.Map{}
			for _, bodyNode := range loopBodyNodes {
				if err := e.executeNodeRecursive(ctx, bp, bodyNode, info, loopExecuted, returned, depth+1); err != nil {
					return err
				}
			}

			count++
		}
	}

	// 设置最终输出
	node.SetOutputValue("count", float64(count))

	// 如果已返回，不执行 completed 分支
	if returned.IsReturned() {
		return nil
	}

	// 执行 completed 分支
	for _, completedNode := range completedNodes {
		if err := e.executeNodeRecursive(ctx, bp, completedNode, info, executed, returned, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// executeWhileLoop 执行 WhileLoop 节点的真正循环（优化版）
func (e *Executor) executeWhileLoop(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 获取 loop 和 done 连接的节点
	var loopBodyNodes []*Node
	var doneNodes []*Node

	if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
		for pinName, targets := range execOutputs {
			for _, target := range targets {
				if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
					if pinName == "loop" {
						loopBodyNodes = append(loopBodyNodes, targetNode)
					} else if pinName == "done" {
						doneNodes = append(doneNodes, targetNode)
					}
				}
			}
		}
	}

	// 预缓存条件连接信息（避免每次迭代遍历）
	var conditionSourceNode *Node
	var conditionSourcePin string
	for _, conn := range bp.Connections {
		if conn.TargetNode == node.ID && conn.TargetPin == "condition" {
			conditionSourceNode = bp.nodeMap[conn.SourceNode]
			conditionSourcePin = conn.SourcePin
			break
		}
	}

	// 获取最大迭代次数
	maxIterations := e.options.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10000
	}

	// 优化：预分配并复用 map（避免每次迭代分配新的 sync.Map）
	conditionExecuted := make(map[string]bool, 16)
	loopExecuted := make(map[string]bool, 32)

	// 执行循环
	count := 0
	for {
		if returned.IsReturned() || ctx.IsCancelled() {
			break
		}
		if count >= maxIterations {
			return fmt.Errorf("while_loop exceeded maximum iterations (%d)", maxIterations)
		}

		// 清空并复用 map（Go 1.21+ clear 或手动清空）
		for k := range conditionExecuted {
			delete(conditionExecuted, k)
		}

		// 使用优化版执行数据依赖
		e.executeDataDependenciesFast(ctx, bp, node, info, conditionExecuted)

		// 从缓存的连接获取条件值
		condition := false
		if conditionSourceNode != nil {
			if v, ok := conditionSourceNode.GetOutputValue(conditionSourcePin); ok {
				node.SetInputValue("condition", v)
				switch c := v.(type) {
				case bool:
					condition = c
				case float64:
					condition = c != 0
				case int:
					condition = c != 0
				}
			}
		}

		// 如果条件为 false，退出循环
		if !condition {
			break
		}

		// 清空并复用 loopExecuted map
		for k := range loopExecuted {
			delete(loopExecuted, k)
		}

		// 执行 loop 分支（使用优化的执行路径）
		for _, bodyNode := range loopBodyNodes {
			if err := e.executeNodeRecursiveFast(ctx, bp, bodyNode, info, loopExecuted, returned, depth+1); err != nil {
				return err
			}
		}

		count++
	}

	// 如果已返回，不执行 done 分支
	if returned.IsReturned() {
		return nil
	}

	// 执行 done 分支
	for _, doneNode := range doneNodes {
		if err := e.executeNodeRecursive(ctx, bp, doneNode, info, executed, returned, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// executeNodeRecursiveFast 优化版递归执行（用于 while 循环热路径）
func (e *Executor) executeNodeRecursiveFast(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed map[string]bool, returned *returnSignal, depth int) error {
	if returned.IsReturned() || ctx.IsCancelled() {
		return nil
	}
	if e.options.MaxDepth > 0 && depth > e.options.MaxDepth {
		return fmt.Errorf("execution depth %d exceeds maximum", depth)
	}
	if executed[node.ID] {
		return nil
	}
	executed[node.ID] = true

	// 执行数据依赖
	e.executeDataDependenciesFast(ctx, bp, node, info, executed)

	// 执行当前节点
	if err := e.executeNode(ctx, bp, node); err != nil {
		if e.options.StopOnError {
			return err
		}
	}

	// End 节点标记返回
	if node.Type == NodeTypeEnd {
		returned.SetReturned()
		return nil
	}

	// 执行下一个节点
	if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
		for execPinName, targets := range execOutputs {
			if e.shouldActivateExecPin(node, execPinName) {
				for _, target := range targets {
					if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
						if err := e.executeNodeRecursiveFast(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
							return err
						}
					}
				}
			}
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

// executeDataDependencies 执行纯数据节点依赖（使用预编译的依赖链）
func (e *Executor) executeDataDependencies(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map) {
	// 使用预编译的依赖链，避免运行时遍历连接
	chain := info.DataDependencyChain[node.ID]
	for _, depNode := range chain {
		if _, loaded := executed.LoadOrStore(depNode.ID, true); !loaded {
			e.executeNode(ctx, bp, depNode)
		}
	}
}

// executeDataDependenciesFast 优化版：使用普通 map（用于 while 循环热路径）
func (e *Executor) executeDataDependenciesFast(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed map[string]bool) {
	chain := info.DataDependencyChain[node.ID]
	for _, depNode := range chain {
		if !executed[depNode.ID] {
			executed[depNode.ID] = true
			e.executeNode(ctx, bp, depNode)
		}
	}
}

// ExecFlowTarget 执行流目标
type ExecFlowTarget struct {
	NodeID      string
	ExecPinName string
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

// executeFork 执行 Fork 节点 - 真正的并行执行
// Fork 会同时启动所有分支，每个分支独立运行
func (e *Executor) executeFork(ctx *ExecutionContext, bp *Blueprint, node *Node, nextNodes []*Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	if len(nextNodes) == 0 {
		return nil
	}

	// 获取 Fork 节点的 join_id（如果有，用于 Join 节点等待）
	var joinID string
	if id, ok := node.GetOutputValue("join_id"); ok {
		joinID, _ = id.(string)
	}
	if joinID == "" {
		joinID = node.ID // 默认使用节点ID
	}

	// 创建完成通道
	branchCount := len(nextNodes)
	completedChan := make(chan bool, branchCount)

	// 将 fork 信息存入上下文，供 Join 节点使用
	ctx.SetVariable("__fork_"+joinID+"_total", branchCount)
	ctx.SetVariable("__fork_"+joinID+"_completed", 0)
	ctx.SetVariable("__fork_"+joinID+"_chan", completedChan)

	// 并行启动所有分支
	for _, nextNode := range nextNodes {
		go func(n *Node) {
			defer func() {
				// 通知完成
				completedChan <- true
				// 更新完成计数
				ctx.mu.Lock()
				if count, ok := ctx.variables["__fork_"+joinID+"_completed"].(int); ok {
					ctx.variables["__fork_"+joinID+"_completed"] = count + 1
				}
				ctx.mu.Unlock()
			}()

			// 每个分支有独立的 returnSignal（不影响其他分支）
			branchReturned := &returnSignal{}
			// 每个分支有独立的 executed map（允许同一节点在不同分支执行）
			branchExecuted := &sync.Map{}

			e.executeNodeRecursive(ctx, bp, n, info, branchExecuted, branchReturned, depth+1)
		}(nextNode)
	}

	// Fork 节点不等待分支完成，立即返回
	// 如果需要等待，使用 Join 节点
	return nil
}

// executeDelay 执行 Delay 节点 - 异步延时，让出执行权
func (e *Executor) executeDelay(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 获取延时时间（秒）
	duration := 1.0
	if d, ok := node.GetInputValue("duration"); ok {
		if f, err := toFloat64Value(d); err == nil {
			duration = f
		}
	}

	// 使用 AsyncTaskManager 追踪异步任务
	taskID := fmt.Sprintf("delay_%s_%d", node.ID, time.Now().UnixNano())
	asyncManager := ctx.GetAsyncManager()
	task := asyncManager.CreateTask(taskID)

	// 异步延时：启动 goroutine 等待后继续执行
	go func() {
		defer asyncManager.RemoveTask(taskID) // 自动清理任务

		// 使用 select 响应取消信号
		select {
		case <-time.After(time.Duration(duration * float64(time.Second))):
			// 延时完成
		case <-task.Ctx.Done():
			// 任务被取消
			return
		case <-ctx.Context().Done():
			// 上下文被取消
			return
		}

		// 检查是否已返回
		if returned.IsReturned() {
			return
		}

		// 设置输出
		node.SetOutputValue("completed", true)

		// 继续执行后续节点
		if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
			for execPinName, targets := range execOutputs {
				if e.shouldActivateExecPin(node, execPinName) {
					for _, target := range targets {
						if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
							e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1)
						}
					}
				}
			}
		}
	}()

	// 立即返回，让出执行权
	return nil
}

// toFloat64Value 辅助函数：转换为 float64
func toFloat64Value(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// executeJoin 执行 Join 节点 - 等待所有 Fork 分支完成
func (e *Executor) executeJoin(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 获取要等待的 fork_id
	var forkID string
	if id, ok := node.GetInputValue("fork_id"); ok {
		forkID, _ = id.(string)
	}
	if forkID == "" {
		// 尝试从属性获取
		if id, ok := node.Properties["fork_id"]; ok {
			forkID, _ = id.(string)
		}
	}
	if forkID == "" {
		return fmt.Errorf("join node requires fork_id")
	}

	// 获取超时时间（秒）
	timeout := 30.0
	if t, ok := node.GetInputValue("timeout"); ok {
		if f, err := toFloat64Value(t); err == nil {
			timeout = f
		}
	}

	// 获取 fork 信息
	totalVar, _ := ctx.GetVariable("__fork_" + forkID + "_total")
	total, ok := totalVar.(int)
	if !ok {
		return fmt.Errorf("fork %s not found or not started", forkID)
	}

	chanVar, _ := ctx.GetVariable("__fork_" + forkID + "_chan")
	completedChan, ok := chanVar.(chan bool)
	if !ok {
		return fmt.Errorf("fork %s channel not found", forkID)
	}

	// 等待所有分支完成
	timeoutDuration := time.Duration(timeout * float64(time.Second))
	timer := time.NewTimer(timeoutDuration)
	defer timer.Stop()

	completed := 0
	for completed < total {
		select {
		case <-completedChan:
			completed++
		case <-timer.C:
			return fmt.Errorf("join timeout waiting for fork %s (completed %d/%d)", forkID, completed, total)
		case <-ctx.Context().Done():
			return fmt.Errorf("execution cancelled while waiting for fork %s", forkID)
		}
	}

	// 设置输出
	node.SetOutputValue("completed", true)
	node.SetOutputValue("branch_count", float64(total))

	// 继续执行后续节点
	if execOutputs, exists := info.ExecFlowMap[node.ID]; exists {
		for execPinName, targets := range execOutputs {
			if e.shouldActivateExecPin(node, execPinName) {
				for _, target := range targets {
					if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
						if err := e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// executeSequenceNode 执行 Sequence 节点（顺序触发所有输出引脚）
func (e *Executor) executeSequenceNode(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 获取所有执行输出引脚并排序（Then 0, Then 1, Then 2...）
	execOutputs, exists := info.ExecFlowMap[node.ID]
	if !exists {
		return nil
	}

	// 收集所有 exec 输出引脚
	var pinNames []string
	for pinName := range execOutputs {
		pinNames = append(pinNames, pinName)
	}

	// 按引脚名排序（保证 Then 0 -> Then 1 -> Then 2）
	// 简单排序即可，因为 Then 0, Then 1, Then 2 按字符串排序也是正确顺序
	// 但为了更准确，我们可以手动排序
	for i := 0; i < len(pinNames); i++ {
		for j := i + 1; j < len(pinNames); j++ {
			if pinNames[i] > pinNames[j] {
				pinNames[i], pinNames[j] = pinNames[j], pinNames[i]
			}
		}
	}

	// 按顺序执行每个输出分支
	for _, pinName := range pinNames {
		if returned.IsReturned() {
			return nil
		}

		targets := execOutputs[pinName]
		for _, target := range targets {
			if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
				if err := e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
					if e.options.StopOnError {
						return err
					}
				}
			}
		}
	}

	return nil
}

// executeSwitchNode 执行 Switch 节点（根据值选择分支）
func (e *Executor) executeSwitchNode(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 从输出缓存获取选择值
	selectedIndex, hasIndex := node.GetOutputValue("selected_index")
	selectedValue, hasValue := node.GetOutputValue("selected_value")

	execOutputs, exists := info.ExecFlowMap[node.ID]
	if !exists {
		return nil
	}

	// 根据选择值找到对应的输出引脚
	var targetPinName string

	if hasIndex {
		// 数值类型：匹配 case_0, case_1, case_2...
		idx := 0
		if floatVal, err := toFloat64Value(selectedIndex); err == nil {
			idx = int(floatVal)
		}
		targetPinName = fmt.Sprintf("case_%d", idx)
	} else if hasValue {
		// 字符串类型：匹配 case_xxx
		str := selectedValue.(string)
		targetPinName = fmt.Sprintf("case_%s", str)
	}

	// 尝试执行匹配的分支
	if targets, exists := execOutputs[targetPinName]; exists {
		for _, target := range targets {
			if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
				if err := e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// 如果没有匹配，执行 default 分支
	if targets, exists := execOutputs["default"]; exists {
		for _, target := range targets {
			if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
				if err := e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// executeForEach 执行 ForEach 节点（遍历数组）
func (e *Executor) executeForEach(ctx *ExecutionContext, bp *Blueprint, node *Node, info *FlowInfo, executed *sync.Map, returned *returnSignal, depth int) error {
	// 从输出缓存获取数组
	arrayValue, exists := node.GetOutputValue("array")
	if !exists {
		return fmt.Errorf("foreach array not found")
	}

	array, ok := arrayValue.([]interface{})
	if !ok {
		return fmt.Errorf("foreach array must be []interface{}, got %T", arrayValue)
	}

	// 查找 loop_body 执行引脚连接
	execOutputs, exists := info.ExecFlowMap[node.ID]
	if !exists {
		return nil
	}

	loopBodyTargets, hasLoopBody := execOutputs["loop_body"]
	if !hasLoopBody || len(loopBodyTargets) == 0 {
		// 没有 loop body，跳过循环，直接执行 completed
		if completedTargets, exists := execOutputs["completed"]; exists {
			for _, target := range completedTargets {
				if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
					return e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1)
				}
			}
		}
		return nil
	}

	// 遍历数组
	iterationCount := 0
	for idx, element := range array {
		// 检查是否已返回
		if returned.IsReturned() {
			break
		}

		// 检查是否取消
		if ctx.IsCancelled() {
			return fmt.Errorf("execution cancelled during foreach")
		}

		// 检查最大迭代次数
		if e.options.MaxIterations > 0 && iterationCount >= e.options.MaxIterations {
			return fmt.Errorf("foreach iteration count %d exceeds maximum allowed %d", iterationCount, e.options.MaxIterations)
		}

		// 更新循环索引和元素到输出缓存
		node.SetOutputValue("index", float64(idx))
		node.SetOutputValue("element", element)
		node.SetOutputValue("array_index", float64(idx))

		// 创建局部执行标记（每次循环体独立）
		loopExecuted := &sync.Map{}

		// 执行循环体
		for _, target := range loopBodyTargets {
			if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
				if err := e.executeNodeRecursive(ctx, bp, targetNode, info, loopExecuted, returned, depth+1); err != nil {
					if e.options.StopOnError {
						return err
					}
				}
			}
		}

		iterationCount++
	}

	// 循环完成后，执行 completed 分支
	if completedTargets, exists := execOutputs["completed"]; exists {
		for _, target := range completedTargets {
			if targetNode := bp.nodeMap[target.NodeID]; targetNode != nil {
				if err := e.executeNodeRecursive(ctx, bp, targetNode, info, executed, returned, depth+1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
