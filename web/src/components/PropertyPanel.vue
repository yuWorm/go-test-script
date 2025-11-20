<template>
  <div class="property-panel" v-if="selectedNode">
    <div class="panel-header">
      <h3>节点属性</h3>
      <button class="btn-close" @click="close">×</button>
    </div>

    <div class="panel-content">
      <!-- 基本信息 -->
      <div class="section">
        <h4>基本信息</h4>
        <div class="property-row">
          <label>标签:</label>
          <input
            v-model="selectedNode.label"
            type="text"
            class="property-input"
            @change="updateNode"
          />
        </div>
        <div class="property-row">
          <label>类型:</label>
          <span class="property-value">{{ selectedNode.type }}</span>
        </div>
        <div class="property-row" v-if="selectedNode.operation">
          <label>操作:</label>
          <span class="property-value">{{ selectedNode.operation }}</span>
        </div>
      </div>

      <!-- 引脚编辑器（仅 Start/End 节点）-->
      <div v-if="canEditPins" class="section">
        <h4>{{ isStartNode ? '蓝图输入' : '蓝图输出' }}</h4>
        <p class="hint">{{ isStartNode ? '添加输出引脚作为蓝图的输入参数' : '添加输入引脚作为蓝图的输出结果' }}</p>

        <!-- 引脚列表 -->
        <div class="pin-list">
          <div
            v-for="(pin, index) in editablePins"
            :key="index"
            class="pin-edit-item"
          >
            <input
              v-model="pin.name"
              type="text"
              placeholder="引脚名称"
              class="pin-name-input"
              @change="updateNode"
            />
            <select
              v-model="pin.type"
              class="pin-type-select"
              @change="updateNode"
            >
              <option value="float">数字</option>
              <option value="string">字符串</option>
              <option value="bool">布尔</option>
              <option value="any">任意</option>
            </select>
            <input
              v-if="!isStartNode && pin.type === 'float'"
              v-model.number="pin.value"
              type="number"
              placeholder="默认值"
              class="pin-value-input"
              @change="updateNode"
            />
            <button class="btn-remove" @click="removePin(index)">×</button>
          </div>
        </div>

        <!-- 添加引脚按钮 -->
        <button class="btn-add" @click="addPin">
          + 添加{{ isStartNode ? '输出' : '输入' }}引脚
        </button>
      </div>

      <!-- 输入引脚默认值 -->
      <div v-if="selectedNode.input_pins && selectedNode.input_pins.length > 0" class="section">
        <h4>输入默认值</h4>
        <div
          v-for="pin in selectedNode.input_pins"
          :key="pin.name"
          class="property-row"
        >
          <label>{{ pin.name }}:</label>
          <input
            v-if="pin.type === 'float' || pin.type === 'int'"
            v-model.number="pin.value"
            type="number"
            class="property-input"
            @change="updateNode"
          />
          <input
            v-else-if="pin.type === 'string'"
            v-model="pin.value"
            type="text"
            class="property-input"
            @change="updateNode"
          />
          <input
            v-else-if="pin.type === 'bool'"
            v-model="pin.value"
            type="checkbox"
            class="property-checkbox"
            @change="updateNode"
          />
          <input
            v-else
            v-model="pin.value"
            type="text"
            class="property-input"
            @change="updateNode"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useBlueprintStore } from '../stores/blueprint'
import type { Pin } from '../types/blueprint'

const blueprintStore = useBlueprintStore()

const selectedNode = computed(() => blueprintStore.selectedNode)

const isStartNode = computed(() => selectedNode.value?.type === 'start')
const isEndNode = computed(() => selectedNode.value?.type === 'end')
const canEditPins = computed(() => isStartNode.value || isEndNode.value)

const editablePins = computed(() => {
  if (!selectedNode.value) return []
  return isStartNode.value
    ? selectedNode.value.output_pins
    : selectedNode.value.input_pins
})

function addPin() {
  if (!selectedNode.value) return

  const newPin: Pin = {
    name: `pin_${editablePins.value.length + 1}`,
    type: 'any',
    value: undefined
  }

  if (isStartNode.value) {
    selectedNode.value.output_pins.push(newPin)
  } else {
    selectedNode.value.input_pins.push(newPin)
  }

  updateNode()
}

function removePin(index: number) {
  if (!selectedNode.value) return

  if (isStartNode.value) {
    selectedNode.value.output_pins.splice(index, 1)
  } else {
    selectedNode.value.input_pins.splice(index, 1)
  }

  updateNode()
}

function updateNode() {
  if (!selectedNode.value) return
  blueprintStore.updateNode(selectedNode.value.id, {
    label: selectedNode.value.label,
    input_pins: selectedNode.value.input_pins,
    output_pins: selectedNode.value.output_pins
  })
}

function close() {
  blueprintStore.selectNode(null)
}
</script>

<style scoped>
.property-panel {
  width: 300px;
  height: 100%;
  background: #2a2a2a;
  border-left: 1px solid #444;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid #444;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-header h3 {
  margin: 0;
  color: #fff;
  font-size: 15px;
}

.btn-close {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid #555;
  border-radius: 4px;
  color: #fff;
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  line-height: 1;
}

.btn-close:hover {
  background: #444;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.section {
  margin-bottom: 24px;
}

.section h4 {
  margin: 0 0 12px 0;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}

.hint {
  margin: 0 0 12px 0;
  color: #888;
  font-size: 11px;
  line-height: 1.4;
}

.property-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.property-row label {
  color: #888;
  font-size: 12px;
  min-width: 80px;
}

.property-value {
  color: #fff;
  font-size: 12px;
  font-family: monospace;
}

.property-input {
  flex: 1;
  padding: 6px 10px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
}

.property-checkbox {
  width: 18px;
  height: 18px;
}

.pin-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.pin-edit-item {
  display: flex;
  gap: 6px;
  align-items: center;
}

.pin-name-input {
  flex: 1;
  padding: 6px 8px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 11px;
}

.pin-type-select {
  width: 80px;
  padding: 6px 8px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 11px;
}

.pin-value-input {
  width: 70px;
  padding: 6px 8px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 11px;
}

.btn-remove {
  width: 24px;
  height: 24px;
  background: #F44336;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  line-height: 1;
}

.btn-remove:hover {
  background: #D32F2F;
}

.btn-add {
  width: 100%;
  padding: 8px;
  background: #4CAF50;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-add:hover {
  background: #45a049;
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
