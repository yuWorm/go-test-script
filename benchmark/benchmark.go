package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

type BenchmarkResult struct {
	Language string  `json:"language"`
	N        int     `json:"n"`
	Result   int64   `json:"result"`
	TimeMs   float64 `json:"time_ms"`
	MemoryKB uint64  `json:"memory_kb"`
}

// Go 原生 Fibonacci (迭代)
func fibonacciGo(n int) int64 {
	if n <= 1 {
		return int64(n)
	}
	var a, b int64 = 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func benchmarkGo(n int) BenchmarkResult {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()
	result := fibonacciGo(n)
	elapsed := time.Since(start)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	return BenchmarkResult{
		Language: "Go",
		N:        n,
		Result:   result,
		TimeMs:   float64(elapsed.Nanoseconds()) / 1e6,
		MemoryKB: (m2.TotalAlloc - m1.TotalAlloc) / 1024,
	}
}

func benchmarkNodeJS(n int) BenchmarkResult {
	jsCode := fmt.Sprintf(`
const start = process.hrtime.bigint();
const startMem = process.memoryUsage().heapUsed;

function fib(n) {
    if (n <= 1) return BigInt(n);
    let a = 0n, b = 1n;
    for (let i = 2; i <= n; i++) {
        [a, b] = [b, a + b];
    }
    return b;
}

const result = fib(%d);
const elapsed = Number(process.hrtime.bigint() - start) / 1e6;
const memUsed = (process.memoryUsage().heapUsed - startMem) / 1024;

console.log(JSON.stringify({
    result: result.toString(),
    time_ms: elapsed,
    memory_kb: Math.max(0, memUsed)
}));
`, n)

	cmd := exec.Command("node", "-e", jsCode)
	output, err := cmd.Output()
	if err != nil {
		return BenchmarkResult{Language: "Node.js", N: n, TimeMs: -1}
	}

	var data struct {
		Result   string  `json:"result"`
		TimeMs   float64 `json:"time_ms"`
		MemoryKB float64 `json:"memory_kb"`
	}
	json.Unmarshal(output, &data)

	var result int64
	fmt.Sscanf(data.Result, "%d", &result)

	return BenchmarkResult{
		Language: "Node.js",
		N:        n,
		Result:   result,
		TimeMs:   data.TimeMs,
		MemoryKB: uint64(data.MemoryKB),
	}
}

func benchmarkPython(n int) BenchmarkResult {
	pyCode := fmt.Sprintf(`
import time
import tracemalloc
import json

tracemalloc.start()
start = time.perf_counter()

def fib(n):
    if n <= 1:
        return n
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    return b

result = fib(%d)
elapsed = (time.perf_counter() - start) * 1000
current, peak = tracemalloc.get_traced_memory()
tracemalloc.stop()

print(json.dumps({
    "result": str(result),
    "time_ms": elapsed,
    "memory_kb": peak / 1024
}))
`, n)

	cmd := exec.Command("python3", "-c", pyCode)
	output, err := cmd.Output()
	if err != nil {
		return BenchmarkResult{Language: "Python", N: n, TimeMs: -1}
	}

	var data struct {
		Result   string  `json:"result"`
		TimeMs   float64 `json:"time_ms"`
		MemoryKB float64 `json:"memory_kb"`
	}
	json.Unmarshal(output, &data)

	var result int64
	fmt.Sscanf(data.Result, "%d", &result)

	return BenchmarkResult{
		Language: "Python",
		N:        n,
		Result:   result,
		TimeMs:   data.TimeMs,
		MemoryKB: uint64(data.MemoryKB),
	}
}

// 蓝图引擎 Fibonacci
func benchmarkBlueprint(n int) BenchmarkResult {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 创建注册表并注册节点
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 创建 Fibonacci 蓝图
	bp := createFibonacciBlueprint(n)

	// 编译
	compiler := blueprint.NewCompiler(registry)
	if err := compiler.Compile(bp); err != nil {
		fmt.Println("Compile error:", err)
		return BenchmarkResult{Language: "Blueprint", N: n, TimeMs: -1}
	}


	// 执行
	executor := blueprint.NewExecutor(nil)
	start := time.Now()
	result, err := executor.Execute(bp, nil)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println("Execute error:", err)
		return BenchmarkResult{Language: "Blueprint", N: n, TimeMs: -1}
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// 获取结果 - 从变量 b 中获取
	var fibResult int64
	// 注意: 蓝图执行流程可能需要进一步调试
	// Variables 和 Outputs 的结果依赖于蓝图的正确配置

	// 先尝试从 Variables 获取
	if result.Variables != nil {
		if v, ok := result.Variables["b"]; ok {
			switch r := v.(type) {
			case float64:
				fibResult = int64(r)
			case int64:
				fibResult = r
			case int:
				fibResult = int64(r)
			}
		}
	}
	// 如果 Variables 没有，尝试从 Outputs 获取
	if fibResult == 0 && result.Outputs != nil {
		for k, v := range result.Outputs {
			if k == "result" || k == "end.result" {
				switch r := v.(type) {
				case float64:
					fibResult = int64(r)
				case int64:
					fibResult = r
				case int:
					fibResult = int64(r)
				}
				break
			}
		}
	}

	return BenchmarkResult{
		Language: "Blueprint",
		N:        n,
		Result:   fibResult,
		TimeMs:   float64(elapsed.Nanoseconds()) / 1e6,
		MemoryKB: (m2.TotalAlloc - m1.TotalAlloc) / 1024,
	}
}

func createFibonacciBlueprint(n int) *blueprint.Blueprint {
	bp := blueprint.NewBlueprint("fibonacci")

	// Start 节点
	startNode := &blueprint.Node{
		ID:        "start",
		Type:      blueprint.NodeTypeStart,
		Operation: "start",
		Label:     "开始",
		Position:  blueprint.Position{X: 0, Y: 0},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "n", Kind: blueprint.PinKindData, Type: "float", Value: float64(n)},
		},
	}

	// 变量节点: a = 0
	varA := &blueprint.Node{
		ID:        "var_a",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "设置 a",
		Position:  blueprint.Position{X: 200, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "a"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float", Value: 0.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// 变量节点: b = 1
	varB := &blueprint.Node{
		ID:        "var_b",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "设置 b",
		Position:  blueprint.Position{X: 400, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "b"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float", Value: 1.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// 变量节点: i = 2
	varI := &blueprint.Node{
		ID:        "var_i",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "设置 i",
		Position:  blueprint.Position{X: 600, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "i"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float", Value: 2.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// While 循环: i <= n
	whileNode := &blueprint.Node{
		ID:        "while",
		Type:      blueprint.NodeTypeFlowControl,
		Operation: "while_loop",
		Label:     "循环",
		Position:  blueprint.Position{X: 800, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "condition", Kind: blueprint.PinKindData, Type: "bool"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "loop", Kind: blueprint.PinKindExecution},
			{Name: "done", Kind: blueprint.PinKindExecution},
		},
	}

	// 条件: i <= n (获取 i)
	getI := &blueprint.Node{
		ID:        "get_i",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取 i",
		Position:  blueprint.Position{X: 700, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "i"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// 条件: i <= n (比较)
	compareNode := &blueprint.Node{
		ID:        "compare",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "less_equal",
		Label:     "i <= n",
		Position:  blueprint.Position{X: 750, Y: 50},
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: float64(n)},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Kind: blueprint.PinKindData, Type: "bool"},
		},
	}

	// 获取 a
	getA := &blueprint.Node{
		ID:        "get_a",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取 a",
		Position:  blueprint.Position{X: 900, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "a"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// 获取 b
	getB := &blueprint.Node{
		ID:        "get_b",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取 b",
		Position:  blueprint.Position{X: 900, Y: 150},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "b"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// a + b
	mathAddNode := &blueprint.Node{
		ID:        "add",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label:     "a + b",
		Position:  blueprint.Position{X: 1000, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// temp = a + b (先保存和，避免覆盖问题)
	setTemp := &blueprint.Node{
		ID:        "set_temp",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "temp = a+b",
		Position:  blueprint.Position{X: 1000, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "temp"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// a = b
	setA := &blueprint.Node{
		ID:        "set_a",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "a = b",
		Position:  blueprint.Position{X: 1100, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "a"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// 获取 temp
	getTemp := &blueprint.Node{
		ID:        "get_temp",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取 temp",
		Position:  blueprint.Position{X: 1150, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "temp"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// b = temp
	setB := &blueprint.Node{
		ID:        "set_b",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "b = temp",
		Position:  blueprint.Position{X: 1200, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "b"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// 获取 i (用于 i++)
	getI2 := &blueprint.Node{
		ID:        "get_i2",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取 i",
		Position:  blueprint.Position{X: 1300, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "i"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// i + 1
	incI := &blueprint.Node{
		ID:        "inc_i",
		Type:      blueprint.NodeTypeArithmetic,
		Operation: "add",
		Label:     "i + 1",
		Position:  blueprint.Position{X: 1400, Y: 100},
		InputPins: []blueprint.Pin{
			{Name: "a", Kind: blueprint.PinKindData, Type: "float"},
			{Name: "b", Kind: blueprint.PinKindData, Type: "float", Value: 1.0},
		},
		OutputPins: []blueprint.Pin{
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// i = i + 1
	setI := &blueprint.Node{
		ID:        "set_i",
		Type:      blueprint.NodeTypeVariable,
		Operation: "set_variable",
		Label:     "i++",
		Position:  blueprint.Position{X: 1400, Y: 0},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "i"},
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
		},
	}

	// 获取最终结果
	getFinalB := &blueprint.Node{
		ID:        "get_final_b",
		Type:      blueprint.NodeTypeData,
		Operation: "get_variable",
		Label:     "获取结果",
		Position:  blueprint.Position{X: 1000, Y: 200},
		InputPins: []blueprint.Pin{
			{Name: "name", Kind: blueprint.PinKindData, Type: "string", Value: "b"},
		},
		OutputPins: []blueprint.Pin{
			{Name: "value", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// End 节点
	endNode := &blueprint.Node{
		ID:        "end",
		Type:      blueprint.NodeTypeEnd,
		Operation: "end",
		Label:     "结束",
		Position:  blueprint.Position{X: 1200, Y: 200},
		InputPins: []blueprint.Pin{
			{Name: "exec", Kind: blueprint.PinKindExecution},
			{Name: "result", Kind: blueprint.PinKindData, Type: "float"},
		},
	}

	// 添加所有节点（带错误检查）
	mustAddNode := func(node *blueprint.Node) {
		if err := bp.AddNode(node); err != nil {
			fmt.Printf("AddNode error for %s: %v\n", node.ID, err)
		}
	}
	mustAddNode(startNode)
	mustAddNode(varA)
	mustAddNode(varB)
	mustAddNode(varI)
	mustAddNode(whileNode)
	mustAddNode(getI)
	mustAddNode(compareNode)
	mustAddNode(getA)
	mustAddNode(getB)
	mustAddNode(mathAddNode)
	mustAddNode(setTemp)
	mustAddNode(setA)
	mustAddNode(getTemp)
	mustAddNode(setB)
	mustAddNode(getI2)
	mustAddNode(incI)
	mustAddNode(setI)
	mustAddNode(getFinalB)
	mustAddNode(endNode)

	// 执行流连接
	bp.AddConnection(blueprint.Connection{ID: "c1", SourceNode: "start", SourcePin: "exec", TargetNode: "var_a", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c2", SourceNode: "var_a", SourcePin: "exec", TargetNode: "var_b", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c3", SourceNode: "var_b", SourcePin: "exec", TargetNode: "var_i", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c4", SourceNode: "var_i", SourcePin: "exec", TargetNode: "while", TargetPin: "exec"})
	// 循环体: set_temp -> set_a -> set_b -> set_i -> while
	bp.AddConnection(blueprint.Connection{ID: "c5", SourceNode: "while", SourcePin: "loop", TargetNode: "set_temp", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c6", SourceNode: "set_temp", SourcePin: "exec", TargetNode: "set_a", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c7", SourceNode: "set_a", SourcePin: "exec", TargetNode: "set_b", TargetPin: "exec"})
	bp.AddConnection(blueprint.Connection{ID: "c8", SourceNode: "set_b", SourcePin: "exec", TargetNode: "set_i", TargetPin: "exec"})
	// Note: 不需要连接 set_i -> while，while loop 内部自动循环
	bp.AddConnection(blueprint.Connection{ID: "c10", SourceNode: "while", SourcePin: "done", TargetNode: "end", TargetPin: "exec"})

	// 数据流连接
	bp.AddConnection(blueprint.Connection{ID: "d1", SourceNode: "get_i", SourcePin: "value", TargetNode: "compare", TargetPin: "a"})
	bp.AddConnection(blueprint.Connection{ID: "d2", SourceNode: "compare", SourcePin: "result", TargetNode: "while", TargetPin: "condition"})
	bp.AddConnection(blueprint.Connection{ID: "d3", SourceNode: "get_a", SourcePin: "value", TargetNode: "add", TargetPin: "a"})
	bp.AddConnection(blueprint.Connection{ID: "d4", SourceNode: "get_b", SourcePin: "value", TargetNode: "add", TargetPin: "b"})
	bp.AddConnection(blueprint.Connection{ID: "d5", SourceNode: "add", SourcePin: "result", TargetNode: "set_temp", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "d6", SourceNode: "get_b", SourcePin: "value", TargetNode: "set_a", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "d7", SourceNode: "get_temp", SourcePin: "value", TargetNode: "set_b", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "d8", SourceNode: "get_i2", SourcePin: "value", TargetNode: "inc_i", TargetPin: "a"})
	bp.AddConnection(blueprint.Connection{ID: "d9", SourceNode: "inc_i", SourcePin: "result", TargetNode: "set_i", TargetPin: "value"})
	bp.AddConnection(blueprint.Connection{ID: "d10", SourceNode: "get_final_b", SourcePin: "value", TargetNode: "end", TargetPin: "result"})

	return bp
}

func main() {
	testCases := []int{10, 20, 30, 40}

	fmt.Println("╔════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Fibonacci 性能基准测试                                   ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-10s │ %-6s │ %-20s │ %-12s │ %-10s ║\n", "语言", "N", "结果", "耗时(ms)", "内存(KB)")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════╣")

	for _, n := range testCases {
		results := []BenchmarkResult{
			benchmarkGo(n),
			benchmarkNodeJS(n),
			benchmarkPython(n),
			benchmarkBlueprint(n),
		}

		for i, r := range results {
			if r.TimeMs < 0 {
				fmt.Printf("║ %-10s │ %-6d │ %-20s │ %-12s │ %-10s ║\n",
					r.Language, r.N, "错误", "-", "-")
			} else {
				resultStr := fmt.Sprintf("%d", r.Result)
				if len(resultStr) > 20 {
					resultStr = resultStr[:17] + "..."
				}
				fmt.Printf("║ %-10s │ %-6d │ %-20s │ %-12.4f │ %-10d ║\n",
					r.Language, r.N, resultStr, r.TimeMs, r.MemoryKB)
			}
			if i == len(results)-1 && n != testCases[len(testCases)-1] {
				fmt.Println("╠────────────┼────────┼──────────────────────┼──────────────┼────────────╣")
			}
		}
	}

	fmt.Println("╚════════════════════════════════════════════════════════════════════════════╝")

	// 保存 JSON 结果
	allResults := make([]BenchmarkResult, 0)
	for _, n := range testCases {
		allResults = append(allResults,
			benchmarkGo(n),
			benchmarkNodeJS(n),
			benchmarkPython(n),
			benchmarkBlueprint(n),
		)
	}
	jsonData, _ := json.MarshalIndent(allResults, "", "  ")
	os.WriteFile("benchmark_results.json", jsonData, 0644)
	fmt.Println("\n结果已保存到 benchmark_results.json")
}
