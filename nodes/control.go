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

	value, exists := ctx.GetVariableFast(varName)
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

	ctx.SetVariableFast(varName, value)

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
	// 开始节点将输入引脚的值传递到输出引脚
	// 输入值来自 InputPins 的默认值或执行时传入的参数

	outputs := make(map[string]interface{})
	outputs["exec"] = true // 执行引脚始终激活

	// 将所有输入值传递到输出（除了exec_in）
	for key, value := range inputs {
		if key != "exec_in" {
			outputs[key] = value
		}
	}

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

// Execute 执行结束节点（类似 UE5 Return 节点）
func (e *EndExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})

	// 保留所有输入数据作为返回值
	for key, value := range inputs {
		// 跳过执行引脚
		if key != "exec_in" {
			outputs[key] = value
		}
	}

	// 添加执行完成标记
	outputs["executed"] = true

	// 设置返回状态，终止后续节点执行（类似 return 语句）
	ctx.SetReturned(outputs)

	return outputs, nil
}

// Validate 验证节点配置
func (e *EndExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// SwitchExecutor Switch 节点执行器（根据输入值选择输出分支）
type SwitchExecutor struct{}

// NewSwitchExecutor 创建 Switch 执行器
func NewSwitchExecutor() *SwitchExecutor {
	return &SwitchExecutor{}
}

// Execute 执行 Switch 节点
func (e *SwitchExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取选择值
	selection, exists := inputs["selection"]
	if !exists {
		return nil, fmt.Errorf("selection input is required")
	}

	outputs := make(map[string]interface{})

	// 根据类型转换选择值
	switch v := selection.(type) {
	case int, int64, int32, float64, float32:
		// 数值类型，标记匹配的输出引脚
		intVal := toInt(v)
		outputs["selected_index"] = intVal
	case string:
		// 字符串类型，标记匹配的输出引脚
		outputs["selected_value"] = v
	default:
		return nil, fmt.Errorf("unsupported selection type: %T", selection)
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *SwitchExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ForEachExecutor ForEach 节点执行器（遍历数组）
type ForEachExecutor struct{}

// NewForEachExecutor 创建 ForEach 执行器
func NewForEachExecutor() *ForEachExecutor {
	return &ForEachExecutor{}
}

// Execute 执行 ForEach 节点
func (e *ForEachExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取数组
	arrayInput, exists := inputs["array"]
	if !exists {
		return nil, fmt.Errorf("array input is required")
	}

	// 转换为数组
	var array []interface{}
	switch v := arrayInput.(type) {
	case []interface{}:
		array = v
	case []int:
		array = make([]interface{}, len(v))
		for i, val := range v {
			array[i] = val
		}
	case []string:
		array = make([]interface{}, len(v))
		for i, val := range v {
			array[i] = val
		}
	case []float64:
		array = make([]interface{}, len(v))
		for i, val := range v {
			array[i] = val
		}
	default:
		return nil, fmt.Errorf("array input must be an array, got %T", arrayInput)
	}

	// ForEach 的实际循环逻辑由执行器框架处理
	// 这里只返回数组信息
	return map[string]interface{}{
		"array": array,
		"count": len(array),
	}, nil
}

// Validate 验证节点配置
func (e *ForEachExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// 辅助函数：将数值转换为 int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		return 0
	}
}

// FunctionCallExecutor 函数调用执行器
type FunctionCallExecutor struct {
	functionName     string
	functionRegistry *blueprint.FunctionRegistry
}

// NewFunctionCallExecutor 创建函数调用执行器
func NewFunctionCallExecutor(functionName string, registry *blueprint.FunctionRegistry) *FunctionCallExecutor {
	return &FunctionCallExecutor{
		functionName:     functionName,
		functionRegistry: registry,
	}
}

// Execute 执行函数调用
func (e *FunctionCallExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取函数蓝图
	functionBP, err := e.functionRegistry.GetFunction(e.functionName)
	if err != nil {
		return nil, fmt.Errorf("function %s not found: %w", e.functionName, err)
	}

	// 创建新的执行器来执行函数蓝图
	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())

	// 将输入参数传递给函数（通过变量映射）
	functionInputs := make(map[string]interface{})
	for key, value := range inputs {
		// 跳过执行引脚
		if key != "exec_in" {
			functionInputs[key] = value
		}
	}

	// 执行函数蓝图
	result, err := executor.Execute(functionBP, functionInputs)
	if err != nil {
		return nil, fmt.Errorf("function %s execution failed: %w", e.functionName, err)
	}

	// 从结果中提取返回值（从 return.* 输出中获取）
	outputs := make(map[string]interface{})
	for key, value := range result.Outputs {
		// 提取 return.* 输出作为函数返回值
		if len(key) > 7 && key[:7] == "return." {
			outputName := key[7:]
			outputs[outputName] = value
		}
	}

	// 标记函数执行完成
	outputs["executed"] = true

	return outputs, nil
}

// Validate 验证节点配置
func (e *FunctionCallExecutor) Validate(node *blueprint.Node) error {
	// 验证函数是否存在
	_, err := e.functionRegistry.GetFunction(e.functionName)
	return err
}
