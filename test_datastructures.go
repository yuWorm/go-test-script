package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 测试数据结构：Array 和 Map ===\n")

	// 测试 1: Array 基本操作
	testArrayBasic()

	// 测试 2: Map 基本操作
	testMapBasic()

	// 测试 3: Array + ForEach 组合
	testArrayWithForEach()
}

// 测试 Array 基本操作
func testArrayBasic() {
	fmt.Println("--- 测试 1: Array 基本操作 ---")

	bp := blueprint.NewBlueprint("test_array_basic")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// Make Array 节点
	makeArray := &blueprint.Node{
		ID:        "make_array",
		Type:      blueprint.NodeTypeData,
		Operation: "make_array",
		Label:     "Make Array",
		InputPins: []blueprint.Pin{
			{Name: "element_0", Kind: blueprint.PinKindData, Type: "any", Value: 10.0},
			{Name: "element_1", Kind: blueprint.PinKindData, Type: "any", Value: 20.0},
			{Name: "element_2", Kind: blueprint.PinKindData, Type: "any", Value: 30.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "length", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Array Get 节点
	arrayGet := &blueprint.Node{
		ID:        "array_get",
		Type:      blueprint.NodeTypeData,
		Operation: "array_get",
		Label:     "Array Get",
		InputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "index", Kind: blueprint.PinKindData, Type: "float", Value: 1.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "element", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "found", Kind: blueprint.PinKindData, Type: "bool"},
		},
	}

	// Array Add 节点
	arrayAdd := &blueprint.Node{
		ID:        "array_add",
		Type:      blueprint.NodeTypeData,
		Operation: "array_add",
		Label:     "Array Add",
		InputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "element", Kind: blueprint.PinKindData, Type: "any", Value: 40.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "length", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Print 节点 - 打印获取的元素
	print1 := &blueprint.Node{
		ID:        "print1",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Element",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Array[1]"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点 - 打印数组长度
	print2 := &blueprint.Node{
		ID:        "print2",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Length",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Array Length"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(makeArray)
	bp.AddNode(arrayGet)
	bp.AddNode(arrayAdd)
	bp.AddNode(print1)
	bp.AddNode(print2)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "print1", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "make_array", SourcePin: "array", TargetNode: "array_get", TargetPin: "array"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "array_get", SourcePin: "element", TargetNode: "print1", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "print1", SourcePin: "exec_out", TargetNode: "print2", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "make_array", SourcePin: "array", TargetNode: "array_add", TargetPin: "array"})
	bp.AddConnection(blueprint.Connection{ID: "c6", SourceNode: "array_add", SourcePin: "length", TargetNode: "print2", TargetPin: "value"})

	// 编译并执行
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		fmt.Printf("编译失败: %v\n", err)
		return
	}

	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(bp, nil)
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		return
	}

	fmt.Printf("执行成功: %v\n\n", result.Success)
}

// 测试 Map 基本操作
func testMapBasic() {
	fmt.Println("--- 测试 2: Map 基本操作 ---")

	bp := blueprint.NewBlueprint("test_map_basic")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// Make Map 节点
	makeMap := &blueprint.Node{
		ID:        "make_map",
		Type:      blueprint.NodeTypeData,
		Operation: "make_map",
		Label:     "Make Map",
		InputPins: []blueprint.Pin{
			{Name: "key_0", Kind: blueprint.PinKindData, Type: "string", Value: "name"},
			{Name: "value_0", Kind: blueprint.PinKindData, Type: "any", Value: "Alice"},
			{Name: "key_1", Kind: blueprint.PinKindData, Type: "string", Value: "age"},
			{Name: "value_1", Kind: blueprint.PinKindData, Type: "any", Value: 25.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "map", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "size", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Map Get 节点
	mapGet := &blueprint.Node{
		ID:        "map_get",
		Type:      blueprint.NodeTypeData,
		Operation: "map_get",
		Label:     "Map Get",
		InputPins: []blueprint.Pin{
			{Name: "map", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "key", Kind: blueprint.PinKindData, Type: "string", Value: "name"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "found", Kind: blueprint.PinKindData, Type: "bool"},
		},
	}

	// Map Set 节点
	mapSet := &blueprint.Node{
		ID:        "map_set",
		Type:      blueprint.NodeTypeData,
		Operation: "map_set",
		Label:     "Map Set",
		InputPins: []blueprint.Pin{
			{Name: "map", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "key", Kind: blueprint.PinKindData, Type: "string", Value: "city"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any", Value: "Beijing"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "map", Kind: blueprint.PinKindData, Type: "any"},
		},
	}

	// Map Size 节点
	mapSize := &blueprint.Node{
		ID:        "map_size",
		Type:      blueprint.NodeTypeData,
		Operation: "map_size",
		Label:     "Map Size",
		InputPins: []blueprint.Pin{
			{Name: "map", Kind: blueprint.PinKindData, Type: "any"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "size", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Print 节点 - 打印获取的值
	print1 := &blueprint.Node{
		ID:        "print1",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Name",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Name"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点 - 打印 Map 大小
	print2 := &blueprint.Node{
		ID:        "print2",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Size",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Map Size"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(makeMap)
	bp.AddNode(mapGet)
	bp.AddNode(mapSet)
	bp.AddNode(mapSize)
	bp.AddNode(print1)
	bp.AddNode(print2)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "print1", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "make_map", SourcePin: "map", TargetNode: "map_get", TargetPin: "map"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "map_get", SourcePin: "value", TargetNode: "print1", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "print1", SourcePin: "exec_out", TargetNode: "print2", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "make_map", SourcePin: "map", TargetNode: "map_set", TargetPin: "map"})
	bp.AddConnection(blueprint.Connection{ID: "c6", SourceNode: "map_set", SourcePin: "map", TargetNode: "map_size", TargetPin: "map"})
	bp.AddConnection(blueprint.Connection{ID: "c7", SourceNode: "map_size", SourcePin: "size", TargetNode: "print2", TargetPin: "value"})

	// 编译并执行
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		fmt.Printf("编译失败: %v\n", err)
		return
	}

	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(bp, nil)
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		return
	}

	fmt.Printf("执行成功: %v\n\n", result.Success)
}

// 测试 Array + ForEach 组合
func testArrayWithForEach() {
	fmt.Println("--- 测试 3: Array + ForEach 组合 ---")

	bp := blueprint.NewBlueprint("test_array_foreach")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// Make Array 节点
	makeArray := &blueprint.Node{
		ID:        "make_array",
		Type:      blueprint.NodeTypeData,
		Operation: "make_array",
		Label:     "Make Array [1,2,3,4,5]",
		InputPins: []blueprint.Pin{
			{Name: "element_0", Kind: blueprint.PinKindData, Type: "any", Value: 1.0},
			{Name: "element_1", Kind: blueprint.PinKindData, Type: "any", Value: 2.0},
			{Name: "element_2", Kind: blueprint.PinKindData, Type: "any", Value: 3.0},
			{Name: "element_3", Kind: blueprint.PinKindData, Type: "any", Value: 4.0},
			{Name: "element_4", Kind: blueprint.PinKindData, Type: "any", Value: 5.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "length", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// ForEach 节点
	foreach := &blueprint.Node{
		ID:        "foreach",
		Type:      blueprint.NodeTypeFlowControl,
		Operation: "foreach",
		Label:     "ForEach",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "loop_body", Kind: blueprint.PinKindExecution},
			{Name: "completed", Kind: blueprint.PinKindExecution},
			{Name: "element", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "index", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Multiply 节点 - 每个元素乘以2
	multiply := &blueprint.Node{
		ID:        "multiply",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		Label:     "Multiply by 2",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 2.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Print 节点
	print := &blueprint.Node{
		ID:        "print",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Result",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Element * 2"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print Done
	printDone := &blueprint.Node{
		ID:        "print_done",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Done",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any", Value: "Array processing completed"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Status"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(makeArray)
	bp.AddNode(foreach)
	bp.AddNode(multiply)
	bp.AddNode(print)
	bp.AddNode(printDone)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "foreach", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "make_array", SourcePin: "array", TargetNode: "foreach", TargetPin: "array"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "foreach", SourcePin: "loop_body", TargetNode: "multiply", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "foreach", SourcePin: "element", TargetNode: "multiply", TargetPin: "a"})
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "multiply", SourcePin: "exec_out", TargetNode: "print", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c6", SourceNode: "multiply", SourcePin: "result", TargetNode: "print", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c7", SourceNode: "foreach", SourcePin: "completed", TargetNode: "print_done", TargetPin: "exec_in"})

	// 编译并执行
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		fmt.Printf("编译失败: %v\n", err)
		return
	}

	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(bp, nil)
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		return
	}

	fmt.Printf("执行成功: %v\n", result.Success)

	// 打印部分结果
	resultJSON, _ := json.MarshalIndent(map[string]interface{}{
		"array_length": result.Outputs["make_array.length"],
	}, "", "  ")
	fmt.Printf("结果: %s\n", string(resultJSON))
}
