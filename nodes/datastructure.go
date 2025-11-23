package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// ==================== Array 节点 ====================

// MakeArrayExecutor 创建数组执行器
type MakeArrayExecutor struct{}

func NewMakeArrayExecutor() *MakeArrayExecutor {
	return &MakeArrayExecutor{}
}

func (e *MakeArrayExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 从输入获取数组元素
	var array []interface{}

	// 支持动态输入元素
	for i := 0; i < 10; i++ { // 最多支持 10 个元素
		key := fmt.Sprintf("element_%d", i)
		if val, exists := inputs[key]; exists {
			array = append(array, val)
		}
	}

	return map[string]interface{}{
		"array": array,
		"length": float64(len(array)),
	}, nil
}

func (e *MakeArrayExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ArrayGetExecutor 获取数组元素执行器
type ArrayGetExecutor struct{}

func NewArrayGetExecutor() *ArrayGetExecutor {
	return &ArrayGetExecutor{}
}

func (e *ArrayGetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	arrayInput, exists := inputs["array"]
	if !exists {
		return nil, fmt.Errorf("array input is required")
	}

	indexInput, exists := inputs["index"]
	if !exists {
		return nil, fmt.Errorf("index input is required")
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
		return nil, fmt.Errorf("array must be an array type, got %T", arrayInput)
	}

	// 转换索引
	indexFloat, err := toFloat64(indexInput)
	if err != nil {
		return nil, fmt.Errorf("invalid index: %w", err)
	}
	index := int(indexFloat)
	if index < 0 || index >= len(array) {
		return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index, len(array))
	}

	return map[string]interface{}{
		"element": array[index],
		"found":   true,
	}, nil
}

func (e *ArrayGetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ArraySetExecutor 设置数组元素执行器
type ArraySetExecutor struct{}

func NewArraySetExecutor() *ArraySetExecutor {
	return &ArraySetExecutor{}
}

func (e *ArraySetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	arrayInput, exists := inputs["array"]
	if !exists {
		return nil, fmt.Errorf("array input is required")
	}

	indexInput, exists := inputs["index"]
	if !exists {
		return nil, fmt.Errorf("index input is required")
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value input is required")
	}

	// 转换为数组
	var array []interface{}
	switch v := arrayInput.(type) {
	case []interface{}:
		array = make([]interface{}, len(v))
		copy(array, v)
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
		return nil, fmt.Errorf("array must be an array type, got %T", arrayInput)
	}

	// 转换索引
	indexFloat, err := toFloat64(indexInput)
	if err != nil {
		return nil, fmt.Errorf("invalid index: %w", err)
	}
	index := int(indexFloat)
	if index < 0 || index >= len(array) {
		return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index, len(array))
	}

	array[index] = value

	return map[string]interface{}{
		"array": array,
	}, nil
}

func (e *ArraySetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ArrayAddExecutor 添加数组元素执行器
type ArrayAddExecutor struct{}

func NewArrayAddExecutor() *ArrayAddExecutor {
	return &ArrayAddExecutor{}
}

func (e *ArrayAddExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	arrayInput, exists := inputs["array"]
	if !exists {
		return nil, fmt.Errorf("array input is required")
	}

	element, exists := inputs["element"]
	if !exists {
		return nil, fmt.Errorf("element input is required")
	}

	// 转换为数组
	var array []interface{}
	switch v := arrayInput.(type) {
	case []interface{}:
		array = make([]interface{}, len(v))
		copy(array, v)
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
		return nil, fmt.Errorf("array must be an array type, got %T", arrayInput)
	}

	array = append(array, element)

	return map[string]interface{}{
		"array":  array,
		"length": float64(len(array)),
	}, nil
}

func (e *ArrayAddExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ArrayLengthExecutor 获取数组长度执行器
type ArrayLengthExecutor struct{}

func NewArrayLengthExecutor() *ArrayLengthExecutor {
	return &ArrayLengthExecutor{}
}

func (e *ArrayLengthExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	arrayInput, exists := inputs["array"]
	if !exists {
		return nil, fmt.Errorf("array input is required")
	}

	var length int
	switch v := arrayInput.(type) {
	case []interface{}:
		length = len(v)
	case []int:
		length = len(v)
	case []string:
		length = len(v)
	case []float64:
		length = len(v)
	default:
		return nil, fmt.Errorf("array must be an array type, got %T", arrayInput)
	}

	return map[string]interface{}{
		"length": float64(length),
	}, nil
}

func (e *ArrayLengthExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ==================== Map 节点 ====================

// MakeMapExecutor 创建 Map 执行器
type MakeMapExecutor struct{}

func NewMakeMapExecutor() *MakeMapExecutor {
	return &MakeMapExecutor{}
}

func (e *MakeMapExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	// 创建空 map
	m := make(map[string]interface{})

	// 支持初始键值对输入
	for i := 0; i < 10; i++ { // 最多支持 10 个键值对
		keyName := fmt.Sprintf("key_%d", i)
		valName := fmt.Sprintf("value_%d", i)

		if key, keyExists := inputs[keyName]; keyExists {
			if val, valExists := inputs[valName]; valExists {
				keyStr := fmt.Sprintf("%v", key)
				m[keyStr] = val
			}
		}
	}

	return map[string]interface{}{
		"map":  m,
		"size": float64(len(m)),
	}, nil
}

func (e *MakeMapExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MapGetExecutor 获取 Map 值执行器
type MapGetExecutor struct{}

func NewMapGetExecutor() *MapGetExecutor {
	return &MapGetExecutor{}
}

func (e *MapGetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	mapInput, exists := inputs["map"]
	if !exists {
		return nil, fmt.Errorf("map input is required")
	}

	keyInput, exists := inputs["key"]
	if !exists {
		return nil, fmt.Errorf("key input is required")
	}

	m, ok := mapInput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("map must be a map type, got %T", mapInput)
	}

	key := fmt.Sprintf("%v", keyInput)
	value, found := m[key]

	return map[string]interface{}{
		"value": value,
		"found": found,
	}, nil
}

func (e *MapGetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MapSetExecutor 设置 Map 值执行器
type MapSetExecutor struct{}

func NewMapSetExecutor() *MapSetExecutor {
	return &MapSetExecutor{}
}

func (e *MapSetExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	mapInput, exists := inputs["map"]
	if !exists {
		return nil, fmt.Errorf("map input is required")
	}

	keyInput, exists := inputs["key"]
	if !exists {
		return nil, fmt.Errorf("key input is required")
	}

	value, exists := inputs["value"]
	if !exists {
		return nil, fmt.Errorf("value input is required")
	}

	// 复制 map 避免修改原始数据
	origMap, ok := mapInput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("map must be a map type, got %T", mapInput)
	}

	m := make(map[string]interface{})
	for k, v := range origMap {
		m[k] = v
	}

	key := fmt.Sprintf("%v", keyInput)
	m[key] = value

	return map[string]interface{}{
		"map": m,
	}, nil
}

func (e *MapSetExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MapRemoveExecutor 删除 Map 键执行器
type MapRemoveExecutor struct{}

func NewMapRemoveExecutor() *MapRemoveExecutor {
	return &MapRemoveExecutor{}
}

func (e *MapRemoveExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	mapInput, exists := inputs["map"]
	if !exists {
		return nil, fmt.Errorf("map input is required")
	}

	keyInput, exists := inputs["key"]
	if !exists {
		return nil, fmt.Errorf("key input is required")
	}

	// 复制 map 避免修改原始数据
	origMap, ok := mapInput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("map must be a map type, got %T", mapInput)
	}

	m := make(map[string]interface{})
	for k, v := range origMap {
		m[k] = v
	}

	key := fmt.Sprintf("%v", keyInput)
	delete(m, key)

	return map[string]interface{}{
		"map": m,
	}, nil
}

func (e *MapRemoveExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MapKeysExecutor 获取 Map 所有键执行器
type MapKeysExecutor struct{}

func NewMapKeysExecutor() *MapKeysExecutor {
	return &MapKeysExecutor{}
}

func (e *MapKeysExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	mapInput, exists := inputs["map"]
	if !exists {
		return nil, fmt.Errorf("map input is required")
	}

	m, ok := mapInput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("map must be a map type, got %T", mapInput)
	}

	keys := make([]interface{}, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return map[string]interface{}{
		"keys":  keys,
		"count": float64(len(keys)),
	}, nil
}

func (e *MapKeysExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// MapSizeExecutor 获取 Map 大小执行器
type MapSizeExecutor struct{}

func NewMapSizeExecutor() *MapSizeExecutor {
	return &MapSizeExecutor{}
}

func (e *MapSizeExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	mapInput, exists := inputs["map"]
	if !exists {
		return nil, fmt.Errorf("map input is required")
	}

	m, ok := mapInput.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("map must be a map type, got %T", mapInput)
	}

	return map[string]interface{}{
		"size": float64(len(m)),
	}, nil
}

func (e *MapSizeExecutor) Validate(node *blueprint.Node) error {
	return nil
}
