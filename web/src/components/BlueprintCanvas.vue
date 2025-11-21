<template>
  <div class="blueprint-canvas">
    <VueFlow
      v-model:nodes="nodes"
      v-model:edges="edges"
      :default-viewport="{ zoom: 1 }"
      :min-zoom="0.2"
      :max-zoom="4"
      @nodes-change="onNodesChange"
      @edges-change="onEdgesChange"
      @connect="onConnect"
      @node-click="onNodeClick"
    >
      <Background pattern-color="#aaa" :gap="16" />
      <Controls />
      <MiniMap />

      <template #node-custom="{ data }">
        <CustomNode :data="data" />
      </template>
    </VueFlow>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import type { Node, Edge, Connection } from '@vue-flow/core'
import { useBlueprintStore } from '../stores/blueprint'
import CustomNode from './CustomNode.vue'

const blueprintStore = useBlueprintStore()
const { fitView } = useVueFlow()

// Vue Flow 节点和边
const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])

// 将蓝图节点转换为 Vue Flow 节点
function convertToVueFlowNodes(): Node[] {
  return blueprintStore.blueprint.nodes.map(node => ({
    id: node.id,
    type: 'custom',
    position: { x: node.position.x, y: node.position.y },
    data: {
      ...node,
      debugInfo: blueprintStore.getNodeDebugInfo(node.id)
    }
  }))
}

// 将蓝图连接转换为 Vue Flow 边
function convertToVueFlowEdges(): Edge[] {
  return blueprintStore.blueprint.connections.map(conn => ({
    id: conn.id,
    source: conn.source_node,
    sourceHandle: conn.source_pin,
    target: conn.target_node,
    targetHandle: conn.target_pin,
    animated: blueprintStore.isExecuting,
    style: { stroke: '#555' }
  }))
}

// 监听蓝图变化 - 使用计数器触发更新，避免 deep watch 性能问题
watch(
  () => blueprintStore.updateCounter,
  () => {
    nodes.value = convertToVueFlowNodes()
    edges.value = convertToVueFlowEdges()
  },
  { immediate: true }
)

// 节点变化处理
function onNodesChange(changes: any[]) {
  changes.forEach(change => {
    if (change.type === 'position' && change.position) {
      blueprintStore.updateNode(change.id, {
        position: change.position
      })
    }
  })
}

// 边变化处理
function onEdgesChange(changes: any[]) {
  changes.forEach(change => {
    if (change.type === 'remove') {
      blueprintStore.deleteConnection(change.id)
    }
  })
}

// 连接处理
function onConnect(connection: Connection) {
  if (connection.source && connection.target) {
    const conn: any = {
      id: `${connection.source}-${connection.sourceHandle}-${connection.target}-${connection.targetHandle}`,
      source_node: connection.source,
      source_pin: connection.sourceHandle || '',
      target_node: connection.target,
      target_pin: connection.targetHandle || ''
    }
    blueprintStore.addConnection(conn)
  }
}

// 节点点击
function onNodeClick(event: { node: Node }) {
  blueprintStore.selectNode(event.node.id)
}

// 暴露方法
defineExpose({
  fitView
})
</script>

<style scoped>
.blueprint-canvas {
  width: 100%;
  height: 100%;
  background: #1a1a1a;
}

:deep(.vue-flow__node) {
  border-radius: 8px;
  font-size: 12px;
}

:deep(.vue-flow__handle) {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
</style>
