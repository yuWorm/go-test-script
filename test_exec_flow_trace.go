package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 执行流跟踪测试 ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 手动创建简单测试
	bp := &blueprint.Blueprint{
		Name:        "Exec Flow Trace",
		Description: "执行流跟踪测试",
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
		},
	}
	bp.AddNode(start)

	// End节点
	end := &blueprint.Node{
		ID:    "end",
		Type:  blueprint.NodeTypeEnd,
		Label: "结束",
		InputPins: []blueprint.Pin{
			{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
		},
		OutputPins: []blueprint.Pin{},
	}
	bp.AddNode(end)

	// 添加执行流连接
	bp.AddConnection(blueprint.Connection{
		ID:         "e1",
		SourceNode: "start",
		SourcePin:  "exec",
		TargetNode: "end",
		TargetPin:  "exec_in",
	})

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("Failed to compile: %v\n", err)
	}

	fmt.Println("✓ 蓝图编译成功")
	fmt.Println("  节点: start, end")
	fmt.Println("  连接: start.exec -> end.exec_in\n")

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
		fmt.Printf("✓ 执行成功\n\n")

		// 打印所有节点状态
		fmt.Println("节点执行状态:")
		for _, node := range bp.Nodes {
			outputs := result.GetNodeOutputs(node.ID)
			if len(outputs) > 0 {
				fmt.Printf("  ✓ %s: 已执行\n", node.ID)
				for key, value := range outputs {
					fmt.Printf("    %s = %v\n", key, value)
				}
			} else {
				fmt.Printf("  ✗ %s: 未执行\n", node.ID)
			}
		}
	} else {
		fmt.Println("✗ 执行失败")
		for _, err := range result.Errors {
			fmt.Printf("  错误: %v\n", err)
		}
	}
}
