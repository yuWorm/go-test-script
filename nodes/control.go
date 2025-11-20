package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// ConditionExecutor 条件分支执行器
type ConditionExecutor struct{}

// NewConditionExecutor 创建条件分支执行器
func NewConditionExecutor() *ConditionExecutor {
	return &ConditionExecutor{}
}

// Execute 执行条件分支
func (e *ConditionExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	condition, err := toBool(inputs["condition"])
	if err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	outputs := make(map[string]interface{})

	if condition {
		if trueValue, exists := inputs["true_value"]; exists {
			outputs["result"] = trueValue
		}
		outputs["is_true"] = true
		outputs["is_false"] = false
	} else {
		if falseValue, exists := inputs["false_value"]; exists {
			outputs["result"] = falseValue
		}
		outputs["is_true"] = false
		outputs["is_false"] = true
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *ConditionExecutor) Validate(node *blueprint.Node) error {
	hasCondition := false
	for _, pin := range node.InputPins {
		if pin.Name == "condition" {
			hasCondition = true
			break
		}
	}

	if !hasCondition {
		return fmt.Errorf("condition node must have 'condition' input pin")
	}

	return nil
}

// ConstantExecutor 常量节点执行器
type ConstantExecutor struct{}

// NewConstantExecutor 创建常量节点执行器
func NewConstantExecutor() *ConstantExecutor {
	return &ConstantExecutor{}
}

// Execute 执行常量节点（直接输出配置的值）
func (e *ConstantExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 常量节点从属性中获取值
	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("constant node requires 'value' input")
	}

	return map[string]interface{}{
		"output": value,
	}, nil
}

// Validate 验证节点配置
func (e *ConstantExecutor) Validate(node *blueprint.Node) error {
	if len(node.OutputPins) < 1 {
		return fmt.Errorf("constant node requires at least 1 output pin")
	}
	return nil
}

// VariableGetExecutor 获取变量执行器
type VariableGetExecutor struct{}

// NewVariableGetExecutor 创建获取变量执行器
func NewVariableGetExecutor() *VariableGetExecutor {
	return &VariableGetExecutor{}
}

// Execute 执行获取变量
func (e *VariableGetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	varName, ok := inputs["name"].(string)
	if !ok {
		return nil, fmt.Errorf("variable name must be a string")
	}

	value, exists := ctx.GetVariable(varName)
	if !exists {
		return nil, fmt.Errorf("variable %s not found", varName)
	}

	return map[string]interface{}{
		"value": value,
	}, nil
}

// Validate 验证节点配置
func (e *VariableGetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// VariableSetExecutor 设置变量执行器
type VariableSetExecutor struct{}

// NewVariableSetExecutor 创建设置变量执行器
func NewVariableSetExecutor() *VariableSetExecutor {
	return &VariableSetExecutor{}
}

// Execute 执行设置变量
func (e *VariableSetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	varName, ok := inputs["name"].(string)
	if !ok {
		return nil, fmt.Errorf("variable name must be a string")
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value is required")
	}

	ctx.SetVariable(varName, value)

	return map[string]interface{}{
		"value": value,
	}, nil
}

// Validate 验证节点配置
func (e *VariableSetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// StartExecutor 开始节点执行器
type StartExecutor struct{}

// NewStartExecutor 创建开始节点执行器
func NewStartExecutor() *StartExecutor {
	return &StartExecutor{}
}

// Execute 执行开始节点
func (e *StartExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 开始节点没有输入，但有输出引脚
	// 输出引脚的值来自引脚的默认值
	// 这些值已经在节点的 OutputPins 中定义
	// executeNode 会从输入获取，但对于 Start 节点，我们返回空映射
	// 实际的值会在 executeNode 中从 pin.Value 读取

	// 输出执行信号
	outputs := make(map[string]interface{})
	outputs["exec"] = true // 执行引脚始终激活

	return outputs, nil
}

// Validate 验证节点配置
func (e *StartExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// EndExecutor 结束节点执行器
type EndExecutor struct{}

// NewEndExecutor 创建结束节点执行器
func NewEndExecutor() *EndExecutor {
	return &EndExecutor{}
}

// Execute 执行结束节点
func (e *EndExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 结束节点接收输入并标记为已执行
	outputs := make(map[string]interface{})

	// 保留所有输入数据
	for key, value := range inputs {
		// 跳过执行引脚
		if key != "exec_in" {
			outputs[key] = value
		}
	}

	// 添加执行完成标记
	outputs["executed"] = true

	return outputs, nil
}

// Validate 验证节点配置
func (e *EndExecutor) Validate(node *blueprint.Node) error {
	return nil
}
