package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// ArithmeticExecutor 算术运算执行器
type ArithmeticExecutor struct {
	operation string
}

// NewArithmeticExecutor 创建算术运算执行器
func NewArithmeticExecutor(operation string) *ArithmeticExecutor {
	return &ArithmeticExecutor{operation: operation}
}

// Execute 执行算术运算
func (e *ArithmeticExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 获取输入值
	a, err := toFloat64(inputs["a"])
	if err != nil {
		return nil, fmt.Errorf("invalid input 'a': %w", err)
	}

	b, err := toFloat64(inputs["b"])
	if err != nil {
		return nil, fmt.Errorf("invalid input 'b': %w", err)
	}

	var result float64

	switch e.operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	case "modulo":
		if b == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		result = float64(int(a) % int(b))
	case "power":
		result = 1
		for i := 0; i < int(b); i++ {
			result *= a
		}
	default:
		return nil, fmt.Errorf("unknown arithmetic operation: %s", e.operation)
	}

	outputs := map[string]interface{}{
		"result": result,
	}

	// 如果有执行输入，则输出执行信号
	if _, hasExecIn := inputs["exec_in"]; hasExecIn {
		outputs["exec_out"] = true
	}

	return outputs, nil
}

// Compile 在编译时对未连接的输入引脚进行类型转换
func (e *ArithmeticExecutor) Compile(node *blueprint.Node, connectedInputs map[string]bool) error {
	// 遍历所有输入引脚
	for i := range node.InputPins {
		pin := &node.InputPins[i]

		// 跳过执行引脚和错误引脚
		if pin.Kind == blueprint.PinKindExecution || pin.Kind == blueprint.PinKindError {
			continue
		}

		// 只处理未连接且有默认值的引脚
		if !connectedInputs[pin.Name] && pin.Value != nil {
			// 将值转换为 float64 类型
			floatVal, err := toFloat64(pin.Value)
			if err != nil {
				return fmt.Errorf("failed to convert pin '%s' value to float: %w", pin.Name, err)
			}
			pin.Value = floatVal
		}
	}

	return nil
}

// Validate 验证节点配置
func (e *ArithmeticExecutor) Validate(node *blueprint.Node) error {
	// 检查输入引脚
	if len(node.InputPins) < 2 {
		return fmt.Errorf("arithmetic node requires at least 2 input pins")
	}

	hasA := false
	hasB := false
	for _, pin := range node.InputPins {
		if pin.Name == "a" {
			hasA = true
		}
		if pin.Name == "b" {
			hasB = true
		}
	}

	if !hasA || !hasB {
		return fmt.Errorf("arithmetic node must have 'a' and 'b' input pins")
	}

	// 检查输出引脚
	if len(node.OutputPins) < 1 {
		return fmt.Errorf("arithmetic node requires at least 1 output pin")
	}

	hasResult := false
	for _, pin := range node.OutputPins {
		if pin.Name == "result" {
			hasResult = true
		}
	}

	if !hasResult {
		return fmt.Errorf("arithmetic node must have 'result' output pin")
	}

	return nil
}

// ComparisonExecutor 比较运算执行器
type ComparisonExecutor struct {
	operation string
}

// NewComparisonExecutor 创建比较运算执行器
func NewComparisonExecutor(operation string) *ComparisonExecutor {
	return &ComparisonExecutor{operation: operation}
}

// Execute 执行比较运算
func (e *ComparisonExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	a, err := toFloat64(inputs["a"])
	if err != nil {
		return nil, fmt.Errorf("invalid input 'a': %w", err)
	}

	b, err := toFloat64(inputs["b"])
	if err != nil {
		return nil, fmt.Errorf("invalid input 'b': %w", err)
	}

	var result bool

	switch e.operation {
	case "equal":
		result = a == b
	case "not_equal":
		result = a != b
	case "greater":
		result = a > b
	case "greater_equal":
		result = a >= b
	case "less":
		result = a < b
	case "less_equal":
		result = a <= b
	default:
		return nil, fmt.Errorf("unknown comparison operation: %s", e.operation)
	}

	return map[string]interface{}{
		"result": result,
	}, nil
}

// Compile 在编译时对未连接的输入引脚进行类型转换
func (e *ComparisonExecutor) Compile(node *blueprint.Node, connectedInputs map[string]bool) error {
	// 遍历所有输入引脚
	for i := range node.InputPins {
		pin := &node.InputPins[i]

		// 跳过执行引脚和错误引脚
		if pin.Kind == blueprint.PinKindExecution || pin.Kind == blueprint.PinKindError {
			continue
		}

		// 只处理未连接且有默认值的引脚
		if !connectedInputs[pin.Name] && pin.Value != nil {
			// 将值转换为 float64 类型
			floatVal, err := toFloat64(pin.Value)
			if err != nil {
				return fmt.Errorf("failed to convert pin '%s' value to float: %w", pin.Name, err)
			}
			pin.Value = floatVal
		}
	}

	return nil
}

// Validate 验证节点配置
func (e *ComparisonExecutor) Validate(node *blueprint.Node) error {
	if len(node.InputPins) < 2 {
		return fmt.Errorf("comparison node requires at least 2 input pins")
	}
	if len(node.OutputPins) < 1 {
		return fmt.Errorf("comparison node requires at least 1 output pin")
	}
	return nil
}

// toFloat64 转换为 float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		// 尝试解析字符串为浮点数
		if f, err := parseFloat(val); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("cannot convert string '%s' to float64", val)
	case bool:
		if val {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// parseFloat 解析字符串为浮点数
func parseFloat(s string) (float64, error) {
	// 简单的浮点数解析
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
