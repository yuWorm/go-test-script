package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 测试执行引脚功能（调试模式） ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 手动创建简单测试
	bp := &blueprint.Blueprint{
		Name:        "Simple Exec Pin Test",
		Description: "简单的执行引脚测试",
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
			{Name: "value", Kind: blueprint.PinKindData, Type: "float", Value: 42.0},
		},
	}
	bp.AddNode(start)

	// Add节点
	add := &blueprint.Node{
		ID:        "add",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label:     "加法",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 8.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec_out", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}
	bp.AddNode(add)

	// End节点
	end := &blueprint.Node{
		ID:    "end",
		Type:  blueprint.NodeTypeEnd,
		Label: "结束",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
			{Name: "final_result", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{},
	}
	bp.AddNode(end)

	// 添加连接
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
		SourcePin:  "value",
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
		TargetPin:  "final_result",
	})

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	fmt.Println("✓ 蓝图编译成功\n")

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
		fmt.Println("所有节点状态:")
		for _, node := range bp.Nodes {
			outputs := result.GetNodeOutputs(node.ID)
			if len(outputs) > 0 {
				fmt.Printf("  ✓ %s (%s):\n", node.ID, node.Label)
				for key, value := range outputs {
					fmt.Printf("    %s = %v\n", key, value)
				}
			} else {
				fmt.Printf("  ✗ %s (%s): 未执行\n", node.ID, node.Label)
			}
		}

		// 验证结果
		fmt.Println("\n最终检查:")
		if finalResult, ok := result.GetOutput("end", "final_result"); ok {
			fmt.Printf("  ✓ 最终结果: %v (预期: 50 = 42 + 8)\n", finalResult)
		} else {
			fmt.Println("  ✗ 未找到最终结果")
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
