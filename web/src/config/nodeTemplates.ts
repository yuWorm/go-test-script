import type { NodeTemplate } from '../types/blueprint'

/**
 * 所有可用的节点模板
 *
 * 引脚设计原则：
 * 1. 可执行节点：有 exec_in 和 exec_out（或自定义执行输出引脚）
 * 2. 纯数据节点：没有执行引脚，只有数据引脚（如常量、获取变量）
 * 3. 控制流节点：有自定义执行输出引脚（如 true_exec, false_exec, loop_body）
 */
export const nodeTemplates: NodeTemplate[] = [
  // ==================== 算术运算节点（可执行） ====================
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
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'float' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'float' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 1 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'float' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 1 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'float' }
    ]
  },

  // ==================== 比较运算节点（可执行） ====================
  {
    type: 'arithmetic',
    operation: 'greater',
    label: '大于',
    description: '比较 a 是否大于 b',
    category: '比较运算',
    icon: '>',
    color: '#FF9800',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'not_equal',
    label: '不等于',
    description: '比较 a 是否不等于 b',
    category: '比较运算',
    icon: '≠',
    color: '#FF9800',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'greater_equal',
    label: '大于等于',
    description: '比较 a 是否大于等于 b',
    category: '比较运算',
    icon: '≥',
    color: '#FF9800',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },
  {
    type: 'arithmetic',
    operation: 'less_equal',
    label: '小于等于',
    description: '比较 a 是否小于等于 b',
    category: '比较运算',
    icon: '≤',
    color: '#FF9800',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'float', value: 0 },
      { name: 'b', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },

  // ==================== 逻辑运算节点（可执行） ====================
  {
    type: 'logic',
    operation: 'and',
    label: '逻辑与',
    description: 'a AND b',
    category: '逻辑运算',
    icon: '&&',
    color: '#2196F3',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'bool', value: false },
      { name: 'b', kind: 'data', type: 'bool', value: false }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'bool', value: false },
      { name: 'b', kind: 'data', type: 'bool', value: false }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
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
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'bool', value: false }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },
  {
    type: 'logic',
    operation: 'xor',
    label: '异或',
    description: 'a XOR b',
    category: '逻辑运算',
    icon: '⊕',
    color: '#2196F3',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'a', kind: 'data', type: 'bool', value: false },
      { name: 'b', kind: 'data', type: 'bool', value: false }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'bool' }
    ]
  },

  // ==================== 控制流节点（自定义执行引脚） ====================
  {
    type: 'flow_control',
    operation: 'branch',
    label: '分支',
    description: '根据条件选择执行分支（true/false）',
    category: '控制流',
    icon: '◆',
    color: '#9C27B0',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'condition', kind: 'data', type: 'bool', value: true }
    ],
    outputPins: [
      { name: 'true_exec', kind: 'exec', type: 'exec' },
      { name: 'false_exec', kind: 'exec', type: 'exec' }
    ]
  },
  {
    type: 'flow_control',
    operation: 'for_loop',
    label: 'For 循环',
    description: '从 start 到 end 循环执行',
    category: '控制流',
    icon: 'FOR',
    color: '#9C27B0',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'start', kind: 'data', type: 'float', value: 0 },
      { name: 'end', kind: 'data', type: 'float', value: 10 },
      { name: 'step', kind: 'data', type: 'float', value: 1 }
    ],
    outputPins: [
      { name: 'loop_body', kind: 'exec', type: 'exec' },
      { name: 'completed', kind: 'exec', type: 'exec' },
      { name: 'index', kind: 'data', type: 'float' },
      { name: 'count', kind: 'data', type: 'float' }
    ]
  },
  {
    type: 'flow_control',
    operation: 'while_loop',
    label: 'While 循环',
    description: '当条件为真时循环执行',
    category: '控制流',
    icon: 'WHILE',
    color: '#9C27B0',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'condition', kind: 'data', type: 'bool', value: true }
    ],
    outputPins: [
      { name: 'loop_body', kind: 'exec', type: 'exec' },
      { name: 'completed', kind: 'exec', type: 'exec' },
      { name: 'iterations', kind: 'data', type: 'float' }
    ]
  },
  {
    type: 'flow_control',
    operation: 'sequence',
    label: '序列',
    description: '按顺序执行多个分支',
    category: '控制流',
    icon: '→',
    color: '#9C27B0',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' }
    ],
    outputPins: [
      { name: 'then_0', kind: 'exec', type: 'exec' },
      { name: 'then_1', kind: 'exec', type: 'exec' },
      { name: 'then_2', kind: 'exec', type: 'exec' }
    ]
  },

  // ==================== 纯数据节点（无执行引脚） ====================
  {
    type: 'data',
    operation: 'constant',
    label: '常量',
    description: '输出一个常量值（纯数据节点）',
    category: '数据',
    icon: '123',
    color: '#607D8B',
    inputPins: [
      { name: 'value', kind: 'data', type: 'any', value: 0 }
    ],
    outputPins: [
      { name: 'output', kind: 'data', type: 'any' }
    ]
  },
  {
    type: 'data',
    operation: 'get_variable',
    label: '获取变量',
    description: '获取变量值（纯数据节点）',
    category: '数据',
    icon: 'GET',
    color: '#607D8B',
    inputPins: [
      { name: 'name', kind: 'data', type: 'string', value: '' }
    ],
    outputPins: [
      { name: 'value', kind: 'data', type: 'any' }
    ]
  },
  {
    type: 'data',
    operation: 'make_float',
    label: '浮点数',
    description: '创建浮点数（纯数据节点）',
    category: '数据',
    icon: '1.0',
    color: '#4CAF50',
    inputPins: [
      { name: 'value', kind: 'data', type: 'float', value: 0 }
    ],
    outputPins: [
      { name: 'output', kind: 'data', type: 'float' }
    ]
  },
  {
    type: 'data',
    operation: 'make_bool',
    label: '布尔值',
    description: '创建布尔值（纯数据节点）',
    category: '数据',
    icon: 'T/F',
    color: '#F44336',
    inputPins: [
      { name: 'value', kind: 'data', type: 'bool', value: false }
    ],
    outputPins: [
      { name: 'output', kind: 'data', type: 'bool' }
    ]
  },
  {
    type: 'data',
    operation: 'make_string',
    label: '字符串',
    description: '创建字符串（纯数据节点）',
    category: '数据',
    icon: 'ABC',
    color: '#E91E63',
    inputPins: [
      { name: 'value', kind: 'data', type: 'string', value: '' }
    ],
    outputPins: [
      { name: 'output', kind: 'data', type: 'string' }
    ]
  },

  // ==================== 变量操作节点（可执行） ====================
  {
    type: 'variable',
    operation: 'set_variable',
    label: '设置变量',
    description: '设置变量值（可执行节点）',
    category: '变量',
    icon: 'SET',
    color: '#795548',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'name', kind: 'data', type: 'string', value: '' },
      { name: 'value', kind: 'data', type: 'any' }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'value', kind: 'data', type: 'any' }
    ]
  },

  // ==================== 特殊节点 ====================
  {
    type: 'start',
    label: '开始',
    description: '蓝图开始节点',
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
    description: '蓝图结束节点',
    category: '特殊',
    icon: '■',
    color: '#F44336',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'result', kind: 'data', type: 'any' }
    ],
    outputPins: []
  },

  // ==================== 输出/调试节点（可执行） ====================
  {
    type: 'debug',
    operation: 'print',
    label: '打印',
    description: '打印值到控制台',
    category: '调试',
    icon: '📝',
    color: '#00BCD4',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'value', kind: 'data', type: 'any' },
      { name: 'label', kind: 'data', type: 'string', value: '' }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' }
    ]
  },

  // ==================== 延时/异步节点（可执行） ====================
  {
    type: 'async',
    operation: 'delay',
    label: '延时',
    description: '延时指定毫秒后继续执行',
    category: '异步',
    icon: '⏱',
    color: '#00BCD4',
    inputPins: [
      { name: 'exec_in', kind: 'exec', type: 'exec' },
      { name: 'duration', kind: 'data', type: 'float', value: 1000 }
    ],
    outputPins: [
      { name: 'exec_out', kind: 'exec', type: 'exec' },
      { name: 'duration', kind: 'data', type: 'float' }
    ]
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
