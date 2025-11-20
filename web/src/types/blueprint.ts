/**
 * 蓝图类型定义
 * 与 Go 后端的 JSON 格式保持一致
 */

export type NodeType = 'start' | 'end' | 'arithmetic' | 'logic' | 'condition' | 'function'

export interface Position {
  x: number
  y: number
}

export type PinKind = 'exec' | 'data' | 'error'

export interface Pin {
  name: string
  kind?: PinKind // 引脚类型：执行、数据或错误
  type: string
  value?: any
}

export interface Node {
  id: string
  type: NodeType
  operation?: string
  label: string
  input_pins: Pin[]
  output_pins: Pin[]
  properties?: Record<string, any>
  position: Position
}

export interface Connection {
  id: string
  source_node: string
  source_pin: string
  target_node: string
  target_pin: string
}

export interface Blueprint {
  name: string
  version: string
  description?: string
  nodes: Node[]
  connections: Connection[]
  variables?: Record<string, any>
  metadata?: Record<string, any>
}

/**
 * 节点执行信息
 */
export interface NodeExecutionInfo {
  node_id: string
  status: 'idle' | 'success' | 'error' | 'skipped'
  error?: string
  outputs: Record<string, any>
  duration: number
}

/**
 * 执行结果
 */
export interface ExecutionResult {
  success: boolean
  errors?: string[]
  duration: number // 纳秒
  variables: Record<string, any>
  outputs: Record<string, any> // nodeID.pinName -> value
  nodes: Record<string, NodeExecutionInfo> // 每个节点的执行信息
}

/**
 * 节点模板（用于调色板）
 */
export interface NodeTemplate {
  type: NodeType
  operation?: string
  label: string
  description: string
  category: string
  icon?: string
  inputPins: Pin[]
  outputPins: Pin[]
  color?: string
}

/**
 * 执行选项
 */
export interface ExecutionOptions {
  mode: 'sequential' | 'parallel'
  timeout?: number
  maxConcurrency?: number
  stopOnError: boolean
}

/**
 * 调试信息
 */
export interface NodeDebugInfo {
  nodeId: string
  status: 'idle' | 'running' | 'success' | 'error'
  inputs?: Record<string, any>
  outputs?: Record<string, any>
  error?: string
  duration?: number
}

export interface DebugSession {
  blueprintName: string
  startTime: number
  endTime?: number
  nodes: Map<string, NodeDebugInfo>
  status: 'idle' | 'running' | 'completed' | 'failed'
}
