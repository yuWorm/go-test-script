package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 简单异步测试 ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 手动创建一个简单的蓝图
	bp := &blueprint.Blueprint{
		Name:        "Manual Async Test",
		Description: "手动创建的异步测试",
		Nodes:       make([]*blueprint.Node, 0),
		Connections: make([]blueprint.Connection, 0),
		Variables:   make(map[string]interface{}),
	}
	bp.InitNodeMap()

	// 创建节点
	start := &blueprint.Node{
		ID:         "start",
		Type:       blueprint.NodeTypeStart,
		Label:      "开始",
		InputPins:  []blueprint.Pin{},
		OutputPins: []blueprint.Pin{},
	}
	bp.AddNode(start)

	asyncStart := &blueprint.Node{
		ID:        "async_start",
		Type:      blueprint.NodeTypeFunction,
		Operation: "async_start",
		Label:     "异步开始",
		InputPins: []blueprint.Pin{},
		OutputPins: []blueprint.Pin{
			{Name: "task_id", Type: "string"},
			{Name: "started", Type: "bool"},
		},
	}
	bp.AddNode(asyncStart)

	constant := &blueprint.Node{
		ID:         "constant",
		Type:       blueprint.NodeTypeFunction,
		Operation:  "constant",
		Label:      "常量值",
		InputPins:  []blueprint.Pin{{Name: "value", Type: "float", Value: 42.0}},
		OutputPins: []blueprint.Pin{{Name: "output", Type: "any"}},
	}
	bp.AddNode(constant)

	asyncEnd := &blueprint.Node{
		ID:        "async_end",
		Type:      blueprint.NodeTypeFunction,
		Operation: "async_end",
		Label:     "异步结束",
		InputPins: []blueprint.Pin{
			{Name: "task_id", Type: "string"},
			{Name: "result", Type: "any"},
		},
		OutputPins: []blueprint.Pin{{Name: "completed", Type: "bool"}},
	}
	bp.AddNode(asyncEnd)

	await := &blueprint.Node{
		ID:        "await",
		Type:      blueprint.NodeTypeFunction,
		Operation: "await",
		Label:     "等待结果",
		InputPins: []blueprint.Pin{
			{Name: "task_id", Type: "string"},
			{Name: "timeout", Type: "float", Value: 5.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Type: "any"},
			{Name: "completed", Type: "bool"},
		},
	}
	bp.AddNode(await)

	end := &blueprint.Node{
		ID:         "end",
		Type:       blueprint.NodeTypeEnd,
		Label:      "结束",
		InputPins:  []blueprint.Pin{{Name: "result", Type: "any"}},
		OutputPins: []blueprint.Pin{},
	}
	bp.AddNode(end)

	// 添加连接，确保执行顺序: start -> asyncStart -> constant -> asyncEnd -> await -> end
	bp.AddConnection(blueprint.Connection{
		ID:         "c1",
		SourceNode: "async_start",
		SourcePin:  "task_id",
		TargetNode: "async_end",
		TargetPin:  "task_id",
	})

	bp.AddConnection(blueprint.Connection{
		ID:         "c2",
		SourceNode: "constant",
		SourcePin:  "output",
		TargetNode: "async_end",
		TargetPin:  "result",
	})

	bp.AddConnection(blueprint.Connection{
		ID:         "c3",
		SourceNode: "async_start",
		SourcePin:  "task_id",
		TargetNode: "await",
		TargetPin:  "task_id",
	})

	// 关键: 让 await 依赖于 async_end 的完成，通过连接 completed 到一个虚拟输入
	// 但这会导致类型问题，所以我们换个方式：让 await 依赖于 asyncEnd 节点本身
	// 通过让 constant 依赖于 asyncStart
	bp.Connections[1].SourceNode = "async_start"
	bp.Connections[1].SourcePin = "started"
	// 实际上，我们需要 asyncEnd 在 await 之前执行
	// 最简单的方式是让 await 依赖于 asyncEnd 的输出

	// 重新设计连接
	bp.Connections = []blueprint.Connection{
		{
			ID:         "c1",
			SourceNode: "async_start",
			SourcePin:  "task_id",
			TargetNode: "async_end",
			TargetPin:  "task_id",
		},
		{
			ID:         "c2",
			SourceNode: "constant",
			SourcePin:  "output",
			TargetNode: "async_end",
			TargetPin:  "result",
		},
		{
			ID:         "c3",
			SourceNode: "async_end",
			SourcePin:  "task_id",
			TargetNode: "await",
			TargetPin:  "task_id",
		},
		{
			ID:         "c4",
			SourceNode: "await",
			SourcePin:  "result",
			TargetNode: "end",
			TargetPin:  "result",
		},
	}

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	fmt.Println("✓ 蓝图编译成功")

	// 执行
	executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeSequential,
		StopOnError: true,
	})

	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Fatalf("Execution error: %v\n", err)
	}

	if result.Success {
		fmt.Printf("✓ 执行成功\n")
		fmt.Printf("执行时长: %v\n", result.Duration)

		// 打印所有输出
		fmt.Println("\n所有节点输出:")
		for _, node := range bp.Nodes {
			outputs := result.GetNodeOutputs(node.ID)
			if len(outputs) > 0 {
				fmt.Printf("  %s (%s):\n", node.ID, node.Label)
				for key, value := range outputs {
					fmt.Printf("    %s = %v\n", key, value)
				}
			}
		}

		if finalResult, ok := result.GetOutput("end", "result"); ok {
			fmt.Printf("\n✓ 最终结果: %v (预期: 42)\n", finalResult)
		} else {
			fmt.Printf("\n✗ 未找到最终结果\n")
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
