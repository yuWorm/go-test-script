<template>
  <div
    class="custom-node"
    :class="[`status-${debugInfo?.status || 'idle'}`, `type-${data.type}`]"
    :style="{ borderColor: nodeColor }"
  >
    <!-- 节点头部 -->
    <div class="node-header" :style="{ background: nodeColor }">
      <span class="node-icon">{{ nodeIcon }}</span>
      <span class="node-label">{{ data.label }}</span>
      <span v-if="debugInfo?.status === 'running'" class="status-indicator">⚡</span>
      <span v-else-if="debugInfo?.status === 'success'" class="status-indicator">✓</span>
      <span v-else-if="debugInfo?.status === 'error'" class="status-indicator">✗</span>
    </div>

    <!-- 输入引脚 -->
    <div v-if="data.input_pins && data.input_pins.length > 0" class="node-pins">
      <div
        v-for="pin in data.input_pins"
        :key="pin.name"
        class="pin-row input-pin"
        :class="{ 'exec-pin': isExecPin(pin) }"
      >
        <Handle
          :id="pin.name"
          type="target"
          :position="Position.Left"
          :class="['pin-handle', getPinHandleClass(pin)]"
        />
        <span class="pin-icon">{{ getPinIcon(pin) }}</span>
        <span class="pin-label">{{ pin.name }}</span>
        <span class="pin-type">{{ pin.type }}</span>
        <!-- 只显示值，不在节点上编辑（编辑在 PropertyPanel 中进行）-->
        <span v-if="pin.value !== undefined && !isExecPin(pin)" class="pin-default-value">
          {{ formatValue(pin.value) }}
        </span>
      </div>
    </div>

    <!-- 输出引脚 -->
    <div v-if="data.output_pins && data.output_pins.length > 0" class="node-pins">
      <div
        v-for="pin in data.output_pins"
        :key="pin.name"
        class="pin-row output-pin"
        :class="{ 'exec-pin': isExecPin(pin) }"
      >
        <span class="pin-icon">{{ getPinIcon(pin) }}</span>
        <span class="pin-label">{{ pin.name }}</span>
        <span class="pin-type">{{ pin.type }}</span>
        <span v-if="debugInfo?.outputs?.[pin.name] !== undefined && !isExecPin(pin)" class="pin-output-value">
          {{ formatValue(debugInfo.outputs[pin.name]) }}
        </span>
        <Handle
          :id="pin.name"
          type="source"
          :position="Position.Right"
          :class="['pin-handle', getPinHandleClass(pin)]"
        />
      </div>
    </div>

    <!-- 调试信息 -->
    <div v-if="debugInfo?.error" class="node-error">
      {{ debugInfo.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import type { Node as BPNode, NodeDebugInfo } from '../types/blueprint'
import { getNodeTemplate } from '../config/nodeTemplates'

interface Props {
  data: BPNode & { debugInfo?: NodeDebugInfo }
}

const props = defineProps<Props>()

// 获取节点模板
const template = computed(() => getNodeTemplate(props.data.type, props.data.operation))

// 节点颜色
const nodeColor = computed(() => template.value?.color || '#666')

// 节点图标
const nodeIcon = computed(() => template.value?.icon || '')

// 调试信息
const debugInfo = computed(() => props.data.debugInfo)

// 格式化值
function formatValue(value: any): string {
  if (typeof value === 'number') {
    return value.toFixed(2)
  }
  if (typeof value === 'boolean') {
    return value ? 'true' : 'false'
  }
  return String(value)
}

// 判断是否为执行引脚
function isExecPin(pin: any): boolean {
  return pin.kind === 'exec' || pin.type === 'exec'
}

// 判断是否为错误引脚
function isErrorPin(pin: any): boolean {
  return pin.kind === 'error' || pin.type === 'error'
}

// 获取引脚图标
function getPinIcon(pin: any): string {
  if (isExecPin(pin)) {
    return '▶' // 执行引脚：白色箭头
  }
  if (isErrorPin(pin)) {
    return '⚡' // 错误引脚：闪电
  }
  return '●' // 数据引脚：圆点
}

// 获取引脚样式类
function getPinHandleClass(pin: any): string {
  if (isExecPin(pin)) {
    return 'pin-handle-exec'
  }
  if (isErrorPin(pin)) {
    return 'pin-handle-error'
  }
  return `pin-handle-${pin.type || 'any'}`
}
</script>

<style scoped>
.custom-node {
  min-width: 180px;
  background: #2a2a2a;
  border: 2px solid #666;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  transition: all 0.3s ease;
}

.custom-node:hover {
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4);
  transform: translateY(-2px);
}

.custom-node.status-running {
  border-color: #FFC107 !important;
  animation: pulse 1s infinite;
}

.custom-node.status-success {
  border-color: #4CAF50 !important;
}

.custom-node.status-error {
  border-color: #F44336 !important;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.node-header {
  padding: 8px 12px;
  border-radius: 6px 6px 0 0;
  color: white;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-icon {
  font-size: 14px;
  min-width: 20px;
  text-align: center;
}

.node-label {
  flex: 1;
  font-size: 13px;
}

.status-indicator {
  font-size: 14px;
}

.node-pins {
  padding: 8px 0;
}

.pin-row {
  display: flex;
  align-items: center;
  padding: 4px 12px;
  gap: 8px;
  position: relative;
  color: #ccc;
  font-size: 11px;
}

.pin-row:hover {
  background: rgba(255, 255, 255, 0.05);
}

.pin-label {
  flex: 1;
  font-weight: 500;
}

.pin-type {
  color: #888;
  font-size: 10px;
}

.pin-value {
  width: 60px;
  padding: 2px 6px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 3px;
  color: #fff;
  font-size: 11px;
}

.pin-output-value {
  color: #4CAF50;
  font-weight: bold;
  font-size: 11px;
}

.pin-icon {
  font-size: 12px;
  margin-right: 4px;
  line-height: 1;
}

.exec-pin .pin-icon {
  color: #fff;
  font-weight: bold;
}

.pin-handle {
  width: 12px !important;
  height: 12px !important;
  background: #666;
  border: 2px solid #fff;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.pin-handle:hover {
  background: #4CAF50;
  transform: scale(1.2);
}

/* 执行引脚样式 - 白色箭头形状 */
.pin-handle-exec {
  background: #fff !important;
  clip-path: polygon(0% 50%, 40% 0%, 40% 35%, 100% 35%, 100% 65%, 40% 65%, 40% 100%);
  width: 14px !important;
  height: 14px !important;
  border: none !important;
}

.pin-handle-exec:hover {
  background: #FFC107 !important;
}

/* 错误引脚样式 - 红色闪电形状 */
.pin-handle-error {
  background: #F44336 !important;
  clip-path: polygon(50% 0%, 70% 40%, 100% 40%, 55% 100%, 45% 60%, 0% 60%);
  width: 14px !important;
  height: 14px !important;
  border: none !important;
}

.pin-handle-error:hover {
  background: #FF5722 !important;
}

/* 数据引脚颜色 */
.pin-handle-float,
.pin-handle-number {
  background: #4CAF50 !important;
}

.pin-handle-string {
  background: #E91E63 !important;
}

.pin-handle-bool,
.pin-handle-boolean {
  background: #F44336 !important;
}

.pin-handle-any {
  background: #9C27B0 !important;
}

.input-pin .pin-handle {
  left: -7px;
}

.output-pin .pin-handle {
  right: -7px;
}

/* 执行引脚行样式 */
.exec-pin {
  background: rgba(255, 255, 255, 0.05);
  border-left: 3px solid #fff;
  padding-left: 9px !important;
}

.exec-pin .pin-label {
  color: #fff;
  font-weight: bold;
}

.node-error {
  padding: 8px 12px;
  background: rgba(244, 67, 54, 0.1);
  border-top: 1px solid rgba(244, 67, 54, 0.3);
  color: #F44336;
  font-size: 10px;
  word-wrap: break-word;
}
</style>
