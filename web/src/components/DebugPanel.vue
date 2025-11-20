<template>
  <div class="debug-panel" :class="{ collapsed: isCollapsed }">
    <div class="panel-header" @click="toggleCollapse">
      <span class="panel-icon">{{ isCollapsed ? '◀' : '▶' }}</span>
      <h3>调试面板</h3>
      <span v-if="debugSession" class="session-status" :class="`status-${debugSession.status}`">
        {{ statusText }}
      </span>
    </div>

    <div v-if="!isCollapsed" class="panel-content">
      <div v-if="!debugSession" class="empty-state">
        <p>暂无调试信息</p>
        <p class="hint">执行蓝图后将显示详细的执行信息</p>
      </div>

      <div v-else class="debug-info">
        <!-- 执行摘要 -->
        <div class="section">
          <h4>执行摘要</h4>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">蓝图:</span>
              <span class="info-value">{{ debugSession.blueprintName }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">状态:</span>
              <span class="info-value" :class="`status-${debugSession.status}`">
                {{ statusText }}
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">耗时:</span>
              <span class="info-value">{{ formatDuration() }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">节点数:</span>
              <span class="info-value">{{ debugSession.nodes.size }}</span>
            </div>
          </div>
        </div>

        <!-- 节点详情 -->
        <div class="section">
          <h4>节点执行详情</h4>
          <div class="node-list">
            <div
              v-for="[nodeId, nodeDebug] in debugSession.nodes"
              :key="nodeId"
              class="node-debug-item"
              :class="`status-${nodeDebug.status}`"
              @click="selectNode(nodeId)"
            >
              <div class="node-debug-header">
                <span class="status-dot"></span>
                <span class="node-id">{{ getNodeLabel(nodeId) }}</span>
                <span class="node-status">{{ getStatusText(nodeDebug.status) }}</span>
              </div>

              <!-- 输入 -->
              <div v-if="nodeDebug.inputs" class="node-debug-section">
                <div class="section-title">输入:</div>
                <div class="pin-values">
                  <div v-for="(value, key) in nodeDebug.inputs" :key="key" class="pin-value">
                    <span class="pin-name">{{ key }}:</span>
                    <span class="pin-val">{{ formatValue(value) }}</span>
                  </div>
                </div>
              </div>

              <!-- 输出 -->
              <div v-if="nodeDebug.outputs" class="node-debug-section">
                <div class="section-title">输出:</div>
                <div class="pin-values">
                  <div v-for="(value, key) in nodeDebug.outputs" :key="key" class="pin-value">
                    <span class="pin-name">{{ key }}:</span>
                    <span class="pin-val success">{{ formatValue(value) }}</span>
                  </div>
                </div>
              </div>

              <!-- 错误 -->
              <div v-if="nodeDebug.error" class="node-debug-section error">
                <div class="section-title">错误:</div>
                <div class="error-message">{{ nodeDebug.error }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 执行结果 -->
        <div v-if="executionResult" class="section">
          <h4>执行结果</h4>
          <div v-if="executionResult.errors && executionResult.errors.length > 0" class="error-list">
            <div v-for="(error, index) in executionResult.errors" :key="index" class="error-item">
              {{ error }}
            </div>
          </div>
          <div v-else class="success-message">
            ✓ 执行成功
          </div>
        </div>

        <!-- 清除按钮 -->
        <button class="btn-clear" @click="clearDebug">
          清除调试信息
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useBlueprintStore } from '../stores/blueprint'

const blueprintStore = useBlueprintStore()
const isCollapsed = ref(false)

const debugSession = computed(() => blueprintStore.debugSession)
const executionResult = computed(() => blueprintStore.executionResult)

const statusText = computed(() => {
  if (!debugSession.value) return ''
  switch (debugSession.value.status) {
    case 'idle': return '空闲'
    case 'running': return '执行中'
    case 'completed': return '已完成'
    case 'failed': return '失败'
    default: return ''
  }
})

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function formatDuration() {
  if (!debugSession.value) return '-'
  const start = debugSession.value.startTime
  const end = debugSession.value.endTime || Date.now()
  return `${(end - start).toFixed(0)}ms`
}

function getNodeLabel(nodeId: string): string {
  const node = blueprintStore.getNode(nodeId)
  return node ? `${node.label} (${nodeId})` : nodeId
}

function getStatusText(status: string): string {
  switch (status) {
    case 'idle': return '待执行'
    case 'running': return '执行中'
    case 'success': return '成功'
    case 'error': return '错误'
    default: return status
  }
}

function formatValue(value: any): string {
  if (typeof value === 'number') {
    return value.toFixed(2)
  }
  if (typeof value === 'boolean') {
    return value ? 'true' : 'false'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

function selectNode(nodeId: string) {
  blueprintStore.selectNode(nodeId)
}

function clearDebug() {
  blueprintStore.clearDebugSession()
}
</script>

<style scoped>
.debug-panel {
  width: 350px;
  height: 100%;
  background: #2a2a2a;
  border-left: 1px solid #444;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width 0.3s;
}

.debug-panel.collapsed {
  width: 40px;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid #444;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.panel-header:hover {
  background: rgba(255, 255, 255, 0.05);
}

.panel-icon {
  font-size: 12px;
  color: #888;
}

.panel-header h3 {
  flex: 1;
  margin: 0;
  color: #fff;
  font-size: 15px;
}

.session-status {
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}

.session-status.status-running {
  background: rgba(255, 193, 7, 0.2);
  color: #FFC107;
}

.session-status.status-completed {
  background: rgba(76, 175, 80, 0.2);
  color: #4CAF50;
}

.session-status.status-failed {
  background: rgba(244, 67, 54, 0.2);
  color: #F44336;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.empty-state {
  text-align: center;
  color: #888;
  padding: 40px 20px;
}

.empty-state p {
  margin: 8px 0;
}

.empty-state .hint {
  font-size: 12px;
  color: #666;
}

.debug-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section h4 {
  margin: 0 0 12px 0;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  font-size: 12px;
}

.info-label {
  color: #888;
}

.info-value {
  color: #fff;
  font-weight: 500;
}

.node-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-debug-item {
  background: #1a1a1a;
  border-left: 3px solid #666;
  border-radius: 4px;
  padding: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.node-debug-item:hover {
  background: #252525;
}

.node-debug-item.status-success {
  border-left-color: #4CAF50;
}

.node-debug-item.status-error {
  border-left-color: #F44336;
}

.node-debug-item.status-running {
  border-left-color: #FFC107;
}

.node-debug-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #666;
}

.status-success .status-dot {
  background: #4CAF50;
}

.status-error .status-dot {
  background: #F44336;
}

.status-running .status-dot {
  background: #FFC107;
  animation: pulse 1s infinite;
}

.node-id {
  flex: 1;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
}

.node-status {
  color: #888;
  font-size: 11px;
}

.node-debug-section {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #333;
}

.section-title {
  color: #888;
  font-size: 11px;
  margin-bottom: 6px;
}

.pin-values {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pin-value {
  display: flex;
  gap: 8px;
  font-size: 11px;
}

.pin-name {
  color: #888;
}

.pin-val {
  color: #fff;
  font-weight: 500;
}

.pin-val.success {
  color: #4CAF50;
}

.error-message {
  color: #F44336;
  font-size: 11px;
  line-height: 1.4;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.error-item {
  padding: 8px;
  background: rgba(244, 67, 54, 0.1);
  border-left: 2px solid #F44336;
  border-radius: 3px;
  color: #F44336;
  font-size: 11px;
}

.success-message {
  padding: 8px;
  background: rgba(76, 175, 80, 0.1);
  border-left: 2px solid #4CAF50;
  border-radius: 3px;
  color: #4CAF50;
  font-size: 12px;
}

.btn-clear {
  width: 100%;
  padding: 8px;
  background: #3a3a3a;
  border: 1px solid #555;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover {
  background: #444;
}

/* 滚动条样式 */
.panel-content::-webkit-scrollbar {
  width: 6px;
}

.panel-content::-webkit-scrollbar-track {
  background: #1a1a1a;
}

.panel-content::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 3px;
}

.panel-content::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
