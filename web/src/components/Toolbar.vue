<template>
  <div class="toolbar">
    <div class="toolbar-section">
      <h2 class="blueprint-title" contenteditable @blur="updateBlueprintName">
        {{ blueprintStore.blueprint.name }}
      </h2>
      <span class="node-count">{{ blueprintStore.nodeCount }} 节点 / {{ blueprintStore.connectionCount }} 连接</span>
    </div>

    <div class="toolbar-section toolbar-actions">
      <button class="btn btn-primary" @click="handleExecute" :disabled="blueprintStore.isExecuting">
        <span class="btn-icon">{{ blueprintStore.isExecuting ? '⏳' : '▶' }}</span>
        {{ blueprintStore.isExecuting ? '执行中...' : '执行' }}
      </button>

      <button class="btn" @click="handleClear">
        <span class="btn-icon">🗑</span>
        清空
      </button>

      <button class="btn" @click="handleExport">
        <span class="btn-icon">💾</span>
        导出
      </button>

      <button class="btn" @click="handleImport">
        <span class="btn-icon">📁</span>
        导入
      </button>

      <input
        ref="fileInput"
        type="file"
        accept=".json"
        style="display: none"
        @change="onFileChange"
      />
    </div>

    <!-- 执行结果提示 -->
    <div v-if="showResult" class="result-toast" :class="resultClass">
      {{ resultMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useBlueprintStore } from '../stores/blueprint'

const blueprintStore = useBlueprintStore()
const fileInput = ref<HTMLInputElement>()
const showResult = ref(false)
const resultMessage = ref('')
const resultClass = ref('')

// 更新蓝图名称
function updateBlueprintName(event: FocusEvent) {
  const target = event.target as HTMLElement
  const newName = target.textContent?.trim() || '未命名蓝图'
  blueprintStore.blueprint.name = newName
}

// 执行蓝图
async function handleExecute() {
  try {
    showResult.value = false
    const result = await blueprintStore.execute()

    resultMessage.value = result.success
      ? `执行成功！耗时: ${(result.duration / 1000000).toFixed(2)}ms`
      : `执行失败: ${result.errors?.join(', ')}`

    resultClass.value = result.success ? 'success' : 'error'
    showResult.value = true

    setTimeout(() => {
      showResult.value = false
    }, 5000)
  } catch (error: any) {
    resultMessage.value = `执行错误: ${error.message}`
    resultClass.value = 'error'
    showResult.value = true

    setTimeout(() => {
      showResult.value = false
    }, 5000)
  }
}

// 清空蓝图
function handleClear() {
  if (confirm('确定要清空整个蓝图吗？此操作不可撤销。')) {
    blueprintStore.clearBlueprint()
  }
}

// 导出蓝图
function handleExport() {
  const blueprint = blueprintStore.exportBlueprint()
  const json = JSON.stringify(blueprint, null, 2)
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${blueprint.name}.json`
  a.click()
  URL.revokeObjectURL(url)

  resultMessage.value = '蓝图已导出'
  resultClass.value = 'success'
  showResult.value = true
  setTimeout(() => {
    showResult.value = false
  }, 3000)
}

// 导入蓝图
function handleImport() {
  fileInput.value?.click()
}

// 文件选择
function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const json = e.target?.result as string
      const blueprint = JSON.parse(json)
      blueprintStore.loadBlueprint(blueprint)

      resultMessage.value = '蓝图导入成功'
      resultClass.value = 'success'
      showResult.value = true
      setTimeout(() => {
        showResult.value = false
      }, 3000)
    } catch (error: any) {
      resultMessage.value = `导入失败: ${error.message}`
      resultClass.value = 'error'
      showResult.value = true
      setTimeout(() => {
        showResult.value = false
      }, 3000)
    }
  }
  reader.readAsText(file)

  // 重置输入以允许重复导入同一文件
  target.value = ''
}
</script>

<style scoped>
.toolbar {
  height: 60px;
  background: #2a2a2a;
  border-bottom: 1px solid #444;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  position: relative;
}

.toolbar-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.blueprint-title {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  outline: none;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.blueprint-title:hover {
  background: rgba(255, 255, 255, 0.05);
}

.blueprint-title:focus {
  background: rgba(255, 255, 255, 0.1);
}

.node-count {
  color: #888;
  font-size: 12px;
}

.toolbar-actions {
  gap: 8px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #3a3a3a;
  border: 1px solid #555;
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:hover:not(:disabled) {
  background: #444;
  border-color: #666;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #4CAF50;
  border-color: #4CAF50;
}

.btn-primary:hover:not(:disabled) {
  background: #45a049;
  border-color: #45a049;
}

.btn-icon {
  font-size: 14px;
}

.result-toast {
  position: absolute;
  top: 70px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 20px;
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
  z-index: 1000;
  animation: slideDown 0.3s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.result-toast.success {
  background: #4CAF50;
}

.result-toast.error {
  background: #F44336;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}
</style>
