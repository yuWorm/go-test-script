import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'

// Vue Flow 样式
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.mount('#app')
