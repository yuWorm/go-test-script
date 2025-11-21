package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// DataConstantExecutor 纯数据常量节点执行器（无执行引脚）
type DataConstantExecutor struct{}

// NewDataConstantExecutor 创建纯数据常量节点执行器
func NewDataConstantExecutor() *DataConstantExecutor {
	return &DataConstantExecutor{}
}

// Execute 执行常量节点（直接输出配置的值）
func (e *DataConstantExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	value, exists := inputs["value"]
	if !exists {
		value = 0 // 默认值
	}

	return map[string]interface{}{
		"output": value,
	}, nil
}

// Validate 验证节点配置
func (e *DataConstantExecutor) Validate(node *blueprint.Node) error {
	if len(node.OutputPins) < 1 {
		return fmt.Errorf("data constant node requires at least 1 output pin")
	}
	return nil
}

// DataGetVariableExecutor 纯数据获取变量节点执行器（无执行引脚）
type DataGetVariableExecutor struct{}

// NewDataGetVariableExecutor 创建纯数据获取变量节点执行器
func NewDataGetVariableExecutor() *DataGetVariableExecutor {
	return &DataGetVariableExecutor{}
}

// Execute 执行获取变量
func (e *DataGetVariableExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	varName, ok := inputs["name"].(string)
	if !ok {
		return nil, fmt.Errorf("variable name must be a string")
	}

	value, exists := ctx.GetVariable(varName)
	if !exists {
		// 变量不存在时返回 nil，而不是报错
		return map[string]interface{}{
			"value": nil,
		}, nil
	}

	return map[string]interface{}{
		"value": value,
	}, nil
}

// Validate 验证节点配置
func (e *DataGetVariableExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MakeValueExecutor 创建值节点执行器
type MakeValueExecutor struct {
	valueType string
}

// NewMakeValueExecutor 创建值节点执行器
func NewMakeValueExecutor(valueType string) *MakeValueExecutor {
	return &MakeValueExecutor{valueType: valueType}
}

// Execute 执行创建值
func (e *MakeValueExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	value := inputs["value"]

	// 根据类型进行转换
	switch e.valueType {
	case "float":
		f, err := toFloat64(value)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"output": f}, nil
	case "bool":
		b, err := toBool(value)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"output": b}, nil
	case "string":
		return map[string]interface{}{"output": fmt.Sprintf("%v", value)}, nil
	default:
		return map[string]interface{}{"output": value}, nil
	}
}

// Validate 验证节点配置
func (e *MakeValueExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// PrintExecutor 打印节点执行器
type PrintExecutor struct{}

// NewPrintExecutor 创建打印节点执行器
func NewPrintExecutor() *PrintExecutor {
	return &PrintExecutor{}
}

// Execute 执行打印
func (e *PrintExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	value := inputs["value"]
	label := ""
	if l, ok := inputs["label"].(string); ok {
		label = l
	}

	if label != "" {
		fmt.Printf("[%s] %v\n", label, value)
	} else {
		fmt.Printf("%v\n", value)
	}

	return map[string]interface{}{}, nil
}

// Validate 验证节点配置
func (e *PrintExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// DelayExecutor 延时节点执行器
type DelayExecutor struct{}

// NewDelayExecutor 创建延时节点执行器
func NewDelayExecutor() *DelayExecutor {
	return &DelayExecutor{}
}

// Execute 执行延时
func (e *DelayExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	duration, err := toFloat64(inputs["duration"])
	if err != nil {
		duration = 1000 // 默认1秒
	}

	// 模拟延时（实际实现可能需要异步处理）
	// time.Sleep(time.Duration(duration) * time.Millisecond)

	return map[string]interface{}{
		"duration": duration,
	}, nil
}

// Validate 验证节点配置
func (e *DelayExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// SequenceExecutor 序列节点执行器
type SequenceExecutor struct{}

// NewSequenceExecutor 创建序列节点执行器
func NewSequenceExecutor() *SequenceExecutor {
	return &SequenceExecutor{}
}

// Execute 执行序列（按顺序激活所有输出执行引脚）
func (e *SequenceExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 序列节点会按顺序激活所有输出执行引脚
	return map[string]interface{}{
		"then_0": true,
		"then_1": true,
		"then_2": true,
	}, nil
}

// Validate 验证节点配置
func (e *SequenceExecutor) Validate(node *blueprint.Node) error {
	return nil
}
