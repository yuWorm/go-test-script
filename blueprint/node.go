package blueprint

import (
	"fmt"
	"sync"
)

// NodeType 定义节点类型
type NodeType string

const (
	NodeTypeStart      NodeType = "start"
	NodeTypeEnd        NodeType = "end"
	NodeTypeArithmetic NodeType = "arithmetic"
	NodeTypeLogic      NodeType = "logic"
	NodeTypeCondition  NodeType = "condition"
	NodeTypeFunction   NodeType = "function"
)

// Pin 表示节点的输入或输出引脚
type Pin struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`  // "int", "float", "string", "bool", "any"
	Value interface{} `json:"value"` // 默认值或实际值
}

// Node 表示蓝图中的一个节点
type Node struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Operation   string                 `json:"operation,omitempty"` // 具体操作，如 "add", "subtract"
	Label       string                 `json:"label"`
	InputPins   []Pin                  `json:"input_pins"`
	OutputPins  []Pin                  `json:"output_pins"`
	Properties  map[string]interface{} `json:"properties,omitempty"` // 节点属性
	Position    Position               `json:"position"`             // 用于可视化编辑器
	mu          sync.RWMutex           // 用于并发安全
	executor    NodeExecutor           // 节点执行器（编译时设置）
	inputCache  map[string]interface{} // 输入缓存（运行时）
	outputCache map[string]interface{} // 输出缓存（运行时）
}

// Position 表示节点在可视化编辑器中的位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Connection 表示节点之间的连接
type Connection struct {
	ID         string `json:"id"`
	SourceNode string `json:"source_node"`
	SourcePin  string `json:"source_pin"`
	TargetNode string `json:"target_node"`
	TargetPin  string `json:"target_pin"`
}

// NodeExecutor 定义节点执行器接口
type NodeExecutor interface {
	// Execute 执行节点逻辑
	// inputs: 输入引脚值映射
	// 返回: 输出引脚值映射和错误
	Execute(ctx *ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error)

	// Validate 验证节点配置是否有效
	Validate(node *Node) error
}

// GetInputValue 并发安全地获取输入值
func (n *Node) GetInputValue(pinName string) (interface{}, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.inputCache != nil {
		val, ok := n.inputCache[pinName]
		return val, ok
	}

	// 如果缓存为空，查找默认值
	for _, pin := range n.InputPins {
		if pin.Name == pinName {
			return pin.Value, pin.Value != nil
		}
	}
	return nil, false
}

// SetInputValue 并发安全地设置输入值
func (n *Node) SetInputValue(pinName string, value interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.inputCache == nil {
		n.inputCache = make(map[string]interface{})
	}
	n.inputCache[pinName] = value
}

// GetOutputValue 并发安全地获取输出值
func (n *Node) GetOutputValue(pinName string) (interface{}, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.outputCache != nil {
		val, ok := n.outputCache[pinName]
		return val, ok
	}
	return nil, false
}

// SetOutputValue 并发安全地设置输出值
func (n *Node) SetOutputValue(pinName string, value interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.outputCache == nil {
		n.outputCache = make(map[string]interface{})
	}
	n.outputCache[pinName] = value
}

// ResetCache 重置节点缓存（用于重新执行）
func (n *Node) ResetCache() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.inputCache = make(map[string]interface{})
	n.outputCache = make(map[string]interface{})
}

// Validate 验证节点配置
func (n *Node) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if n.Type == "" {
		return fmt.Errorf("node %s: type cannot be empty", n.ID)
	}

	// 检查引脚名称唯一性
	pinNames := make(map[string]bool)
	for _, pin := range n.InputPins {
		if pin.Name == "" {
			return fmt.Errorf("node %s: input pin name cannot be empty", n.ID)
		}
		if pinNames[pin.Name] {
			return fmt.Errorf("node %s: duplicate pin name %s", n.ID, pin.Name)
		}
		pinNames[pin.Name] = true
	}

	for _, pin := range n.OutputPins {
		if pin.Name == "" {
			return fmt.Errorf("node %s: output pin name cannot be empty", n.ID)
		}
		if pinNames[pin.Name] {
			return fmt.Errorf("node %s: duplicate pin name %s", n.ID, pin.Name)
		}
		pinNames[pin.Name] = true
	}

	// 如果有执行器，调用其验证方法
	if n.executor != nil {
		return n.executor.Validate(n)
	}

	return nil
}
