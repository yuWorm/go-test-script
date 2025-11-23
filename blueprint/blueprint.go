package blueprint

import (
	"fmt"
	"sync"
)

// Blueprint 表示一个完整的蓝图
type Blueprint struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description,omitempty"`
	Nodes       []*Node                `json:"nodes"`
	Connections []Connection           `json:"connections"`
	Variables   map[string]interface{} `json:"variables,omitempty"` // 全局变量
	Metadata    map[string]interface{} `json:"metadata,omitempty"`  // 元数据

	// 编译后的数据（不序列化）
	compiled       bool
	nodeMap        map[string]*Node        // ID -> Node 快速查找
	executionOrder []*Node                 // 拓扑排序后的执行顺序
	connectionMap  map[string][]Connection // target_node -> connections
	flowInfo       *FlowInfo               // 执行流信息（编译时构建）
	mu             sync.RWMutex            // 并发安全
}

// NewBlueprint 创建一个新的蓝图
func NewBlueprint(name string) *Blueprint {
	return &Blueprint{
		Name:      name,
		Version:   "1.0.0",
		Nodes:     make([]*Node, 0),
		Connections: make([]Connection, 0),
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		nodeMap:   make(map[string]*Node),
	}
}

// AddNode 添加节点到蓝图
func (b *Blueprint) AddNode(node *Node) error {
	if node == nil {
		return fmt.Errorf("cannot add nil node")
	}

	if err := node.Validate(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.nodeMap[node.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	b.Nodes = append(b.Nodes, node)
	b.nodeMap[node.ID] = node
	b.compiled = false // 需要重新编译

	return nil
}

// AddConnection 添加连接到蓝图
func (b *Blueprint) AddConnection(conn Connection) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 验证连接的节点和引脚是否存在
	sourceNode, exists := b.nodeMap[conn.SourceNode]
	if !exists {
		return fmt.Errorf("source node %s not found", conn.SourceNode)
	}

	targetNode, exists := b.nodeMap[conn.TargetNode]
	if !exists {
		return fmt.Errorf("target node %s not found", conn.TargetNode)
	}

	// 验证源引脚
	var sourcePin *Pin
	for i := range sourceNode.OutputPins {
		if sourceNode.OutputPins[i].Name == conn.SourcePin {
			sourcePin = &sourceNode.OutputPins[i]
			break
		}
	}
	if sourcePin == nil {
		return fmt.Errorf("source pin %s not found in node %s", conn.SourcePin, conn.SourceNode)
	}

	// 验证目标引脚
	var targetPin *Pin
	for i := range targetNode.InputPins {
		if targetNode.InputPins[i].Name == conn.TargetPin {
			targetPin = &targetNode.InputPins[i]
			break
		}
	}
	if targetPin == nil {
		return fmt.Errorf("target pin %s not found in node %s", conn.TargetPin, conn.TargetNode)
	}

	// 验证引脚类型匹配
	// 如果Kind为空，默认为数据引脚（向后兼容）
	sourcePinKind := sourcePin.Kind
	if sourcePinKind == "" {
		sourcePinKind = PinKindData
	}
	targetPinKind := targetPin.Kind
	if targetPinKind == "" {
		targetPinKind = PinKindData
	}

	// 执行引脚只能连接到执行引脚，数据引脚只能连接到数据引脚
	if sourcePinKind != targetPinKind {
		return fmt.Errorf("pin kind mismatch: cannot connect %s pin to %s pin", sourcePinKind, targetPinKind)
	}

	b.Connections = append(b.Connections, conn)
	b.compiled = false // 需要重新编译

	return nil
}

// GetNode 根据ID获取节点
func (b *Blueprint) GetNode(nodeID string) (*Node, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	node, exists := b.nodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node, nil
}

// Validate 验证蓝图的完整性
func (b *Blueprint) Validate() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Name == "" {
		return fmt.Errorf("blueprint name cannot be empty")
	}

	// 验证蓝图名称格式
	if len(b.Name) > 256 {
		return fmt.Errorf("blueprint name too long (max 256 characters)")
	}

	// 验证所有节点
	nodeIDs := make(map[string]bool)
	for _, node := range b.Nodes {
		if err := node.Validate(); err != nil {
			return err
		}
		// 验证节点ID格式
		if len(node.ID) > 128 {
			return fmt.Errorf("node ID too long: %s (max 128 characters)", node.ID)
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// 验证变量名
	for varName := range b.Variables {
		if len(varName) > 128 {
			return fmt.Errorf("variable name too long: %s (max 128 characters)", varName)
		}
		if !isValidIdentifier(varName) {
			return fmt.Errorf("invalid variable name: %s (must be alphanumeric with underscores)", varName)
		}
	}

	// 验证所有连接
	for i, conn := range b.Connections {
		if conn.SourceNode == "" || conn.TargetNode == "" {
			return fmt.Errorf("connection %d: source or target node cannot be empty", i)
		}
		if conn.SourcePin == "" || conn.TargetPin == "" {
			return fmt.Errorf("connection %d: source or target pin cannot be empty", i)
		}

		// 检查节点是否存在
		if !nodeIDs[conn.SourceNode] {
			return fmt.Errorf("connection %d: source node %s not found", i, conn.SourceNode)
		}
		if !nodeIDs[conn.TargetNode] {
			return fmt.Errorf("connection %d: target node %s not found", i, conn.TargetNode)
		}
	}

	// 检查循环依赖
	if err := b.detectCycles(); err != nil {
		return err
	}

	return nil
}

// detectCycles 检测蓝图中是否存在循环依赖
// 注意：只检查数据流连接，执行流连接允许循环（如 while 循环）
func (b *Blueprint) detectCycles() error {
	// 构建邻接表（只包含数据流连接）
	graph := make(map[string][]string)
	for _, conn := range b.Connections {
		sourceNode := b.nodeMap[conn.SourceNode]
		if sourceNode == nil {
			continue
		}

		// 检查是否为执行流连接
		isExecFlow := false
		for _, pin := range sourceNode.OutputPins {
			if pin.Name == conn.SourcePin && pin.Kind == PinKindExecution {
				isExecFlow = true
				break
			}
		}

		// 跳过执行流连接
		if isExecFlow {
			continue
		}

		graph[conn.SourceNode] = append(graph[conn.SourceNode], conn.TargetNode)
	}

	// DFS 检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(nodeID string) bool
	hasCycle = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, neighbor := range graph[nodeID] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[nodeID] = false
		return false
	}

	for nodeID := range b.nodeMap {
		if !visited[nodeID] {
			if hasCycle(nodeID) {
				return fmt.Errorf("circular dependency detected in blueprint")
			}
		}
	}

	return nil
}

// IsCompiled 检查蓝图是否已编译
func (b *Blueprint) IsCompiled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.compiled
}

// GetExecutionOrder 获取节点执行顺序（仅在编译后可用）
func (b *Blueprint) GetExecutionOrder() ([]*Node, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.compiled {
		return nil, fmt.Errorf("blueprint is not compiled")
	}

	return b.executionOrder, nil
}

// InitNodeMap 初始化节点映射（用于从 JSON 反序列化后）
func (b *Blueprint) InitNodeMap() {
	b.nodeMap = make(map[string]*Node)
	for _, node := range b.Nodes {
		b.nodeMap[node.ID] = node
	}
}

// isValidIdentifier 验证标识符是否有效（字母数字下划线，不以数字开头）
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}

// FunctionRegistry 函数蓝图注册表
type FunctionRegistry struct {
	functions map[string]*Blueprint // function_name -> Blueprint
	mu        sync.RWMutex
}

// NewFunctionRegistry 创建函数注册表
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]*Blueprint),
	}
}

// RegisterFunction 注册函数蓝图
func (fr *FunctionRegistry) RegisterFunction(name string, bp *Blueprint) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if !bp.IsCompiled() {
		return fmt.Errorf("function blueprint %s must be compiled before registration", name)
	}

	fr.functions[name] = bp
	return nil
}

// GetFunction 获取函数蓝图
func (fr *FunctionRegistry) GetFunction(name string) (*Blueprint, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	if bp, exists := fr.functions[name]; exists {
		return bp, nil
	}
	return nil, fmt.Errorf("function %s not found", name)
}
