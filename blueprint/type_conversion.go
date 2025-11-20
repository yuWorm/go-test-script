package blueprint

import (
	"fmt"
	"strconv"
)

// ConvertValue 将值转换为指定类型
// 在编译时用于转换引脚的默认值
func ConvertValue(value interface{}, targetType string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 如果已经是目标类型，直接返回
	switch targetType {
	case "float", "number":
		return toFloat64(value)
	case "int":
		return toInt(value)
	case "string":
		return toString(value)
	case "bool", "boolean":
		return toBool(value)
	case "any":
		return value, nil
	default:
		// 未知类型，保持原样
		return value, nil
	}
}

// toFloat64 转换为浮点数
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to float: %w", v, err)
		}
		return f, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert type %T to float", value)
	}
}

// toInt 转换为整数
func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to int: %w", v, err)
		}
		return i, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert type %T to int", value)
	}
}

// toString 转换为字符串
func toString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

// toBool 转换为布尔值
func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int32, int64:
		return v != 0, nil
	case float32, float64:
		return v != 0.0, nil
	case string:
		// 支持多种字符串格式
		switch v {
		case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
			return true, nil
		case "false", "False", "FALSE", "0", "no", "No", "NO", "":
			return false, nil
		default:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return false, fmt.Errorf("cannot convert string '%s' to bool: %w", v, err)
			}
			return b, nil
		}
	default:
		return false, fmt.Errorf("cannot convert type %T to bool", value)
	}
}

// ConvertPinValue 转换引脚的值到目标类型
// 如果转换失败，返回错误
func ConvertPinValue(pin *Pin) error {
	if pin.Value == nil {
		return nil
	}

	// 如果引脚类型是exec或error，不需要转换
	if pin.Kind == PinKindExecution || pin.Kind == PinKindError {
		return nil
	}

	// 转换值
	converted, err := ConvertValue(pin.Value, pin.Type)
	if err != nil {
		return fmt.Errorf("pin '%s': %w", pin.Name, err)
	}

	pin.Value = converted
	return nil
}
