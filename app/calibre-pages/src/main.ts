import './assets/main.css'
import './styles/index.scss'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/display.css'

import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
const app = createApp(App)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.use(router)

app.mount('#app')
