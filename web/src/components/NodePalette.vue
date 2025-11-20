<template>
  <div class="node-palette">
    <div class="palette-header">
      <h3>节点库</h3>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="搜索节点..."
        class="search-input"
      />
    </div>

    <div class="palette-content">
      <div
        v-for="category in filteredCategories"
        :key="category"
        class="category"
      >
        <div class="category-header" @click="toggleCategory(category)">
          <span class="category-icon">{{ expandedCategories.has(category) ? '▼' : '▶' }}</span>
          <span class="category-name">{{ category }}</span>
        </div>

        <div v-if="expandedCategories.has(category)" class="category-nodes">
          <div
            v-for="template in getNodesByCategory(category)"
            :key="`${template.type}-${template.operation || ''}`"
            class="node-item"
            :style="{ borderLeft: `3px solid ${template.color}` }"
            draggable="true"
            @dragstart="onDragStart($event, template)"
            @click="addNodeToCanvas(template)"
          >
            <div class="node-item-header">
              <span class="node-item-icon">{{ template.icon }}</span>
              <span class="node-item-label">{{ template.label }}</span>
            </div>
            <div class="node-item-description">{{ template.description }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { nodeTemplates, getCategories, getNodesByCategory } from '../config/nodeTemplates'
import type { NodeTemplate } from '../types/blueprint'
import { useBlueprintStore } from '../stores/blueprint'

const blueprintStore = useBlueprintStore()

const searchQuery = ref('')
const expandedCategories = ref(new Set(['算术运算', '比较运算', '逻辑运算', '控制流', '函数', '特殊']))

// 过滤后的分类
const filteredCategories = computed(() => {
  if (!searchQuery.value) {
    return getCategories()
  }

  const query = searchQuery.value.toLowerCase()
  const matchingTemplates = nodeTemplates.filter(
    t => t.label.toLowerCase().includes(query) ||
         t.description.toLowerCase().includes(query)
  )

  return Array.from(new Set(matchingTemplates.map(t => t.category)))
})

// 切换分类展开状态
function toggleCategory(category: string) {
  if (expandedCategories.value.has(category)) {
    expandedCategories.value.delete(category)
  } else {
    expandedCategories.value.add(category)
  }
}

// 拖拽开始
function onDragStart(event: DragEvent, template: NodeTemplate) {
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'copy'
    event.dataTransfer.setData('application/vueflow', JSON.stringify(template))
  }
}

// 添加节点到画布
function addNodeToCanvas(template: NodeTemplate) {
  const id = `node_${Date.now()}`
  const node = {
    id,
    type: template.type,
    operation: template.operation,
    label: template.label,
    input_pins: JSON.parse(JSON.stringify(template.inputPins)),
    output_pins: JSON.parse(JSON.stringify(template.outputPins)),
    position: {
      x: Math.random() * 400 + 100,
      y: Math.random() * 400 + 100
    }
  }
  blueprintStore.addNode(node)
}
</script>

<style scoped>
.node-palette {
  width: 280px;
  height: 100%;
  background: #2a2a2a;
  border-right: 1px solid #444;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.palette-header {
  padding: 16px;
  border-bottom: 1px solid #444;
}

.palette-header h3 {
  margin: 0 0 12px 0;
  color: #fff;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
}

.search-input::placeholder {
  color: #666;
}

.palette-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.category {
  margin-bottom: 8px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #333;
  border-radius: 4px;
  cursor: pointer;
  color: #fff;
  font-weight: 500;
  font-size: 13px;
  transition: background 0.2s;
}

.category-header:hover {
  background: #3a3a3a;
}

.category-icon {
  font-size: 10px;
}

.category-nodes {
  padding: 4px 0;
}

.node-item {
  padding: 10px 12px;
  margin: 4px 0;
  background: #1a1a1a;
  border-radius: 4px;
  cursor: grab;
  transition: all 0.2s;
}

.node-item:hover {
  background: #252525;
  transform: translateX(4px);
}

.node-item:active {
  cursor: grabbing;
}

.node-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.node-item-icon {
  font-size: 14px;
}

.node-item-label {
  color: #fff;
  font-weight: 500;
  font-size: 12px;
}

.node-item-description {
  color: #888;
  font-size: 11px;
  line-height: 1.4;
}

/* 滚动条样式 */
.palette-content::-webkit-scrollbar {
  width: 6px;
}

.palette-content::-webkit-scrollbar-track {
  background: #1a1a1a;
}

.palette-content::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 3px;
}

.palette-content::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
