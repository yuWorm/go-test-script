import type { NodeTemplate } from '../types/blueprint'

/**
 * 所有可用的节点模板
 */
export const nodeTemplates: NodeTemplate[] = [
  // 算术运算节点
  {
    type: 'arithmetic',
    operation: 'add',
    label: '加法',
    description: '计算两个数的和 (a + b)',
    category: '算术运算',
    icon: '+',
    color: '#4CAF50',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float' },
      { name: 'b', kind: 'data', type: 'float' }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'float' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'subtract',
    label: '减法',
    description: '计算两个数的差 (a - b)',
    category: '算术运算',
    icon: '-',
    color: '#4CAF50',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'float' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'multiply',
    label: '乘法',
    description: '计算两个数的积 (a × b)',
    category: '算术运算',
    icon: '×',
    color: '#4CAF50',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'float' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'divide',
    label: '除法',
    description: '计算两个数的商 (a ÷ b)',
    category: '算术运算',
    icon: '÷',
    color: '#4CAF50',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'float' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'power',
    label: '幂运算',
    description: '计算 a 的 b 次方',
    category: '算术运算',
    icon: '^',
    color: '#4CAF50',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'float' }
    ]
  },

  // 比较运算节点
  {
    type: 'arithmetic',
    operation: 'greater',
    label: '大于',
    description: '比较 a 是否大于 b',
    category: '比较运算',
    icon: '>',
    color: '#FF9800',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'less',
    label: '小于',
    description: '比较 a 是否小于 b',
    category: '比较运算',
    icon: '<',
    color: '#FF9800',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'equal',
    label: '等于',
    description: '比较 a 是否等于 b',
    category: '比较运算',
    icon: '=',
    color: '#FF9800',
    inputPins: [
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },

  // 逻辑运算节点
  {
    type: 'logic',
    operation: 'and',
    label: '逻辑与',
    description: 'a AND b',
    category: '逻辑运算',
    icon: '&&',
    color: '#2196F3',
    inputPins: [
      { name: 'a', type: 'bool' },
      { name: 'b', type: 'bool' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },
  {
    type: 'logic',
    operation: 'or',
    label: '逻辑或',
    description: 'a OR b',
    category: '逻辑运算',
    icon: '||',
    color: '#2196F3',
    inputPins: [
      { name: 'a', type: 'bool' },
      { name: 'b', type: 'bool' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },
  {
    type: 'logic',
    operation: 'not',
    label: '逻辑非',
    description: 'NOT a',
    category: '逻辑运算',
    icon: '!',
    color: '#2196F3',
    inputPins: [
      { name: 'a', type: 'bool' }
    ],
    outputPins: [
      { name: 'result', type: 'bool' }
    ]
  },

  // 控制流节点
  {
    type: 'condition',
    label: '条件分支',
    description: '根据条件选择不同的值',
    category: '控制流',
    icon: '?',
    color: '#9C27B0',
    inputPins: [
      { name: 'condition', type: 'bool' },
      { name: 'true_value', type: 'any' },
      { name: 'false_value', type: 'any' }
    ],
    outputPins: [
      { name: 'result', type: 'any' },
      { name: 'is_true', type: 'bool' },
      { name: 'is_false', type: 'bool' }
    ]
  },
  {
    type: 'condition',
    operation: 'if_else',
    label: 'If/Else',
    description: 'If 条件判断，执行不同分支',
    category: '控制流',
    icon: 'IF',
    color: '#9C27B0',
    inputPins: [
      { name: 'condition', type: 'bool' },
      { name: 'then_value', type: 'any' },
      { name: 'else_value', type: 'any' }
    ],
    outputPins: [
      { name: 'result', type: 'any' },
      { name: 'then_exec', type: 'bool' },
      { name: 'else_exec', type: 'bool' }
    ]
  },
  {
    type: 'condition',
    operation: 'for_loop',
    label: 'For 循环',
    description: '从 start 到 end 循环，步长为 step',
    category: '控制流',
    icon: 'FOR',
    color: '#9C27B0',
    inputPins: [
      { name: 'start', type: 'float', value: 0 },
      { name: 'end', type: 'float', value: 10 },
      { name: 'step', type: 'float', value: 1 }
    ],
    outputPins: [
      { name: 'index', type: 'float' },
      { name: 'count', type: 'float' },
      { name: 'completed', type: 'bool' }
    ]
  },
  {
    type: 'condition',
    operation: 'while_loop',
    label: 'While 循环',
    description: '当条件为真时循环执行',
    category: '控制流',
    icon: 'WHILE',
    color: '#9C27B0',
    inputPins: [
      { name: 'condition', type: 'bool', value: true }
    ],
    outputPins: [
      { name: 'iterations', type: 'float' },
      { name: 'completed', type: 'bool' }
    ]
  },
  {
    type: 'condition',
    operation: 'break',
    label: 'Break',
    description: '跳出循环',
    category: '控制流',
    icon: '⊗',
    color: '#E91E63',
    inputPins: [],
    outputPins: [
      { name: 'break', type: 'bool' }
    ]
  },
  {
    type: 'condition',
    operation: 'continue',
    label: 'Continue',
    description: '继续下一次循环',
    category: '控制流',
    icon: '↻',
    color: '#E91E63',
    inputPins: [],
    outputPins: [
      { name: 'continue', type: 'bool' }
    ]
  },

  // 函数节点
  {
    type: 'function',
    operation: 'constant',
    label: '常量',
    description: '输出一个常量值',
    category: '函数',
    icon: '123',
    color: '#607D8B',
    inputPins: [
      { name: 'value', type: 'any', value: 0 }
    ],
    outputPins: [
      { name: 'output', type: 'any' }
    ]
  },
  {
    type: 'function',
    operation: 'get_variable',
    label: '获取变量',
    description: '从上下文获取变量值',
    category: '函数',
    icon: 'VAR',
    color: '#607D8B',
    inputPins: [
      { name: 'name', type: 'string' }
    ],
    outputPins: [
      { name: 'value', type: 'any' }
    ]
  },
  {
    type: 'function',
    operation: 'set_variable',
    label: '设置变量',
    description: '设置变量到上下文',
    category: '函数',
    icon: 'SET',
    color: '#607D8B',
    inputPins: [
      { name: 'name', type: 'string' },
      { name: 'value', type: 'any' }
    ],
    outputPins: [
      { name: 'value', type: 'any' }
    ]
  },
  {
    type: 'function',
    operation: 'sleep',
    label: '休眠',
    description: '休眠指定毫秒数（模拟耗时操作）',
    category: '函数',
    icon: '💤',
    color: '#607D8B',
    inputPins: [
      { name: 'duration', type: 'float', value: 1000 }
    ],
    outputPins: [
      { name: 'duration', type: 'float' },
      { name: 'done', type: 'bool' }
    ]
  },

  // 异步执行节点
  {
    type: 'function',
    operation: 'async_start',
    label: '异步开始',
    description: '创建异步任务，后续节点异步执行',
    category: '异步',
    icon: '🚀',
    color: '#00BCD4',
    inputPins: [
      { name: 'task_id', type: 'string', value: '' }
    ],
    outputPins: [
      { name: 'task_id', type: 'string' },
      { name: 'started', type: 'bool' }
    ]
  },
  {
    type: 'function',
    operation: 'async_end',
    label: '异步结束',
    description: '结束异步任务，返回结果',
    category: '异步',
    icon: '🏁',
    color: '#00BCD4',
    inputPins: [
      { name: 'task_id', type: 'string' },
      { name: 'result', type: 'any' }
    ],
    outputPins: [
      { name: 'task_id', type: 'string' },
      { name: 'completed', type: 'bool' },
      { name: 'result', type: 'any' }
    ]
  },
  {
    type: 'function',
    operation: 'await',
    label: '等待结果',
    description: '等待异步任务完成并获取结果',
    category: '异步',
    icon: '⏳',
    color: '#00BCD4',
    inputPins: [
      { name: 'task_id', type: 'string' },
      { name: 'timeout', type: 'float', value: 30 }
    ],
    outputPins: [
      { name: 'result', type: 'any' },
      { name: 'completed', type: 'bool' }
    ]
  },
  {
    type: 'function',
    operation: 'parallel',
    label: '并行等待',
    description: '等待多个异步任务完成',
    category: '异步',
    icon: '⚡',
    color: '#00BCD4',
    inputPins: [
      { name: 'task_id_1', type: 'string' },
      { name: 'task_id_2', type: 'string' }
    ],
    outputPins: [
      { name: 'result_0', type: 'any' },
      { name: 'result_1', type: 'any' },
      { name: 'count', type: 'float' },
      { name: 'all_completed', type: 'bool' }
    ]
  },

  // 执行流控制节点
  {
    type: 'function',
    operation: 'branch',
    label: '分支',
    description: '根据条件选择执行分支',
    category: '执行流',
    icon: '◆',
    color: '#FF5722',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'condition', kind: 'data', type: 'bool', value: true }
    ],
    outputPins: [
      { name: 'true_exec', kind: 'exec', type: 'exec' },
      { name: 'false_exec', kind: 'exec', type: 'exec' }
    ]
  },

  // 特殊节点
  {
    type: 'start',
    label: '开始',
    description: '蓝图开始节点（可添加输出引脚作为蓝图输入）',
    category: '特殊',
    icon: '▶',
    color: '#8BC34A',
    inputPins: [],
    outputPins: [
      { name: 'exec', kind: 'exec', type: 'exec' }
    ]
  },
  {
    type: 'end',
    label: '结束',
    description: '蓝图结束节点（可添加输入引脚作为蓝图输出）',
    category: '特殊',
    icon: '■',
    color: '#F44336',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' }
    ],
    outputPins: []
  }
]

/**
 * 根据类型和操作获取节点模板
 */
export function getNodeTemplate(type: string, operation?: string): NodeTemplate | undefined {
  return nodeTemplates.find(t =>
    t.type === type && (operation ? t.operation === operation : !t.operation)
  )
}

/**
 * 获取所有分类
 */
export function getCategories(): string[] {
  const categories = new Set(nodeTemplates.map(t => t.category))
  return Array.from(categories)
}

/**
 * 根据分类获取节点
 */
export function getNodesByCategory(category: string): NodeTemplate[] {
  return nodeTemplates.filter(t => t.category === category)
}
