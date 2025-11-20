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
      { name: 'a', type: 'float' },
      { name: 'b', type: 'float' }
    ],
    outputPins: [
      { name: 'result', type: 'float' }
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

  // 特殊节点
  {
    type: 'start',
    label: '开始',
    description: '蓝图开始节点',
    category: '特殊',
    icon: '▶',
    color: '#8BC34A',
    inputPins: [],
    outputPins: []
  },
  {
    type: 'end',
    label: '结束',
    description: '蓝图结束节点',
    category: '特殊',
    icon: '■',
    color: '#F44336',
    inputPins: [],
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
