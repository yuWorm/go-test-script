import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Blueprint,
  Node,
  Connection,
  ExecutionResult,
  NodeDebugInfo,
  DebugSession
} from '../types/blueprint'
import { executeBlueprint } from '../api/blueprint'

export const useBlueprintStore = defineStore('blueprint', () => {
  // 状态
  const blueprint = ref<Blueprint>({
    name: '新建蓝图',
    version: '1.0.0',
    description: '',
    nodes: [],
    connections: [],
    variables: {},
    metadata: {}
  })

  const selectedNodeId = ref<string | null>(null)
  const isExecuting = ref(false)
  const executionResult = ref<ExecutionResult | null>(null)
  const debugSession = ref<DebugSession | null>(null)

  // 计算属性
  const selectedNode = computed(() => {
    if (!selectedNodeId.value) return null
    return blueprint.value.nodes.find(n => n.id === selectedNodeId.value)
  })

  const nodeCount = computed(() => blueprint.value.nodes.length)
  const connectionCount = computed(() => blueprint.value.connections.length)

  // 方法：节点操作
  function addNode(node: Node) {
    blueprint.value.nodes.push(node)
  }

  function updateNode(nodeId: string, updates: Partial<Node>) {
    const index = blueprint.value.nodes.findIndex(n => n.id === nodeId)
    if (index !== -1 && blueprint.value.nodes[index]) {
      Object.assign(blueprint.value.nodes[index], updates)
    }
  }

  function deleteNode(nodeId: string) {
    // 删除节点
    blueprint.value.nodes = blueprint.value.nodes.filter(n => n.id !== nodeId)
    // 删除相关连接
    blueprint.value.connections = blueprint.value.connections.filter(
      c => c.source_node !== nodeId && c.target_node !== nodeId
    )
  }

  function getNode(nodeId: string): Node | undefined {
    return blueprint.value.nodes.find(n => n.id === nodeId)
  }

  // 方法：连接操作
  function addConnection(connection: Connection) {
    // 检查是否已存在
    const exists = blueprint.value.connections.some(
      c => c.source_node === connection.source_node &&
           c.source_pin === connection.source_pin &&
           c.target_node === connection.target_node &&
           c.target_pin === connection.target_pin
    )
    if (!exists) {
      blueprint.value.connections.push(connection)
    }
  }

  function deleteConnection(connectionId: string) {
    blueprint.value.connections = blueprint.value.connections.filter(
      c => c.id !== connectionId
    )
  }

  function getConnections(nodeId: string): Connection[] {
    return blueprint.value.connections.filter(
      c => c.source_node === nodeId || c.target_node === nodeId
    )
  }

  // 方法：选择
  function selectNode(nodeId: string | null) {
    selectedNodeId.value = nodeId
  }

  // 方法：蓝图操作
  function clearBlueprint() {
    blueprint.value = {
      name: '新建蓝图',
      version: '1.0.0',
      description: '',
      nodes: [],
      connections: [],
      variables: {},
      metadata: {}
    }
    selectedNodeId.value = null
    executionResult.value = null
    debugSession.value = null
  }

  function loadBlueprint(data: Blueprint) {
    blueprint.value = data
    selectedNodeId.value = null
    executionResult.value = null
    debugSession.value = null
  }

  function exportBlueprint(): Blueprint {
    return JSON.parse(JSON.stringify(blueprint.value))
  }

  // 方法：执行
  async function execute() {
    try {
      isExecuting.value = true
      executionResult.value = null

      // 创建调试会话
      debugSession.value = {
        blueprintName: blueprint.value.name,
        startTime: Date.now(),
        nodes: new Map(),
        status: 'running'
      }

      // 初始化所有节点的调试信息
      blueprint.value.nodes.forEach(node => {
        debugSession.value!.nodes.set(node.id, {
          nodeId: node.id,
          status: 'idle'
        })
      })

      // 调用后端执行
      const result = await executeBlueprint(blueprint.value)
      executionResult.value = result

      // 更新调试会话
      if (debugSession.value && result.nodes) {
        debugSession.value.endTime = Date.now()
        debugSession.value.status = result.success ? 'completed' : 'failed'

        // 更新每个节点的状态（使用后端返回的节点信息）
        for (const nodeId in result.nodes) {
          const nodeInfo = result.nodes[nodeId]
          const nodeDebug = debugSession.value.nodes.get(nodeId)

          if (nodeDebug) {
            nodeDebug.status = nodeInfo.status
            nodeDebug.error = nodeInfo.error
            nodeDebug.outputs = nodeInfo.outputs
            nodeDebug.duration = nodeInfo.duration
          }
        }
      }

      return result
    } catch (error: any) {
      console.error('Execute error:', error)
      if (debugSession.value) {
        debugSession.value.status = 'failed'
        debugSession.value.endTime = Date.now()
      }
      throw error
    } finally {
      isExecuting.value = false
    }
  }

  // 方法：调试
  function getNodeDebugInfo(nodeId: string): NodeDebugInfo | undefined {
    return debugSession.value?.nodes.get(nodeId)
  }

  function clearDebugSession() {
    debugSession.value = null
    executionResult.value = null
  }

  return {
    // 状态
    blueprint,
    selectedNodeId,
    selectedNode,
    isExecuting,
    executionResult,
    debugSession,

    // 计算属性
    nodeCount,
    connectionCount,

    // 方法
    addNode,
    updateNode,
    deleteNode,
    getNode,
    addConnection,
    deleteConnection,
    getConnections,
    selectNode,
    clearBlueprint,
    loadBlueprint,
    exportBlueprint,
    execute,
    getNodeDebugInfo,
    clearDebugSession
  }
})
