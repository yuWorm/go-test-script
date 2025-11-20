package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 测试执行引脚功能 ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 测试1: 从JSON加载带执行引脚的蓝图
	fmt.Println("测试1: 执行引脚分支测试")
	testExecutionPinsFromJSON(registry)
}

func testExecutionPinsFromJSON(registry *blueprint.NodeRegistry) {
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile("examples/execution_pins_demo.json")
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v\n", err)
	}

	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Description: %s\n\n", bp.Description)

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	fmt.Println("✓ 蓝图编译成功")

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

		// 打印节点输出
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
		if finalResult, ok := result.GetOutput("end", "final_result"); ok {
			fmt.Printf("\n✓ 最终结果: %v\n", finalResult)
			fmt.Println("  预期: 30 (condition=true时: (10+5)*2=30)")
			fmt.Println("  预期: 10 (condition=false时: (10+5)-5=10)")
		} else {
			fmt.Println("\n✗ 未找到最终结果")
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
