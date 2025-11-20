package blueprint_test

import (
	"testing"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func TestBasicArithmetic(t *testing.T) {
	// 创建蓝图：(10 + 20) * 2 = 60
	bp := blueprint.NewBlueprint("test_arithmetic")

	// 创建常量节点
	constA := &blueprint.Node{
		ID:    "const_a",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Constant A",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 10.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
	}

	constB := &blueprint.Node{
		ID:    "const_b",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Constant B",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 20.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
	}

	// 创建加法节点
	addNode := &blueprint.Node{
		ID:    "add",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label: "Add",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "float"},
		},
	}

	// 创建乘法节点
	constMultiplier := &blueprint.Node{
		ID:    "const_mult",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Multiplier",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 2.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
	}

	multiplyNode := &blueprint.Node{
		ID:    "multiply",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		Label: "Multiply",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "float"},
		},
	}

	// 添加节点
	if err := bp.AddNode(constA); err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}
	if err := bp.AddNode(constB); err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}
	if err := bp.AddNode(addNode); err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}
	if err := bp.AddNode(constMultiplier); err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}
	if err := bp.AddNode(multiplyNode); err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	// 添加连接
	connections := []blueprint.Connection{
		{ID: "c1", SourceNode: "const_a", SourcePin: "output", TargetNode: "add", TargetPin: "a"},
		{ID: "c2", SourceNode: "const_b", SourcePin: "output", TargetNode: "add", TargetPin: "b"},
		{ID: "c3", SourceNode: "add", SourcePin: "result", TargetNode: "multiply", TargetPin: "a"},
		{ID: "c4", SourceNode: "const_mult", SourcePin: "output", TargetNode: "multiply", TargetPin: "b"},
	}

	for _, conn := range connections {
		if err := bp.AddConnection(conn); err != nil {
			t.Fatalf("Failed to add connection: %v", err)
		}
	}

	// 创建注册表并注册节点
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 编译蓝图
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		t.Fatalf("Failed to compile blueprint: %v", err)
	}

	// 执行蓝图
	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(bp, nil)
	if err != nil {
		t.Fatalf("Failed to execute blueprint: %v", err)
	}

	if !result.Success {
		t.Fatalf("Execution failed with errors: %v", result.Errors)
	}

	// 验证结果
	output, ok := result.GetOutput("multiply", "result")
	if !ok {
		t.Fatal("Output not found")
	}

	expectedResult := 60.0
	if output != expectedResult {
		t.Fatalf("Expected %v, got %v", expectedResult, output)
	}

	t.Logf("Test passed! Result: %v", output)
}

func TestConditionalLogic(t *testing.T) {
	// 测试条件逻辑：if (15 > 10) then 15 * 2 else 10 * 3
	bp := blueprint.NewBlueprint("test_conditional")

	// 创建常量节点
	constA := &blueprint.Node{
		ID:    "const_a",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 15.0}},
		OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
	}

	constB := &blueprint.Node{
		ID:    "const_b",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 10.0}},
		OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
	}

	// 比较节点
	compareNode := &blueprint.Node{
		ID:    "compare",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "greater",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{{Name: "result", Type: "bool"}},
	}

	// 乘法节点（true 分支）
	const2 := &blueprint.Node{
		ID:    "const_2",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 2.0}},
		OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
	}

	multiplyA := &blueprint.Node{
		ID:    "mult_a",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{{Name: "result", Type: "float"}},
	}

	// 乘法节点（false 分支）
	const3 := &blueprint.Node{
		ID:    "const_3",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 3.0}},
		OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
	}

	multiplyB := &blueprint.Node{
		ID:    "mult_b",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{{Name: "result", Type: "float"}},
	}

	// 条件节点
	conditionNode := &blueprint.Node{
		ID:    "condition",
		Type:  blueprint.NodeTypeCondition,
		InputPins: []blueprint.Pin{
			{Name: "condition", Type: "bool"},
			{Name: "true_value", Type: "any"},
			{Name: "false_value", Type: "any"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "any"},
			{Name: "is_true", Type: "bool"},
			{Name: "is_false", Type: "bool"},
		},
	}

	// 添加所有节点
	testNodes := []*blueprint.Node{constA, constB, compareNode, const2, const3, multiplyA, multiplyB, conditionNode}
	for _, node := range testNodes {
		if err := bp.AddNode(node); err != nil {
			t.Fatalf("Failed to add node: %v", err)
		}
	}

	// 添加连接
	connections := []blueprint.Connection{
		{ID: "c1", SourceNode: "const_a", SourcePin: "output", TargetNode: "compare", TargetPin: "a"},
		{ID: "c2", SourceNode: "const_b", SourcePin: "output", TargetNode: "compare", TargetPin: "b"},
		{ID: "c3", SourceNode: "const_a", SourcePin: "output", TargetNode: "mult_a", TargetPin: "a"},
		{ID: "c4", SourceNode: "const_2", SourcePin: "output", TargetNode: "mult_a", TargetPin: "b"},
		{ID: "c5", SourceNode: "const_b", SourcePin: "output", TargetNode: "mult_b", TargetPin: "a"},
		{ID: "c6", SourceNode: "const_3", SourcePin: "output", TargetNode: "mult_b", TargetPin: "b"},
		{ID: "c7", SourceNode: "compare", SourcePin: "result", TargetNode: "condition", TargetPin: "condition"},
		{ID: "c8", SourceNode: "mult_a", SourcePin: "result", TargetNode: "condition", TargetPin: "true_value"},
		{ID: "c9", SourceNode: "mult_b", SourcePin: "result", TargetNode: "condition", TargetPin: "false_value"},
	}

	for _, conn := range connections {
		if err := bp.AddConnection(conn); err != nil {
			t.Fatalf("Failed to add connection: %v", err)
		}
	}

	// 编译和执行
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		t.Fatalf("Failed to compile blueprint: %v", err)
	}

	executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
	result, err := executor.Execute(bp, nil)
	if err != nil {
		t.Fatalf("Failed to execute blueprint: %v", err)
	}

	if !result.Success {
		t.Fatalf("Execution failed with errors: %v", result.Errors)
	}

	// 验证结果（15 > 10 为 true，所以应该是 15 * 2 = 30）
	output, ok := result.GetOutput("condition", "result")
	if !ok {
		t.Fatal("Output not found")
	}

	expectedResult := 30.0
	if output != expectedResult {
		t.Fatalf("Expected %v, got %v", expectedResult, output)
	}

	t.Logf("Test passed! Result: %v", output)
}

func TestParallelExecution(t *testing.T) {
	// 测试并行执行模式
	bp := blueprint.NewBlueprint("test_parallel")

	// 创建多个独立的计算分支
	testNodes := []*blueprint.Node{
		{
			ID: "c1", Type: blueprint.NodeTypeFunction, Operation: "constant",
			InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 5.0}},
			OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
		},
		{
			ID: "c2", Type: blueprint.NodeTypeFunction, Operation: "constant",
			InputPins: []blueprint.Pin{{Name: "value", Type: "float", Value: 3.0}},
			OutputPins: []blueprint.Pin{{Name: "output", Type: "float"}},
		},
		{
			ID: "add1", Type: blueprint.NodeTypeArithmetic, Operation: "add",
			InputPins: []blueprint.Pin{{Name: "a", Type: "float"}, {Name: "b", Type: "float"}},
			OutputPins: []blueprint.Pin{{Name: "result", Type: "float"}},
		},
	}

	for _, node := range testNodes {
		if err := bp.AddNode(node); err != nil {
			t.Fatalf("Failed to add node: %v", err)
		}
	}

	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "c1", SourcePin: "output", TargetNode: "add1", TargetPin: "a"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "c2", SourcePin: "output", TargetNode: "add1", TargetPin: "b"})

	// 使用并行执行模式
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	options := &blueprint.ExecutionOptions{
		Mode:           blueprint.ExecutionModeParallel,
		MaxConcurrency: 4,
		StopOnError:    true,
	}

	executor := blueprint.NewExecutor(options)
	result, err := executor.Execute(bp, nil)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	if !result.Success {
		t.Fatalf("Execution failed: %v", result.Errors)
	}

	output, _ := result.GetOutput("add1", "result")
	if output != 8.0 {
		t.Fatalf("Expected 8.0, got %v", output)
	}

	t.Logf("Parallel execution test passed! Duration: %v", result.Duration)
}
