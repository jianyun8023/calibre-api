<template>
  <div class="app-layout">
    <!-- 顶部导航 -->
    <header class="site-header">
      <SiteHeader />
    </header>

    <div class="main-container">
      <!-- 桌面端侧边栏 -->
      <aside class="sidebar-container hidden-sm-and-down">
        <Sidebar />
      </aside>

      <!-- 主要内容区域 -->
      <main class="content-container">
        <div class="scrollable-content">
          <router-view v-slot="{ Component }" :key="$route.fullPath">
            <transition name="fade" mode="out-in">
              <component :is="Component" :key="$route.fullPath" />
            </transition>
          </router-view>
          
          <footer class="site-footer">
            <SiteFooter />
          </footer>
        </div>
      </main>
    </div>

    <!-- 移动端底部导航 -->
    <div class="bottom-nav-container hidden-md-and-up">
      <BottomNav />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useTheme } from './composables/useTheme'
import SiteHeader from './components/Header.vue'
import Sidebar from './components/Sidebar.vue'
import SiteFooter from './components/Footer.vue'
import BottomNav from './components/BottomNav.vue'

// 初始化主题系统
const { theme } = useTheme()

onMounted(() => {
  console.log('Current theme:', theme.value)
})
</script>

<style scoped lang="scss">
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--bg-gradient);
  background-size: 400% 400%;
  /* 性能优化: 移除持续动画以减少 CPU 占用 */
  /* animation: gradient-shift 15s ease infinite; */
  transition: background 0.3s ease;
}

/* 深色模式下的背景动画调整 */
:global([data-theme="dark"]) .app-layout {
  /* 性能优化: 移除持续动画 */
  /* animation: gradient-shift 20s ease infinite; */
}

.site-header {
  flex-shrink: 0;
  z-index: 100;
}

.main-container {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

.sidebar-container {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.02);
}

.content-container {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止 flex 子项溢出 */
}

.scrollable-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--spacing-lg);
  
  /* 自定义滚动条 */
  &::-webkit-scrollbar {
    width: 8px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  
  &::-webkit-scrollbar-thumb {
    background: rgba(156, 163, 175, 0.3);
    border-radius: 4px;
    
    &:hover {
      background: rgba(156, 163, 175, 0.5);
    }
  }
}

.site-footer {
  margin-top: auto;
  padding-top: var(--spacing-xl);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .scrollable-content {
    padding: var(--spacing-md);
    padding-bottom: 80px; /* 为底部导航留出空间 */
  }
  
  .site-footer {
    padding-bottom: 20px;
  }
}

/* 路由过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

