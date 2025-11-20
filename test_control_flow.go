package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 解析控制流演示蓝图
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile("examples/control_flow_demo.json")
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v\n", err)
	}

	fmt.Printf("=== 控制流演示 ===\n")
	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Description: %s\n\n", bp.Description)

	// 编译蓝图
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile blueprint: %v\n", err)
	}
	fmt.Println("✓ 蓝图编译成功")

	// 执行蓝图
	options := &blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	}
	executor := blueprint.NewExecutor(options)

	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	// 显示结果
	if result.Success {
		fmt.Println("✓ 执行成功")
		fmt.Printf("执行时长: %v\n\n", result.Duration)

		fmt.Println("输出结果:")
		for key, value := range result.Outputs {
			fmt.Printf("  %s = %v\n", key, value)
		}

		// 特别显示关键结果
		fmt.Println("\n关键结果:")
		if ifResult, ok := result.GetOutput("if_else", "result"); ok {
			fmt.Printf("  If/Else 结果: %v (15 > 10 为真，返回 then_value = 100)\n", ifResult)
		}
		if loopCount, ok := result.GetOutput("for_loop", "count"); ok {
			fmt.Printf("  循环次数: %v (从0到10，步长1，共10次迭代)\n", loopCount)
		}
		if loopIndex, ok := result.GetOutput("for_loop", "index"); ok {
			fmt.Printf("  最后索引: %v\n", loopIndex)
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
