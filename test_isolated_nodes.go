package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 孤立节点测试 ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 手动创建测试蓝图
	bp := &blueprint.Blueprint{
		Name:        "Isolated Node Test",
		Description: "测试孤立节点不会执行",
		Nodes:       make([]*blueprint.Node, 0),
		Connections: make([]blueprint.Connection, 0),
		Variables:   make(map[string]interface{}),
	}
	bp.InitNodeMap()

	// Start节点
	start := &blueprint.Node{
		ID:    "start",
		Type:  blueprint.NodeTypeStart,
		Label: "开始",
		InputPins: []blueprint.Pin{},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "value_a", Kind: blueprint.PinKindData, Type: "float", Value: 10.0},
		},
	}
	bp.AddNode(start)

	// 连接的节点：Add
	add := &blueprint.Node{
		ID:        "add",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label:     "加法（连接）",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 5.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}
	bp.AddNode(add)

	// 孤立节点：Multiply（没有连接到Start或End）
	isolatedMultiply := &blueprint.Node{
		ID:        "isolated_multiply",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "multiply",
		Label:     "乘法（孤立）",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float", Value: 100.0},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 2.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}
	bp.AddNode(isolatedMultiply)

	// End节点
	end := &blueprint.Node{
		ID:    "end",
		Type:  blueprint.NodeTypeEnd,
		Label: "结束",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{},
	}
	bp.AddNode(end)

	// 添加连接（不连接孤立节点）
	bp.AddConnection(blueprint.Connection{
		ID:         "e1",
		SourceNode: "start",
		SourcePin:  "exec",
		TargetNode: "add",
		TargetPin:  "exec_in",
	})

	bp.AddConnection(blueprint.Connection{
		ID:         "d1",
		SourceNode: "start",
		SourcePin:  "value_a",
		TargetNode: "add",
		TargetPin:  "a",
	})

	bp.AddConnection(blueprint.Connection{
		ID:         "e2",
		SourceNode: "add",
		SourcePin:  "exec_out",
		TargetNode: "end",
		TargetPin:  "exec_in",
	})

	bp.AddConnection(blueprint.Connection{
		ID:         "d2",
		SourceNode: "add",
		SourcePin:  "result",
		TargetNode: "end",
		TargetPin:  "result",
	})

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	fmt.Println("✓ 蓝图编译成功")
	fmt.Println("  连接路径: Start -> Add -> End")
	fmt.Println("  孤立节点: isolated_multiply (应该不执行)\n")

	// 使用执行流模式执行
	executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeExecutionFlow,
		StopOnError: true,
	})

	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	if result.Success {
		fmt.Printf("✓ 执行成功 (耗时: %v)\n\n", result.Duration)

		// 打印所有节点状态
		fmt.Println("节点执行状态:")
		for _, node := range bp.Nodes {
			outputs := result.GetNodeOutputs(node.ID)
			if len(outputs) > 0 {
				fmt.Printf("  ✓ %s (%s): 已执行\n", node.ID, node.Label)
				for key, value := range outputs {
					fmt.Printf("    %s = %v\n", key, value)
				}
			} else {
				fmt.Printf("  ✗ %s (%s): 未执行 (预期)\n", node.ID, node.Label)
			}
		}

		// 验证结果
		fmt.Println("\n验证:")
		if finalResult, ok := result.GetOutput("end", "result"); ok {
			fmt.Printf("  ✓ 最终结果: %v (预期: 15 = 10 + 5)\n", finalResult)
		}

		// 检查孤立节点确实没有执行
		if isolatedOutput, ok := result.GetOutput("isolated_multiply", "result"); ok {
			fmt.Printf("  ✗ 错误: 孤立节点被执行了，输出: %v\n", isolatedOutput)
		} else {
			fmt.Printf("  ✓ 孤立节点正确地未执行\n")
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
