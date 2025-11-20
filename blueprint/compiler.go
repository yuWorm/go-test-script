package blueprint

import (
	"fmt"
	"sync"
)

// Compiler 负责编译蓝图，预处理为可执行对象
type Compiler struct {
	nodeRegistry *NodeRegistry // 节点注册表，用于关联执行器
	mu           sync.RWMutex
}

// NewCompiler 创建一个新的编译器
func NewCompiler(registry *NodeRegistry) *Compiler {
	return &Compiler{
		nodeRegistry: registry,
	}
}

// Compile 编译蓝图
// 这个方法会：
// 1. 验证蓝图结构
// 2. 进行拓扑排序确定执行顺序
// 3. 关联节点执行器
// 4. 构建连接映射
// 5. 优化执行路径
func (c *Compiler) Compile(bp *Blueprint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. 验证蓝图
	if err := bp.Validate(); err != nil {
		return fmt.Errorf("blueprint validation failed: %w", err)
	}

	// 2. 确保 nodeMap 已初始化
	if bp.nodeMap == nil {
		bp.nodeMap = make(map[string]*Node)
		for _, node := range bp.Nodes {
			bp.nodeMap[node.ID] = node
		}
	}

	// 3. 关联节点执行器
	if err := c.attachExecutors(bp); err != nil {
		return fmt.Errorf("failed to attach executors: %w", err)
	}

	// 4. 构建连接映射（需要在类型转换之前，以便知道哪些引脚有连接）
	bp.connectionMap = c.buildConnectionMap(bp)

	// 5. 调用节点的编译方法（如果节点实现了 NodeCompiler 接口）
	if err := c.compileNodes(bp); err != nil {
		return fmt.Errorf("node compilation failed: %w", err)
	}

	// 6. 拓扑排序
	executionOrder, err := c.topologicalSort(bp)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}
	bp.executionOrder = executionOrder

	// 7. 标记为已编译
	bp.compiled = true

	return nil
}

// attachExecutors 为每个节点关联执行器
func (c *Compiler) attachExecutors(bp *Blueprint) error {
	for _, node := range bp.Nodes {
		executor, err := c.nodeRegistry.GetExecutor(node.Type, node.Operation)
		if err != nil {
			return fmt.Errorf("node %s: %w", node.ID, err)
		}

		node.executor = executor

		// 验证节点配置
		if err := executor.Validate(node); err != nil {
			return fmt.Errorf("node %s validation failed: %w", node.ID, err)
		}
	}
	return nil
}

// topologicalSort 对节点进行拓扑排序
// 使用 Kahn 算法进行拓扑排序，确保依赖关系正确
func (c *Compiler) topologicalSort(bp *Blueprint) ([]*Node, error) {
	// 构建入度表和邻接表
	inDegree := make(map[string]int)
	adjacency := make(map[string][]*Node)

	// 初始化所有节点的入度为 0
	for _, node := range bp.Nodes {
		inDegree[node.ID] = 0
	}

	// 根据连接计算入度和构建邻接表
	for _, conn := range bp.Connections {
		// 增加目标节点的入度
		inDegree[conn.TargetNode]++

		// 添加到邻接表
		sourceNode := bp.nodeMap[conn.SourceNode]
		targetNode := bp.nodeMap[conn.TargetNode]
		adjacency[sourceNode.ID] = append(adjacency[sourceNode.ID], targetNode)
	}

	// 使用队列进行拓扑排序
	queue := make([]*Node, 0)
	result := make([]*Node, 0)

	// 将所有入度为 0 的节点加入队列
	for _, node := range bp.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node)
		}
	}

	// Kahn 算法
	for len(queue) > 0 {
		// 取出队首节点
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 遍历所有邻接节点
		for _, neighbor := range adjacency[current.ID] {
			inDegree[neighbor.ID]--
			if inDegree[neighbor.ID] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// 如果结果数量不等于节点数量，说明存在环
	if len(result) != len(bp.Nodes) {
		return nil, fmt.Errorf("circular dependency detected, cannot compile blueprint")
	}

	return result, nil
}

// buildConnectionMap 构建连接映射，用于快速查找节点的输入连接
// 映射格式: targetNodeID -> []Connection
func (c *Compiler) buildConnectionMap(bp *Blueprint) map[string][]Connection {
	connMap := make(map[string][]Connection)

	for _, conn := range bp.Connections {
		connMap[conn.TargetNode] = append(connMap[conn.TargetNode], conn)
	}

	return connMap
}

// compileNodes 调用每个节点的编译方法（如果实现了 NodeCompiler 接口）
func (c *Compiler) compileNodes(bp *Blueprint) error {
	for _, node := range bp.Nodes {
		// 检查节点执行器是否实现了 NodeCompiler 接口
		if compiler, ok := node.executor.(NodeCompiler); ok {
			// 构建已连接的输入引脚集合
			connectedInputs := make(map[string]bool)
			if connections, exists := bp.connectionMap[node.ID]; exists {
				for _, conn := range connections {
					connectedInputs[conn.TargetPin] = true
				}
			}

			// 调用节点的编译方法
			if err := compiler.Compile(node, connectedInputs); err != nil {
				return fmt.Errorf("node %s (%s): %w", node.ID, node.Label, err)
			}
		}
	}
	return nil
}

// NodeRegistry 节点注册表，管理所有节点类型和执行器
type NodeRegistry struct {
	executors map[string]NodeExecutor // key: "type:operation" or "type"
	mu        sync.RWMutex
}

// NewNodeRegistry 创建一个新的节点注册表
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		executors: make(map[string]NodeExecutor),
	}
}

// Register 注册节点执行器
// 如果 operation 为空，则注册为该类型的默认执行器
func (nr *NodeRegistry) Register(nodeType NodeType, operation string, executor NodeExecutor) {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	var key string
	if operation == "" {
		key = string(nodeType)
	} else {
		key = fmt.Sprintf("%s:%s", nodeType, operation)
	}

	nr.executors[key] = executor
}

// GetExecutor 获取节点执行器
func (nr *NodeRegistry) GetExecutor(nodeType NodeType, operation string) (NodeExecutor, error) {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	// 先尝试精确匹配 type:operation
	if operation != "" {
		key := fmt.Sprintf("%s:%s", nodeType, operation)
		if executor, exists := nr.executors[key]; exists {
			return executor, nil
		}
	}

	// 再尝试匹配类型默认执行器
	key := string(nodeType)
	if executor, exists := nr.executors[key]; exists {
		return executor, nil
	}

	return nil, fmt.Errorf("no executor found for node type %s (operation: %s)", nodeType, operation)
}

// ListRegistered 列出所有已注册的执行器
func (nr *NodeRegistry) ListRegistered() []string {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	keys := make([]string, 0, len(nr.executors))
	for key := range nr.executors {
		keys = append(keys, key)
	}
	return keys
}
