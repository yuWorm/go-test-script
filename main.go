package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== Golang Blueprint Engine Demo ===\n")

	// 创建节点注册表并注册所有内置节点
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	fmt.Println("Registered node types:")
	for _, nodeType := range registry.ListRegistered() {
		fmt.Printf("  - %s\n", nodeType)
	}
	fmt.Println()

	// 示例 1: 从 JSON 加载并执行简单计算器蓝图
	fmt.Println("Example 1: Simple Calculator (10 + 20) * 2")
	fmt.Println("-------------------------------------------")
	runExample("examples/simple_calculator.json", registry)

	fmt.Println("\n")

	// 示例 2: 条件逻辑蓝图
	fmt.Println("Example 2: Conditional Logic")
	fmt.Println("-----------------------------")
	runExample("examples/conditional_logic.json", registry)

	fmt.Println("\n")

	// 示例 3: 程序化创建蓝图
	fmt.Println("Example 3: Programmatic Blueprint Creation")
	fmt.Println("-------------------------------------------")
	runProgrammaticExample(registry)
}

func runExample(filename string, registry *blueprint.NodeRegistry) {
	// 解析蓝图 JSON
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile(filename)
	if err != nil {
		log.Printf("Failed to parse blueprint: %v\n", err)
		return
	}

	fmt.Printf("Blueprint: %s (v%s)\n", bp.Name, bp.Version)
	fmt.Printf("Description: %s\n", bp.Description)
	fmt.Printf("Nodes: %d, Connections: %d\n", len(bp.Nodes), len(bp.Connections))

	// 编译蓝图
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Printf("Failed to compile blueprint: %v\n", err)
		return
	}
	fmt.Println("✓ Blueprint compiled successfully")

	// 执行蓝图（顺序模式）
	options := &blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	}
	executor := blueprint.NewExecutor(options)

	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Printf("Execution error: %v\n", err)
		return
	}

	// 显示结果
	if result.Success {
		fmt.Println("✓ Execution completed successfully")
		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Println("\nOutputs:")
		for key, value := range result.Outputs {
			fmt.Printf("  %s = %v\n", key, value)
		}
	} else {
		fmt.Println("✗ Execution failed")
		for _, err := range result.Errors {
			fmt.Printf("  Error: %v\n", err)
		}
	}

	// 保存蓝图到新文件（演示序列化）
	serializer := blueprint.NewSerializer(true)
	outputPath := filepath.Join("examples", "output_"+filepath.Base(filename))
	if err := serializer.ToFile(bp, outputPath); err != nil {
		log.Printf("Failed to save blueprint: %v\n", err)
	} else {
		fmt.Printf("\n✓ Blueprint saved to: %s\n", outputPath)
	}
}

func runProgrammaticExample(registry *blueprint.NodeRegistry) {
	// 程序化创建一个蓝图：计算 (a + b) * c
	// 其中 a, b, c 是输入参数

	bp := blueprint.NewBlueprint("programmatic_example")
	bp.Description = "Programmatically created blueprint: (a + b) * c"

	// 创建输入常量节点
	nodeA := &blueprint.Node{
		ID:    "input_a",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Input A",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 7.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
		Position: blueprint.Position{X: 100, Y: 100},
	}

	nodeB := &blueprint.Node{
		ID:    "input_b",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Input B",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 3.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
		Position: blueprint.Position{X: 100, Y: 200},
	}

	nodeC := &blueprint.Node{
		ID:    "input_c",
		Type:  blueprint.NodeTypeFunction,
		Operation: "constant",
		Label: "Input C",
		InputPins: []blueprint.Pin{
			{Name: "value", Type: "float", Value: 5.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "output", Type: "float"},
		},
		Position: blueprint.Position{X: 100, Y: 300},
	}

	// 加法节点
	addNode := &blueprint.Node{
		ID:    "add",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label: "Add A + B",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "float"},
		},
		Position: blueprint.Position{X: 300, Y: 150},
	}

	// 乘法节点
	multiplyNode := &blueprint.Node{
		ID:    "multiply",
		Type:  blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		Label: "Multiply by C",
		InputPins: []blueprint.Pin{
			{Name: "a", Type: "float"},
			{Name: "b", Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "float"},
		},
		Position: blueprint.Position{X: 500, Y: 200},
	}

	// 添加节点
	nodes := []*blueprint.Node{nodeA, nodeB, nodeC, addNode, multiplyNode}
	for _, node := range nodes {
		if err := bp.AddNode(node); err != nil {
			log.Printf("Failed to add node: %v\n", err)
			return
		}
	}

	// 添加连接
	connections := []blueprint.Connection{
		{ID: "conn1", SourceNode: "input_a", SourcePin: "output", TargetNode: "add", TargetPin: "a"},
		{ID: "conn2", SourceNode: "input_b", SourcePin: "output", TargetNode: "add", TargetPin: "b"},
		{ID: "conn3", SourceNode: "add", SourcePin: "result", TargetNode: "multiply", TargetPin: "a"},
		{ID: "conn4", SourceNode: "input_c", SourcePin: "output", TargetNode: "multiply", TargetPin: "b"},
	}

	for _, conn := range connections {
		if err := bp.AddConnection(conn); err != nil {
			log.Printf("Failed to add connection: %v\n", err)
			return
		}
	}

	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Nodes: %d, Connections: %d\n", len(bp.Nodes), len(bp.Connections))

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Printf("Failed to compile: %v\n", err)
		return
	}
	fmt.Println("✓ Blueprint compiled successfully")

	// 执行（使用并行模式）
	options := &blueprint.ExecutionOptions{
		Mode:           blueprint.ExecutionModeParallel,
		MaxConcurrency: 4,
		StopOnError:    true,
	}
	executor := blueprint.NewExecutor(options)

	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Printf("Execution error: %v\n", err)
		return
	}

	if result.Success {
		fmt.Println("✓ Execution completed successfully")
		fmt.Printf("Duration: %v\n", result.Duration)

		// 获取最终结果
		if output, ok := result.GetOutput("multiply", "result"); ok {
			fmt.Printf("\nFinal Result: %v\n", output)
			fmt.Printf("Calculation: (7 + 3) * 5 = %v\n", output)
		}
	} else {
		fmt.Println("✗ Execution failed")
		for _, err := range result.Errors {
			fmt.Printf("  Error: %v\n", err)
		}
	}

	// 保存为 JSON
	serializer := blueprint.NewSerializer(true)
	outputFile := "examples/programmatic_example.json"
	if err := serializer.ToFile(bp, outputFile); err != nil {
		log.Printf("Failed to save blueprint: %v\n", err)
	} else {
		fmt.Printf("\n✓ Blueprint saved to: %s\n", outputFile)
	}
}

func init() {
	// 确保 examples 目录存在
	if err := os.MkdirAll("examples", 0755); err != nil {
		log.Fatalf("Failed to create examples directory: %v", err)
	}
}
