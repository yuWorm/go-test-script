package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 测试新功能：Sequence、Switch、ForEach、函数子图 ===\n")

	// 测试 1: Sequence 节点
	testSequence()

	// 测试 2: Switch 节点
	testSwitch()

	// 测试 3: ForEach 节点
	testForEach()

	// 测试 4: 函数子图
	testFunctionSubgraph()
}

// 测试 Sequence 节点
func testSequence() {
	fmt.Println("--- 测试 1: Sequence 节点 ---")

	bp := blueprint.NewBlueprint("test_sequence")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float", Value: 10.0},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 20.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Sequence 节点
	sequence := &blueprint.Node{
		ID:        "seq",
		Type:      blueprint.NodeTypeFlowControl,
		Operation: "sequence",
		Label:     "Sequence",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
		},
		OutputPins: []blueprint.Pin{
			{Name: "Then 0", Kind: blueprint.PinKindExecution},
			{Name: "Then 1", Kind: blueprint.PinKindExecution},
			{Name: "Then 2", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点 1
	print1 := &blueprint.Node{
		ID:        "print1",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print First",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "First"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点 2
	print2 := &blueprint.Node{
		ID:        "print2",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Second",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Second"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点 3
	print3 := &blueprint.Node{
		ID:        "print3",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Third",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Third"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(sequence)
	bp.AddNode(print1)
	bp.AddNode(print2)
	bp.AddNode(print3)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "seq", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "seq", SourcePin: "Then 0", TargetNode: "print1", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "seq", SourcePin: "Then 1", TargetNode: "print2", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "seq", SourcePin: "Then 2", TargetNode: "print3", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "start", SourcePin: "a", TargetNode: "print1", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c6", SourceNode: "start", SourcePin: "b", TargetNode: "print2", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c7", SourceNode: "start", SourcePin: "a", TargetNode: "print3", TargetPin: "value"})

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

// 测试 Switch 节点
func testSwitch() {
	fmt.Println("--- 测试 2: Switch 节点 ---")

	bp := blueprint.NewBlueprint("test_switch")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{
			{Name: "choice", Kind: blueprint.PinKindData, Type: "int", Value: 1.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "choice", Kind: blueprint.PinKindData, Type: "int"},
		},
	}

	// Switch 节点
	switchNode := &blueprint.Node{
		ID:        "switch",
		Type:      blueprint.NodeTypeFlowControl,
		Operation: "switch",
		Label:     "Switch",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "selection", Kind: blueprint.PinKindData, Type: "int"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "case_0", Kind: blueprint.PinKindExecution},
			{Name: "case_1", Kind: blueprint.PinKindExecution},
			{Name: "case_2", Kind: blueprint.PinKindExecution},
			{Name: "default", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 节点
	print1 := &blueprint.Node{
		ID:        "print1",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Case 1",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any", Value: "Case 1 executed"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Switch"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(switchNode)
	bp.AddNode(print1)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "switch", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "start", SourcePin: "choice", TargetNode: "switch", TargetPin: "selection"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "switch", SourcePin: "case_1", TargetNode: "print1", TargetPin: "exec_in"})

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

	if !result.Success {
		fmt.Printf("执行失败，错误: %v\n", result.Errors)
	}
	fmt.Printf("执行成功: %v\n\n", result.Success)
}

// 测试 ForEach 节点
func testForEach() {
	fmt.Println("--- 测试 3: ForEach 节点 ---")

	bp := blueprint.NewBlueprint("test_foreach")

	// Start 节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{
			{Name: "array", Kind: blueprint.PinKindData, Type: "any", Value: []interface{}{10.0, 20.0, 30.0}},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "array", Kind: blueprint.PinKindData, Type: "any"},
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
			{Name: "index", Kind: blueprint.PinKindData, Type: "int"},
		},
	}

	// Print 节点（打印元素）
	printElement := &blueprint.Node{
		ID:        "print_elem",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Element",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Element"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	// Print 完成
	printDone := &blueprint.Node{
		ID:        "print_done",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Done",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any", Value: "ForEach completed"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Done"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	bp.AddNode(start)
	bp.AddNode(foreach)
	bp.AddNode(printElement)
	bp.AddNode(printDone)

	// 连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "foreach", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "start", SourcePin: "array", TargetNode: "foreach", TargetPin: "array"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "foreach", SourcePin: "loop_body", TargetNode: "print_elem", TargetPin: "exec_in"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "foreach", SourcePin: "element", TargetNode: "print_elem", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "foreach", SourcePin: "completed", TargetNode: "print_done", TargetPin: "exec_in"})

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

	if !result.Success {
		fmt.Printf("执行失败，错误: %v\n", result.Errors)
	}
	fmt.Printf("执行成功: %v\n\n", result.Success)
}

// 测试函数子图
func testFunctionSubgraph() {
	fmt.Println("--- 测试 4: 函数子图 ---")

	// 创建函数蓝图：计算两数之和
	funcBP := blueprint.NewBlueprint("add_function")

	// Start 节点（定义函数输入）
	funcStart := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Add 节点
	addNode := &blueprint.Node{
		ID:        "add",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label:     "Add",
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// End 节点（定义函数输出）
	funcEnd := &blueprint.Node{
		ID:    "end",
		Type:  blueprint.NodeTypeEnd,
		Label: "End",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "sum", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{},
	}

	funcBP.AddNode(funcStart)
	funcBP.AddNode(addNode)
	funcBP.AddNode(funcEnd)

	funcBP.AddConnection(blueprint.Connection{ID: "fc1", SourceNode: "start", SourcePin: "exec", TargetNode: "end", TargetPin: "exec_in"})
	funcBP.AddConnection(blueprint.Connection{ID: "fc2", SourceNode: "start", SourcePin: "a", TargetNode: "add", TargetPin: "a"})
	funcBP.AddConnection(blueprint.Connection{ID: "fc3", SourceNode: "start", SourcePin: "b", TargetNode: "add", TargetPin: "b"})
	funcBP.AddConnection(blueprint.Connection{ID: "fc4", SourceNode: "add", SourcePin: "result", TargetNode: "end", TargetPin: "sum"})

	// 编译函数蓝图
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(funcBP); err != nil {
		fmt.Printf("函数编译失败: %v\n", err)
		return
	}

	// 注册函数
	funcRegistry := blueprint.NewFunctionRegistry()
	if err := funcRegistry.RegisterFunction("add", funcBP); err != nil {
		fmt.Printf("函数注册失败: %v\n", err)
		return
	}

	// 创建主蓝图，调用函数
	mainBP := blueprint.NewBlueprint("main_blueprint")

	// Start 节点
	mainStart := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "Start",
		InputPins: []blueprint.Pin{
			{Name: "x", Kind: blueprint.PinKindData, Type: "float", Value: 100.0},
			{Name: "y", Kind: blueprint.PinKindData, Type: "float", Value: 200.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "x", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "y", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// 函数调用节点
	callNode := &blueprint.Node{
		ID:        "call_add",
		Type:      blueprint.NodeTypeFunction,
		Operation: "add",
		Label:     "Call Add",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
			{Name: "sum", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// Print 节点
	printNode := &blueprint.Node{
		ID:        "print",
		Type:      blueprint.NodeTypeDebug,
		Operation: "print",
		Label:     "Print Result",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution},
			{Name: "value", Kind: blueprint.PinKindData, Type: "any"},
			{Name: "label", Kind: blueprint.PinKindData, Type: "string", Value: "Sum"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution},
		},
	}

	mainBP.AddNode(mainStart)
	mainBP.AddNode(callNode)
	mainBP.AddNode(printNode)

	mainBP.AddConnection(blueprint.Connection{ID: "mc1", SourceNode: "start", SourcePin: "exec", TargetNode: "call_add", TargetPin: "exec_in"})
	mainBP.AddConnection(blueprint.Connection{ID: "mc2", SourceNode: "start", SourcePin: "x", TargetNode: "call_add", TargetPin: "a"})
	mainBP.AddConnection(blueprint.Connection{ID: "mc3", SourceNode: "start", SourcePin: "y", TargetNode: "call_add", TargetPin: "b"})
	mainBP.AddConnection(blueprint.Connection{ID: "mc4", SourceNode: "call_add", SourcePin: "exec_out", TargetNode: "print", TargetPin: "exec_in"})
	mainBP.AddConnection(blueprint.Connection{ID: "mc5", SourceNode: "call_add", SourcePin: "sum", TargetNode: "print", TargetPin: "value"})

	// 注册函数调用执行器
	registry.Register(blueprint.NodeTypeFunction, "add", nodes.NewFunctionCallExecutor("add", funcRegistry))

	// 编译并执行主蓝图
	if err := compiler.Compile(mainBP); err != nil {
		fmt.Printf("主蓝图编译失败: %v\n", err)
		return
	}

	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(mainBP, nil)
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		return
	}

	fmt.Printf("执行成功: %v\n", result.Success)

	// 打印结果
	resultJSON, _ := json.MarshalIndent(result.Outputs, "", "  ")
	fmt.Printf("输出: %s\n", string(resultJSON))
}
