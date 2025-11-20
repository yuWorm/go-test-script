package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// LogicExecutor 逻辑运算执行器
type LogicExecutor struct {
	operation string
}

// NewLogicExecutor 创建逻辑运算执行器
func NewLogicExecutor(operation string) *LogicExecutor {
	return &LogicExecutor{operation: operation}
}

// Execute 执行逻辑运算
func (e *LogicExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	switch e.operation {
	case "and":
		a, err := toBool(inputs["a"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'a': %w", err)
		}
		b, err := toBool(inputs["b"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'b': %w", err)
		}
		return map[string]interface{}{"result": a && b}, nil

	case "or":
		a, err := toBool(inputs["a"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'a': %w", err)
		}
		b, err := toBool(inputs["b"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'b': %w", err)
		}
		return map[string]interface{}{"result": a || b}, nil

	case "not":
		a, err := toBool(inputs["a"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'a': %w", err)
		}
		return map[string]interface{}{"result": !a}, nil

	case "xor":
		a, err := toBool(inputs["a"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'a': %w", err)
		}
		b, err := toBool(inputs["b"])
		if err != nil {
			return nil, fmt.Errorf("invalid input 'b': %w", err)
		}
		return map[string]interface{}{"result": a != b}, nil

	default:
		return nil, fmt.Errorf("unknown logic operation: %s", e.operation)
	}
}

// Compile 在编译时对未连接的输入引脚进行类型转换
func (e *LogicExecutor) Compile(node *blueprint.Node, connectedInputs map[string]bool) error {
	// 遍历所有输入引脚
	for i := range node.InputPins {
		pin := &node.InputPins[i]

		// 跳过执行引脚和错误引脚
		if pin.Kind == blueprint.PinKindExecution || pin.Kind == blueprint.PinKindError {
			continue
		}

		// 只处理未连接且有默认值的引脚
		if !connectedInputs[pin.Name] && pin.Value != nil {
			// 将值转换为 bool 类型
			boolVal, err := toBool(pin.Value)
			if err != nil {
				return fmt.Errorf("failed to convert pin '%s' value to bool: %w", pin.Name, err)
			}
			pin.Value = boolVal
		}
	}

	return nil
}

// Validate 验证节点配置
func (e *LogicExecutor) Validate(node *blueprint.Node) error {
	if e.operation == "not" {
		if len(node.InputPins) < 1 {
			return fmt.Errorf("logic NOT node requires at least 1 input pin")
		}
	} else {
		if len(node.InputPins) < 2 {
			return fmt.Errorf("logic node requires at least 2 input pins")
		}
	}

	if len(node.OutputPins) < 1 {
		return fmt.Errorf("logic node requires at least 1 output pin")
	}

	return nil
}

// toBool 转换为布尔值
func toBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case int, int32, int64, uint, uint32, uint64:
		return val != 0, nil
	case float32, float64:
		return val != 0.0, nil
	case string:
		return val != "", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}
