# 新节点使用指南

本文档介绍如何在前端可视化蓝图编辑器中使用新增的节点功能。

## 新增节点

### 1. Sequence 节点（序列）

**功能**：按顺序执行多个输出分支

**使用场景**：需要依次执行多个操作时

**引脚配置**：
- 输入：`exec_in` (执行)
- 输出：`Then 0`, `Then 1`, `Then 2` (执行)

**示例**：
```
Start → Sequence
         ├─ Then 0 → Print "First"
         ├─ Then 1 → Print "Second"
         └─ Then 2 → Print "Third"
```

**图标**：→
**颜色**：紫色 (#9C27B0)

---

### 2. Switch 节点（分支选择）

**功能**：根据输入值选择对应的执行分支

**使用场景**：替代多层嵌套的 Branch 节点，根据数值或字符串选择分支

**引脚配置**：
- 输入：
  - `exec_in` (执行)
  - `selection` (any) - 选择值
- 输出：
  - `case_0`, `case_1`, `case_2` (执行) - 匹配分支
  - `default` (执行) - 默认分支

**示例**：
```
Start → Make Int (value=1) → Switch (selection)
                              ├─ case_0 → Print "Zero"
                              ├─ case_1 → Print "One" ✓
                              ├─ case_2 → Print "Two"
                              └─ default → Print "Other"
```

**图标**：⚡
**颜色**：紫色 (#9C27B0)

**说明**：
- 数值类型：匹配 `case_0`, `case_1`, `case_2`
- 字符串类型：匹配 `case_<value>`
- 无匹配时执行 `default` 分支

---

### 3. ForEach 节点（数组遍历）

**功能**：遍历数组并对每个元素执行循环体

**使用场景**：处理列表数据，对每个元素执行相同操作

**引脚配置**：
- 输入：
  - `exec_in` (执行)
  - `array` (any[]) - 要遍历的数组
- 输出：
  - `loop_body` (执行) - 循环体（每次迭代）
  - `completed` (执行) - 循环完成
  - `element` (any) - 当前元素
  - `index` (float) - 当前索引

**示例**：
```
Start → Make Array [10, 20, 30] → ForEach
                                   ├─ loop_body → Print element
                                   └─ completed → Print "Done"
```

**输出**：
```
Element: 10
Element: 20
Element: 30
Done
```

**图标**：🔁
**颜色**：紫色 (#9C27B0)

---

### 4. Function 节点（函数调用）

**功能**：调用已注册的函数蓝图

**使用场景**：代码复用，将常用逻辑封装为函数

**引脚配置**（示例：Add 函数）：
- 输入：
  - `exec_in` (执行)
  - `a`, `b` (float) - 函数参数
- 输出：
  - `exec_out` (执行)
  - `sum` (float) - 返回值

**示例**：
```
Start → Make Float (100) → Function: Add (a, b)
        Make Float (200) ↗         ↓
                              Print sum (300)
```

**图标**：ƒ
**颜色**：深紫色 (#673AB7)

**说明**：
- 函数通过 Start 和 End 节点定义输入输出
- 需要先注册函数蓝图到 FunctionRegistry
- 每个函数可以有不同的输入输出配置

---

## 前端实现细节

### 节点模板配置

所有节点配置在 `web/src/config/nodeTemplates.ts` 中定义：

```typescript
{
  type: 'flow_control',
  operation: 'foreach',
  label: 'ForEach',
  description: '遍历数组并对每个元素执行',
  category: '控制流',
  icon: '🔁',
  color: '#9C27B0',
  inputPins: [...],
  outputPins: [...]
}
```

### 类型定义

类型定义在 `web/src/types/blueprint.ts` 中更新：

```typescript
export type NodeType =
  | 'start'
  | 'end'
  | 'arithmetic'
  | 'logic'
  | 'flow_control'  // ← 新增
  | 'data'
  | 'variable'
  | 'debug'
  | 'async'
  | 'function'      // ← 新增
```

### 自动渲染

前端节点渲染器 (`CustomNode.vue`) 已经是模板驱动的，可以自动支持：
- ✅ 任意数量的输入/输出引脚
- ✅ 执行引脚（白色箭头）
- ✅ 数据引脚（彩色圆点）
- ✅ 节点颜色和图标
- ✅ 调试状态显示

---

## 测试蓝图示例

### 示例 1：Sequence + ForEach 组合

```json
{
  "name": "sequence_foreach_demo",
  "nodes": [
    {
      "id": "start",
      "type": "start",
      "label": "Start"
    },
    {
      "id": "seq",
      "type": "flow_control",
      "operation": "sequence",
      "label": "Sequence"
    },
    {
      "id": "make_array",
      "type": "data",
      "operation": "constant",
      "label": "Array [1,2,3]",
      "input_pins": [
        {"name": "value", "kind": "data", "type": "any", "value": [1, 2, 3]}
      ]
    },
    {
      "id": "foreach",
      "type": "flow_control",
      "operation": "foreach",
      "label": "ForEach"
    }
  ],
  "connections": [
    {"source_node": "start", "source_pin": "exec", "target_node": "seq", "target_pin": "exec_in"},
    {"source_node": "seq", "source_pin": "Then 0", "target_node": "foreach", "target_pin": "exec_in"},
    {"source_node": "make_array", "source_pin": "output", "target_node": "foreach", "target_pin": "array"}
  ]
}
```

---

## 与后端对接

前端通过 API 与后端交互：

```typescript
// 执行蓝图
const result = await executeBlueprint(blueprint)

// 检查执行结果
if (result.success) {
  console.log('Outputs:', result.outputs)
  console.log('Variables:', result.variables)
  console.log('Nodes:', result.nodes)
}
```

后端会自动识别并执行所有节点类型，包括新增的：
- `Sequence` → `executeSequenceNode()`
- `Switch` → `executeSwitchNode()`
- `ForEach` → `executeForEach()`
- `Function` → `FunctionCallExecutor`

---

## 节点调色板分类

新节点会自动出现在节点调色板中：

- **控制流** 分类：
  - Branch（分支）
  - For Loop（For 循环）
  - While Loop（While 循环）
  - **Sequence**（序列）← 新增
  - **Switch**（分支选择）← 新增
  - **ForEach**（数组遍历）← 新增

- **函数** 分类：
  - **Function: Add**（函数调用）← 新增

---

## 常见问题

### Q: 如何添加更多 Switch 分支？

修改 `nodeTemplates.ts` 中的 Switch 节点配置，添加更多 `case_N` 输出引脚：

```typescript
outputPins: [
  { name: 'case_0', kind: 'exec', type: 'exec' },
  { name: 'case_1', kind: 'exec', type: 'exec' },
  { name: 'case_2', kind: 'exec', type: 'exec' },
  { name: 'case_3', kind: 'exec', type: 'exec' }, // 新增
  { name: 'default', kind: 'exec', type: 'exec' }
]
```

### Q: 如何创建自定义函数节点？

1. 在后端创建函数蓝图（使用 Start 和 End 节点定义输入输出）
2. 注册到 `FunctionRegistry`
3. 在前端 `nodeTemplates.ts` 中添加对应的函数节点模板

```typescript
{
  type: 'function',
  operation: 'my_custom_func',
  label: '函数：MyFunc',
  category: '函数',
  inputPins: [
    { name: 'exec_in', kind: 'exec', type: 'exec' },
    { name: 'param1', kind: 'data', type: 'float' }
  ],
  outputPins: [
    { name: 'exec_out', kind: 'exec', type: 'exec' },
    { name: 'result', kind: 'data', type: 'float' }
  ]
}
```

### Q: ForEach 支持哪些数组类型？

ForEach 支持所有类型的数组：
- `[]interface{}` - 混合类型数组
- `[]int` - 整数数组
- `[]string` - 字符串数组
- `[]float64` - 浮点数数组

数组会自动转换为 `[]interface{}` 类型进行遍历。

---

## 性能提示

- **Sequence** 节点会按顺序执行所有分支，如果某个分支耗时长，会阻塞后续分支
- **Switch** 节点只执行匹配的分支，比多层嵌套 Branch 更高效
- **ForEach** 受 `MaxIterations` 限制（默认 10000），防止无限循环
- **Function** 调用会创建独立的执行上下文，递归调用需注意深度限制

---

## 后续计划

未来可能添加的节点：
- [ ] Break/Continue 节点（提前退出循环）
- [ ] Select 节点（三元运算符）
- [ ] Gate 节点（门控执行）
- [ ] Macro 节点（宏展开）
- [ ] 更多内置函数节点

---

更新日期：2025-01-23
版本：v1.1.0
