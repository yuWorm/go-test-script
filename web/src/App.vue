<template>
  <div class="app">
    <Toolbar />
    <div class="app-content">
      <NodePalette />
      <BlueprintCanvas ref="canvas" />
      <DebugPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Toolbar from './components/Toolbar.vue'
import NodePalette from './components/NodePalette.vue'
import BlueprintCanvas from './components/BlueprintCanvas.vue'
import DebugPanel from './components/DebugPanel.vue'

const canvas = ref<InstanceType<typeof BlueprintCanvas>>()

onMounted(() => {
  // 初始化后自适应视图
  setTimeout(() => {
    canvas.value?.fitView()
  }, 100)
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  width: 100%;
  height: 100%;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

.app {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1a1a1a;
}

.app-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}
</style>
