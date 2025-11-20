package main

import (
	"fmt"
	"log"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

func main() {
	fmt.Println("=== 测试节点编译时类型转换 ===\n")

	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 创建蓝图：测试 AND 节点的类型转换
	// Start -> AND (输入 "1" 和 "0" 字符串) -> End
	bp := &blueprint.Blueprint{
		Name:        "Node Compile Test",
		Version:     "1.0.0",
		Description: "测试节点编译时的类型转换",
		Nodes: []*blueprint.Node{
			{
				ID:    "start",
				Type:  blueprint.NodeTypeStart,
				Label: "开始",
				InputPins: []blueprint.Pin{},
				OutputPins: []blueprint.Pin{
					{Name: "exec", Kind: blueprint.PinKindExecution, Type: "exec"},
				},
				Position: blueprint.Position{X: 100, Y: 200},
			},
			{
				ID:        "and_node",
				Type:      blueprint.NodeTypeLogic,
				Operation: "and",
				Label:     "AND节点",
				InputPins: []blueprint.Pin{
					{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
					{Name: "a", Kind: blueprint.PinKindData, Type: "bool", Value: "1"}, // 字符串 "1" 应该被转换为 true
					{Name: "b", Kind: blueprint.PinKindData, Type: "bool", Value: 1},   // 整数 1 应该被转换为 true
				},
				OutputPins: []blueprint.Pin{
					{Name: "exec_out", Kind: blueprint.PinKindExecution, Type: "exec"},
					{Name: "result", Kind: blueprint.PinKindData, Type: "bool"},
				},
				Position: blueprint.Position{X: 300, Y: 200},
			},
			{
				ID:        "add_node",
				Type:      blueprint.NodeTypeArithmetic,
				Operation: "add",
				Label:     "加法节点",
				InputPins: []blueprint.Pin{
					{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
					{Name: "a", Kind: blueprint.PinKindData, Type: "float", Value: "10.5"}, // 字符串 "10.5" 应该被转换为 10.5
					{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 20},     // 整数 20 应该被转换为 20.0
				},
				OutputPins: []blueprint.Pin{
					{Name: "exec_out", Kind: blueprint.PinKindExecution, Type: "exec"},
					{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
				},
				Position: blueprint.Position{X: 500, Y: 200},
			},
			{
				ID:    "end",
				Type:  blueprint.NodeTypeEnd,
				Label: "结束",
				InputPins: []blueprint.Pin{
					{Name: "exec_in", Kind: blueprint.PinKindExecution, Type: "exec"},
					{Name: "and_result", Kind: blueprint.PinKindData, Type: "bool"},
					{Name: "add_result", Kind: blueprint.PinKindData, Type: "float"},
				},
				OutputPins: []blueprint.Pin{},
				Position:   blueprint.Position{X: 700, Y: 200},
			},
		},
		Connections: []blueprint.Connection{
			{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "and_node", TargetPin: "exec_in"},
			{ID: "c2", SourceNode: "and_node", SourcePin: "exec_out", TargetNode: "add_node", TargetPin: "exec_in"},
			{ID: "c3", SourceNode: "add_node", SourcePin: "exec_out", TargetNode: "end", TargetPin: "exec_in"},
			{ID: "c4", SourceNode: "and_node", SourcePin: "result", TargetNode: "end", TargetPin: "and_result"},
			{ID: "c5", SourceNode: "add_node", SourcePin: "result", TargetNode: "end", TargetPin: "add_result"},
		},
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	// 初始化节点映射
	bp.InitNodeMap()

	fmt.Println("📋 编译前的引脚值:")
	printPinValues(bp)

	// 编译蓝图
	fmt.Println("\n🔧 正在编译蓝图...")
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		log.Fatalf("编译失败: %v", err)
	}
	fmt.Println("✅ 编译成功！\n")

	fmt.Println("📋 编译后的引脚值:")
	printPinValues(bp)

	// 执行蓝图
	fmt.Println("\n⚙️  正在执行蓝图...")
	options := &blueprint.ExecutionOptions{
		Mode:        blueprint.ExecutionModeExecutionFlow,
		StopOnError: true,
	}
	executor := blueprint.NewExecutor(options)
	result, err := executor.Execute(bp, nil)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}

	// 显示结果
	fmt.Println("\n📊 执行结果:")
	fmt.Printf("  成功: %v\n", result.Success)
	fmt.Printf("  耗时: %v\n", result.Duration)

	if len(result.Errors) > 0 {
		fmt.Println("\n❌ 错误:")
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	fmt.Println("\n📤 节点输出:")
	for nodeID, nodeInfo := range result.Nodes {
		fmt.Printf("  %s (%s):\n", nodeID, nodeInfo.Status)
		for key, value := range nodeInfo.Outputs {
			fmt.Printf("    %s = %v (%T)\n", key, value, value)
		}
	}

	fmt.Println("\n✨ 预期结果:")
	fmt.Println("  - AND节点: true (因为 \"1\" 和 1 都被转换为 true)")
	fmt.Println("  - 加法节点: 30.5 (因为 \"10.5\" 转换为 10.5, 20 转换为 20.0)")
}

func printPinValues(bp *blueprint.Blueprint) {
	for _, node := range bp.Nodes {
		if node.Type == blueprint.NodeTypeStart || node.Type == blueprint.NodeTypeEnd {
			continue
		}
		fmt.Printf("  节点 %s (%s):\n", node.ID, node.Label)
		for _, pin := range node.InputPins {
			if pin.Kind != blueprint.PinKindExecution && pin.Value != nil {
				fmt.Printf("    输入 %s = %v (%T)\n", pin.Name, pin.Value, pin.Value)
			}
		}
	}
}
