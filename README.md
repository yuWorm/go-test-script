# Golang Blueprint Engine

一个高性能、并发安全的蓝图可视化编辑器执行引擎，类似于虚幻引擎（UE）的蓝图系统。使用 JSON 格式存储蓝图，支持预编译和高效执行。

## 特性

- ✅ **JSON 存储格式** - 蓝图以 JSON 格式存储，易于编辑和版本控制
- ✅ **预编译优化** - 解析 JSON 后预编译为可执行对象，提升执行性能
- ✅ **高性能执行** - 支持顺序和并行两种执行模式
- ✅ **并发安全** - 所有操作都是线程安全的，支持并发执行
- ✅ **拓扑排序** - 自动分析节点依赖关系，优化执行顺序
- ✅ **循环检测** - 自动检测并拒绝包含循环依赖的蓝图
- ✅ **类型安全** - 节点输入输出类型定义清晰
- ✅ **可扩展** - 易于添加自定义节点类型
- ✅ **内置节点** - 提供算术、逻辑、条件等常用节点

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                    Blueprint JSON                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                      Parser                              │
│  - JSON 反序列化                                         │
│  - 结构验证                                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                     Compiler                             │
│  - 拓扑排序（Kahn 算法）                                 │
│  - 依赖分析                                              │
│  - 循环检测                                              │
│  - 执行器关联                                            │
│  - 连接映射构建                                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                Compiled Blueprint                        │
│  - 优化的执行顺序                                        │
│  - 预关联的执行器                                        │
│  - 快速连接查找                                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                     Executor                             │
│  - 顺序执行模式                                          │
│  - 并行执行模式（带依赖管理）                            │
│  - 数据流传递                                            │
│  - 错误处理                                              │
│  - 超时控制                                              │
└─────────────────────────────────────────────────────────┘
```

## 快速开始

### 安装

```bash
go get github.com/go-blueprint-engine
```

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/go-blueprint-engine/blueprint"
    "github.com/go-blueprint-engine/nodes"
)

func main() {
    // 1. 创建节点注册表并注册内置节点
    registry := blueprint.NewNodeRegistry()
    nodes.RegisterAllNodes(registry)

    // 2. 从 JSON 文件解析蓝图
    parser := blueprint.NewParser()
    bp, err := parser.ParseFromFile("examples/simple_calculator.json")
    if err != nil {
        panic(err)
    }

    // 3. 编译蓝图
    compiler := blueprint.NewCompiler(registry)
    if err := compiler.Compile(bp); err != nil {
        panic(err)
    }

    // 4. 执行蓝图
    executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
    result, err := executor.Execute(bp, nil)
    if err != nil {
        panic(err)
    }

    // 5. 获取结果
    if result.Success {
        fmt.Printf("执行成功! 耗时: %v\n", result.Duration)
        for key, value := range result.Outputs {
            fmt.Printf("%s = %v\n", key, value)
        }
    }
}
```

### 程序化创建蓝图

```go
// 创建蓝图
bp := blueprint.NewBlueprint("my_blueprint")
bp.Description = "计算 (a + b) * c"

// 创建节点
nodeA := &blueprint.Node{
    ID:    "const_a",
    Type:  blueprint.NodeTypeFunction,
    Operation: "constant",
    InputPins: []blueprint.Pin{
        {Name: "value", Type: "float", Value: 10.0},
    },
    OutputPins: []blueprint.Pin{
        {Name: "output", Type: "float"},
    },
}

nodeB := &blueprint.Node{
    ID:    "const_b",
    Type:  blueprint.NodeTypeFunction,
    Operation: "constant",
    InputPins: []blueprint.Pin{
        {Name: "value", Type: "float", Value: 20.0},
    },
    OutputPins: []blueprint.Pin{
        {Name: "output", Type: "float"},
    },
}

addNode := &blueprint.Node{
    ID:    "add",
    Type:  blueprint.NodeTypeArithmetic,
    Operation: "add",
    InputPins: []blueprint.Pin{
        {Name: "a", Type: "float"},
        {Name: "b", Type: "float"},
    },
    OutputPins: []blueprint.Pin{
        {Name: "result", Type: "float"},
    },
}

// 添加节点到蓝图
bp.AddNode(nodeA)
bp.AddNode(nodeB)
bp.AddNode(addNode)

// 添加连接
bp.AddConnection(blueprint.Connection{
    ID: "conn1",
    SourceNode: "const_a",
    SourcePin: "output",
    TargetNode: "add",
    TargetPin: "a",
})

bp.AddConnection(blueprint.Connection{
    ID: "conn2",
    SourceNode: "const_b",
    SourcePin: "output",
    TargetNode: "add",
    TargetPin: "b",
})

// 保存为 JSON
serializer := blueprint.NewSerializer(true)
serializer.ToFile(bp, "my_blueprint.json")
```

## 蓝图 JSON 格式

```json
{
  "name": "Simple Calculator",
  "version": "1.0.0",
  "description": "A simple calculator blueprint",
  "nodes": [
    {
      "id": "const_a",
      "type": "function",
      "operation": "constant",
      "label": "Constant A",
      "input_pins": [
        {
          "name": "value",
          "type": "float",
          "value": 10
        }
      ],
      "output_pins": [
        {
          "name": "output",
          "type": "float"
        }
      ],
      "position": {
        "x": 100,
        "y": 100
      }
    },
    {
      "id": "const_b",
      "type": "function",
      "operation": "constant",
      "label": "Constant B",
      "input_pins": [
        {
          "name": "value",
          "type": "float",
          "value": 20
        }
      ],
      "output_pins": [
        {
          "name": "output",
          "type": "float"
        }
      ],
      "position": {
        "x": 100,
        "y": 200
      }
    },
    {
      "id": "add_node",
      "type": "arithmetic",
      "operation": "add",
      "label": "Add",
      "input_pins": [
        {
          "name": "a",
          "type": "float"
        },
        {
          "name": "b",
          "type": "float"
        }
      ],
      "output_pins": [
        {
          "name": "result",
          "type": "float"
        }
      ],
      "position": {
        "x": 300,
        "y": 150
      }
    }
  ],
  "connections": [
    {
      "id": "conn1",
      "source_node": "const_a",
      "source_pin": "output",
      "target_node": "add_node",
      "target_pin": "a"
    },
    {
      "id": "conn2",
      "source_node": "const_b",
      "source_pin": "output",
      "target_node": "add_node",
      "target_pin": "b"
    }
  ],
  "variables": {},
  "metadata": {
    "author": "Blueprint Engine",
    "created": "2025-11-20"
  }
}
```

## 内置节点类型

### 算术运算节点 (arithmetic)

- `add` - 加法 (a + b)
- `subtract` - 减法 (a - b)
- `multiply` - 乘法 (a * b)
- `divide` - 除法 (a / b)
- `modulo` - 取模 (a % b)
- `power` - 幂运算 (a ^ b)

### 比较运算节点 (arithmetic)

- `equal` - 等于 (a == b)
- `not_equal` - 不等于 (a != b)
- `greater` - 大于 (a > b)
- `greater_equal` - 大于等于 (a >= b)
- `less` - 小于 (a < b)
- `less_equal` - 小于等于 (a <= b)

### 逻辑运算节点 (logic)

- `and` - 逻辑与 (a && b)
- `or` - 逻辑或 (a || b)
- `not` - 逻辑非 (!a)
- `xor` - 逻辑异或 (a XOR b)

### 控制流节点 (condition)

- 条件分支 - 根据条件选择不同的值

### 函数节点 (function)

- `constant` - 常量值
- `get_variable` - 获取变量
- `set_variable` - 设置变量

### 特殊节点

- `start` - 开始节点
- `end` - 结束节点

## 执行模式

### 顺序执行模式

按照拓扑排序的顺序依次执行所有节点。

```go
options := &blueprint.ExecutionOptions{
    Mode:        blueprint.ExecutionModeSequential,
    StopOnError: true,
}
executor := blueprint.NewExecutor(options)
```

### 并行执行模式

在保证依赖关系的前提下，尽可能并行执行无依赖的节点。

```go
options := &blueprint.ExecutionOptions{
    Mode:           blueprint.ExecutionModeParallel,
    MaxConcurrency: 4,        // 最大并发数
    StopOnError:    true,
}
executor := blueprint.NewExecutor(options)
```

## 自定义节点

### 1. 实现 NodeExecutor 接口

```go
type MyCustomExecutor struct{}

func (e *MyCustomExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
    // 获取输入
    a := inputs["a"].(float64)
    b := inputs["b"].(float64)

    // 执行自定义逻辑
    result := a * a + b * b

    // 返回输出
    return map[string]interface{}{
        "result": result,
    }, nil
}

func (e *MyCustomExecutor) Validate(node *blueprint.Node) error {
    // 验证节点配置
    if len(node.InputPins) < 2 {
        return fmt.Errorf("requires at least 2 input pins")
    }
    return nil
}
```

### 2. 注册自定义节点

```go
registry := blueprint.NewNodeRegistry()
registry.Register("custom", "my_operation", &MyCustomExecutor{})
```

## 性能特点

1. **预编译优化** - 蓝图在执行前会被编译，生成优化的执行计划
2. **拓扑排序** - 使用 Kahn 算法进行拓扑排序，确保最优执行顺序
3. **并发安全** - 使用 sync.RWMutex 和 sync.Cond 确保线程安全
4. **零拷贝** - 节点间数据传递通过引用，避免不必要的复制
5. **缓存优化** - 节点输入输出值被缓存，避免重复计算

## 测试

运行所有测试：

```bash
go test -v ./...
```

运行性能测试：

```bash
go test -bench=. -benchmem ./...
```

## 示例

查看 `examples/` 目录中的示例：

- `simple_calculator.json` - 简单计算器示例
- `conditional_logic.json` - 条件逻辑示例
- 运行 `go run main.go` 查看所有示例的执行结果

## 项目结构

```
.
├── blueprint/              # 核心引擎代码
│   ├── node.go            # 节点定义
│   ├── blueprint.go       # 蓝图定义
│   ├── parser.go          # JSON 解析器
│   ├── compiler.go        # 编译器
│   ├── executor.go        # 执行引擎
│   ├── context.go         # 执行上下文
│   └── blueprint_test.go  # 测试文件
├── nodes/                 # 内置节点实现
│   ├── arithmetic.go      # 算术节点
│   ├── logic.go           # 逻辑节点
│   ├── control.go         # 控制流节点
│   └── registry.go        # 节点注册
├── examples/              # 示例蓝图
│   ├── simple_calculator.json
│   └── conditional_logic.json
├── main.go               # 演示程序
├── go.mod
└── README.md
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 技术栈

- Go 1.21+
- 标准库（无外部依赖）
- JSON 序列化/反序列化
- 并发编程（goroutines, channels, sync）
- 图算法（拓扑排序，环检测）

## 未来计划

- [ ] 可视化编辑器前端（Web UI）
- [ ] 更多内置节点类型
- [ ] 子蓝图支持
- [ ] 蓝图调试器
- [ ] 性能分析工具
- [ ] 蓝图版本管理
- [ ] WebAssembly 支持

## 联系方式

如有问题或建议，请创建 Issue 或 Pull Request。
