package main

import (
	"fmt"
	"time"

	"github.com/go-blueprint-engine/blueprint"
)

func main() {
	fmt.Println("=== 测试异步任务机制 ===\n")

	// 创建任务管理器
	manager := blueprint.NewAsyncTaskManager()

	// 测试1: 先设置结果，后等待
	fmt.Println("测试1: 先设置结果，后等待")
	task1 := manager.CreateTask("test1")
	task1.SetResult(42)
	fmt.Println("✓ 结果已设置: 42")

	result1, err := task1.Wait(1 * time.Second)
	if err != nil {
		fmt.Printf("✗ 等待失败: %v\n", err)
	} else {
		fmt.Printf("✓ 等待成功，获取结果: %v\n\n", result1)
	}

	// 测试2: 先等待，后设置结果（在goroutine中）
	fmt.Println("测试2: 先等待，后设置结果")
	task2 := manager.CreateTask("test2")

	go func() {
		time.Sleep(500 * time.Millisecond)
		task2.SetResult(100)
		fmt.Println("  [goroutine] 结果已设置: 100")
	}()

	fmt.Println("  开始等待...")
	result2, err := task2.Wait(2 * time.Second)
	if err != nil {
		fmt.Printf("✗ 等待失败: %v\n", err)
	} else {
		fmt.Printf("✓ 等待成功，获取结果: %v\n\n", result2)
	}

	// 测试3: 超时测试
	fmt.Println("测试3: 超时测试")
	task3 := manager.CreateTask("test3")
	fmt.Println("  开始等待（1秒超时，但结果永不设置）...")
	result3, err := task3.Wait(1 * time.Second)
	if err != nil {
		fmt.Printf("✓ 预期的超时错误: %v\n\n", err)
	} else {
		fmt.Printf("✗ 应该超时但却成功了: %v\n\n", result3)
	}

	fmt.Println("=== 所有测试完成 ===")
}
