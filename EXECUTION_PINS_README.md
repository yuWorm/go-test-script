# 执行引脚功能说明

## 概述

本项目现已支持执行引脚（Execution Pins），类似UE5蓝图编辑器的执行流控制系统。执行引脚让节点的执行顺序更加可视化和可控。

## 引脚类型

### PinKind

引脚现在分为两种类型：

- **execution (exec)**: 执行引脚，用白色箭头表示，控制节点的执行流
- **data**: 数据引脚，用彩色圆点表示，传递数据值

### 类型验证

- 执行引脚只能连接到执行引脚
- 数据引脚只能连接到数据引脚
- 不同类型的引脚之间无法连接

## 执行模式

### ExecutionModeExecutionFlow

新增的执行流模式：

```go
executor := blueprint.NewExecutor(&blueprint.ExecutionOptions{
    Mode:        blueprint.ExecutionModeExecutionFlow,
    StopOnError: true,
})
```

特点：
- 只执行被执行引脚连接到的节点
- 从入口节点开始，沿着执行引脚连接传播
- 支持条件分支（如Branch节点）

## 节点执行引脚

### Start节点

**输出引脚**:
- `exec` (execution): 执行输出，总是激活

**示例**:
```json
{
  "id": "start",
  "type": "start",
  "output_pins": [
    {"name": "exec", "kind": "exec", "type": "exec"},
    {"name": "value", "kind": "data", "type": "float", "value": 42}
  ]
}
```

### End节点

**输入引脚**:
- `exec_in` (execution): 执行输入

**输出**:
- `executed`: true（标记节点已执行）
- 所有数据输入会被传递到输出

**示例**:
```json
{
  "id": "end",
  "type": "end",
  "input_pins": [
    {"name": "exec_in", "kind": "exec", "type": "exec"},
    {"name": "result", "kind": "data", "type": "float"}
  ]
}
```

### 函数节点（如算术节点）

**输入引脚**:
- `exec_in` (execution): 执行输入（可选）

**输出引脚**:
- `exec_out` (execution): 执行输出（如果有exec_in）
- 数据输出（如result）

**示例**:
```json
{
  "id": "add",
  "type": "arithmetic",
  "operation": "add",
  "input_pins": [
    {"name": "exec_in", "kind": "exec", "type": "exec"},
    {"name": "a", "kind": "data", "type": "float"},
    {"name": "b", "kind": "data", "type": "float"}
  ],
  "output_pins": [
    {"name": "exec_out", "kind": "exec", "type": "exec"},
    {"name": "result", "kind": "data", "type": "float"}
  ]
}
```

### Branch节点

**输入引脚**:
- `exec_in` (execution): 执行输入
- `condition` (data, bool): 条件

**输出引脚**:
- `true_exec` (execution): 条件为true时激活
- `false_exec` (execution): 条件为false时激活

**特点**: 根据条件值，只有一个执行输出引脚会被激活

**示例**:
```json
{
  "id": "branch",
  "type": "function",
  "operation": "branch",
  "input_pins": [
    {"name": "exec_in", "kind": "exec", "type": "exec"},
    {"name": "condition", "kind": "data", "type": "bool", "value": true}
  ],
  "output_pins": [
    {"name": "true_exec", "kind": "exec", "type": "exec"},
    {"name": "false_exec", "kind": "exec", "type": "exec"}
  ]
}
```

## 执行流程

### 1. 入口节点识别

系统自动识别入口节点：
- Start节点
- 有执行输出但没有执行输入的节点

### 2. 执行流传播

从入口节点开始：
1. 执行当前节点
2. 检查节点的执行输出引脚
3. 对于普通节点，激活所有执行输出
4. 对于Branch节点，根据条件只激活一个分支
5. 递归执行所有目标节点

### 3. 分支处理

Branch节点根据condition输出：
- condition=true: 激活`true_exec`，执行True分支的节点
- condition=false: 激活`false_exec`，执行False分支的节点

未被激活的分支上的节点不会执行。

## 示例

### 基础示例

```
Start -> Add -> End
```

创建JSON:
```json
{
  "connections": [
    {"source_node": "start", "source_pin": "exec", "target_node": "add", "target_pin": "exec_in"},
    {"source_node": "add", "source_pin": "exec_out", "target_node": "end", "target_pin": "exec_in"}
  ]
}
```

### 分支示例

```
Start -> Add -> Branch
              ├─ True -> Multiply -> End
              └─ False -> Subtract -> End
```

见 `examples/execution_pins_demo.json` 完整示例。

**执行结果**:
- condition=true: 执行路径为 Start -> Add -> Branch -> Multiply -> End，结果30
- condition=false: 执行路径为 Start -> Add -> Branch -> Subtract -> End，结果10

## 技术实现

### 执行流图构建

```go
// execFlowMap: nodeID -> execOutputPin -> [(targetNodeID, targetExecInputPin)]
execFlowMap := make(map[string]map[string][]*execFlowTarget)
```

### 执行流追踪

```go
func executeNodeAndFollowExecFlow(
    ctx *ExecutionContext,
    bp *Blueprint,
    node *Node,
    execFlowMap map[string]map[string][]*execFlowTarget,
) error
```

### 激活检查

```go
func shouldActivateExecPin(node *Node, execPinName string) bool {
    // Branch节点：检查条件输出
    // 普通节点：总是激活
}
```

## 测试

运行执行引脚测试：

```bash
# 基础测试
go run test_exec_flow_trace.go

# 完整分支测试
go run test_execution_pins.go

# 调试模式测试
go run test_execution_pins_debug.go
```

**测试结果**:
```
✓ Start -> Add -> Branch -> Multiply -> End
✓ Subtract节点未执行（False分支）
✓ 最终结果: 30 (预期值)
```

## 向后兼容

- 未指定Kind的引脚默认为`data`类型
- 现有蓝图仍可使用顺序/并行执行模式
- 执行流模式为可选功能

## 未来改进

- [ ] Web UI支持执行引脚可视化
- [ ] 更多控制流节点（Sequence, Select等）
- [ ] 执行引脚样式定制
- [ ] 异步执行流支持

## 贡献

欢迎提交PR改进执行引脚功能！
