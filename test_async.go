package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 测试1：基本异步执行
	fmt.Println("=== 测试1: 基本异步执行 ===")
	testAsyncCalculation(registry)
	fmt.Println()

	// 测试2：并行异步执行
	fmt.Println("=== 测试2: 并行异步执行 ===")
	testAsyncParallel(registry)
	fmt.Println()

	// 测试3：后台异步执行
	fmt.Println("=== 测试3: 后台异步执行 ===")
	testAsyncBackground(registry)
}

func testAsyncCalculation(registry *blueprint.NodeRegistry) {
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile("examples/async_calculation_demo.json")
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v\n", err)
	}

	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Description: %s\n", bp.Description)

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	// 执行
	startTime := time.Now()
	executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	})

	result, err := executor.Execute(bp, nil)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	if result.Success {
		fmt.Printf("✓ 执行成功 (耗时: %v)\n", duration)
		if finalResult, ok := result.GetOutput("end", "final_result"); ok {
			fmt.Printf("  最终结果: %v (预期: 300 = 100 * 3)\n", finalResult)
		}
		fmt.Printf("  说明: 异步任务休眠2秒后完成计算，await等待结果\n")
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}

func testAsyncParallel(registry *blueprint.NodeRegistry) {
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile("examples/async_parallel_demo.json")
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v\n", err)
	}

	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Description: %s\n", bp.Description)

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	// 执行
	startTime := time.Now()
	executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	})

	result, err := executor.Execute(bp, nil)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	if result.Success {
		fmt.Printf("✓ 执行成功 (耗时: %v)\n", duration)
		if total, ok := result.GetOutput("end", "total"); ok {
			fmt.Printf("  合并结果: %v (预期: 80 = 10*2 + 20*3)\n", total)
		}
		if taskCount, ok := result.GetOutput("end", "task_count"); ok {
			fmt.Printf("  任务数量: %v\n", taskCount)
		}
		fmt.Printf("  说明: 两个异步任务并行执行，最长耗时为1.5秒\n")
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}

func testAsyncBackground(registry *blueprint.NodeRegistry) {
	parser := blueprint.NewParser()
	bp, err := parser.ParseFromFile("examples/async_background_demo.json")
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v\n", err)
	}

	fmt.Printf("Blueprint: %s\n", bp.Name)
	fmt.Printf("Description: %s\n", bp.Description)

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	// 执行
	startTime := time.Now()
	executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	})

	result, err := executor.Execute(bp, nil)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	if result.Success {
		fmt.Printf("✓ 主流程执行成功 (耗时: %v)\n", duration)
		if status, ok := result.GetOutput("end", "status"); ok {
			fmt.Printf("  主流程状态: %v\n", status)
		}
		fmt.Printf("  说明: 主流程立即完成（耗时应远小于5秒），后台任务仍在运行\n")
		fmt.Printf("  注意: 后台任务会在5秒后完成计算 42*10=420\n")
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
