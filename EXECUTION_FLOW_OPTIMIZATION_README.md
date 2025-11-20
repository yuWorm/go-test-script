# 执行流优化总结

## 概述

本次优化将蓝图引擎改造为**图形化脚本语言**，类似UE5蓝图但更符合编程逻辑。核心理念：**从Start节点开始，通过执行引脚和数据引脚控制执行流，孤立节点不执行**。

## 核心改进

### 1. 执行流逻辑重构

#### 入口点
- **Start节点作为唯一入口**：每个蓝图必须从Start节点开始
- 类似函数入口点，保证执行顺序可预测

#### 可达性检测
- 使用**广度优先搜索(BFS)**从Start节点遍历
- 只执行从Start可达的节点
- **孤立节点自动过滤**：与Start无连接的节点不会执行

#### 智能依赖分析

```go
// 3种节点激活方式：
1. 执行引脚驱动：有exec_in的节点必须通过执行流激活
2. 数据引脚驱动：无exec_in的节点可通过数据依赖自动激活
3. 混合模式：支持执行流和数据流同时工作
```

**规则**：
- 节点定义了`exec_in`引脚 → 必须通过执行连接激活
- 节点没有`exec_in`引脚 → 可通过数据连接自动激活
- 节点有`exec_in`但未连接 → 不会执行

### 2. 执行流图构建

```
连接类型分离：
├── execFlowMap: 执行引脚连接（白色箭头）
├── dataFlowMap: 数据引脚连接（彩色圆点）
└── needsExecActivation: 需要执行激活的节点标记
```

### 3. 执行顺序

```
Start节点
  ↓ (exec连接)
节点A (执行)
  ↓ (exec连接)
节点B (执行)
  ├→ (exec: true_exec) 节点C (执行)
  └→ (exec: false_exec) 节点D (不执行)
```

## Web UI可视化

### 执行引脚 vs 数据引脚

| 特性 | 执行引脚 | 数据引脚 |
|------|---------|---------|
| 图标 | ▶ 白色箭头 | ● 彩色圆点 |
| 用途 | 控制执行流程 | 传递数据值 |
| 形状 | 箭头形状 | 圆形 |
| 颜色 | 白色/黄色(hover) | 按类型区分 |
| 样式类 | `pin-handle-exec` | `pin-handle-{type}` |

### 数据引脚颜色方案

```css
float/number → 绿色 (#4CAF50)
string → 粉色 (#E91E63)
bool → 红色 (#F44336)
any → 紫色 (#9C27B0)
```

### 节点模板更新

#### Start节点
```javascript
{
  outputPins: [
    { name: 'exec', kind: 'exec', type: 'exec' } // 执行输出
  ]
}
```

#### End节点
```javascript
{
  inputPins: [
    { name: 'exec_in', kind: 'exec', type: 'exec' } // 执行输入
  ]
}
```

#### 函数节点（如Add）
```javascript
{
  inputPins: [
    { name: 'exec_in', kind: 'exec', type: 'exec' },
    { name: 'a', kind: 'data', type: 'float' },
    { name: 'b', kind: 'data', type: 'float' }
  ],
  outputPins: [
    { name: 'exec_out', kind: 'exec', type: 'exec' },
    { name: 'result', kind: 'data', type: 'float' }
  ]
}
```

#### Branch节点
```javascript
{
  inputPins: [
    { name: 'exec_in', kind: 'exec', type: 'exec' },
    { name: 'condition', kind: 'data', type: 'bool' }
  ],
  outputPins: [
    { name: 'true_exec', kind: 'exec', type: 'exec' },  // condition=true时激活
    { name: 'false_exec', kind: 'exec', type: 'exec' }  // condition=false时激活
  ]
}
```

## 使用示例

### 1. 纯执行流示例

```
Start(exec) → Add(exec_in/exec_out) → End(exec_in)
```

特点：通过执行引脚显式控制顺序

### 2. 执行流+数据流混合

```
Start(exec) → Add(exec_in, a, b) → Multiply(a) → End
              ↓(data: result)        ↑(data)
              └─────────────────────┘
```

特点：执行流控制顺序，数据流传递值

### 3. 条件分支

```
Start → Add → Branch(condition)
              ├─ true_exec → Multiply → End
              └─ false_exec → Subtract → End
```

特点：根据条件只激活一个分支

### 4. 孤立节点（不执行）

```
Start → Add → End

[Isolated]
Multiply (无连接，不执行)
```

## 测试验证

### 测试1: 孤立节点检测

```bash
go run test_isolated_nodes.go
```

**结果**：
```
✓ start (开始): 已执行
✓ add (加法（连接）): 已执行
✗ isolated_multiply (乘法（孤立）): 未执行 (预期)
✓ end (结束): 已执行
✓ 孤立节点正确地未执行
```

### 测试2: 执行流分支

```bash
go run test_execution_pins.go
```

**结果**：
```
✓ start: 已执行
✓ add: 已执行
✓ branch: 已执行
✓ multiply (True分支): 已执行
✗ subtract (False分支): 未执行 (预期)
✓ end: 已执行
最终结果: 30 (预期值)
```

## 技术实现

### executeWithExecutionFlow 函数

```go
func (e *Executor) executeWithExecutionFlow(ctx *ExecutionContext, bp *Blueprint) error {
    // 1. 找到Start节点
    // 2. 构建执行流图和数据流图
    // 3. 检测需要执行激活的节点
    // 4. BFS从Start开始执行
    // 5. 根据连接类型决定下一步执行的节点
}
```

**执行策略**：
1. **执行引脚优先**：遇到exec连接，检查shouldActivateExecPin
2. **数据引脚补充**：对于无exec_in的节点，数据连接可触发执行
3. **去重机制**：executed map防止节点重复执行
4. **队列管理**：BFS队列保证执行顺序

### shouldActivateExecPin 函数

```go
func (e *Executor) shouldActivateExecPin(node *Node, execPinName string) bool {
    // Branch节点：检查输出值决定激活哪个分支
    // 普通节点：总是激活所有执行输出
}
```

## 性能优化

1. **避免全图遍历**：只执行可达节点
2. **早期终止**：检测到错误时可立即停止
3. **并发安全**：使用map跟踪执行状态
4. **内存效率**：执行完即释放，不保留无用状态

## 与UE5蓝图的对比

| 特性 | 本引擎 | UE5蓝图 |
|------|-------|---------|
| 执行流 | ✅ 白色箭头 | ✅ 白色线 |
| 数据流 | ✅ 彩色圆点 | ✅ 彩色线 |
| 入口点 | Start节点 | Event节点 |
| 分支 | Branch节点 | Branch节点 |
| 孤立节点 | 自动过滤 | 警告但存在 |
| 纯数据节点 | ✅ 支持 | ✅ 支持 |
| 混合驱动 | ✅ 执行+数据 | ✅ 执行+数据 |

## 文件变更

### 后端
- `blueprint/executor.go`: 重写executeWithExecutionFlow

### 前端
- `web/src/components/CustomNode.vue`: 执行引脚可视化
- `web/src/config/nodeTemplates.ts`: 添加执行引脚定义

### 测试
- `test_isolated_nodes.go`: 孤立节点测试
- `test_execution_pins.go`: 执行流测试

## 下一步改进

- [ ] 更多执行流节点（Sequence, ForEach, While等）
- [ ] 调试模式：高亮当前执行的节点
- [ ] 性能分析：统计每个节点的执行时间
- [ ] 断点支持：在特定节点暂停执行
- [ ] 可视化执行路径：显示实际执行的节点路径

## 总结

本次优化将蓝图引擎转变为真正的**图形化脚本语言**：

✅ **明确的执行入口**：Start节点作为函数入口
✅ **可视化的执行流**：白色箭头清晰表示执行顺序
✅ **智能的节点激活**：执行流和数据流混合驱动
✅ **自动的优化**：孤立节点不执行，提高效率
✅ **完整的UI支持**：执行引脚和数据引脚可视化区分

现在用户可以像编写代码一样使用蓝图，既有图形化的直观性，又有编程语言的精确性！
