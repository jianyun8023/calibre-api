<template>
  <el-container class="app-layout">
    <!-- 顶部导航 -->
    <el-header height="auto" class="site-header">
      <SiteHeader />
    </el-header>

    <el-container class="main-container">
      <!-- 桌面端侧边栏 -->
      <el-aside width="240px" class="sidebar-container hidden-sm-and-down">
        <Sidebar />
      </el-aside>

      <!-- 主要内容区域 -->
      <el-main class="content-container">
        <div class="scrollable-content">
          <router-view v-slot="{ Component }" :key="$route.fullPath">
            <transition name="fade" mode="out-in">
              <component :is="Component" :key="$route.fullPath" />
            </transition>
          </router-view>
          
          <el-footer height="auto" class="site-footer">
            <SiteFooter />
          </el-footer>
        </div>
      </el-main>
    </el-container>

    <!-- 移动端底部导航 -->
    <div class="bottom-nav-container hidden-md-and-up">
      <BottomNav />
    </div>
  </el-container>
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
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--bg-gradient);
  background-size: 400% 400%;
  transition: background 0.3s ease;
}

.site-header {
  padding: 0;
  z-index: 100;
}

.main-container {
  overflow: hidden;
  position: relative;
}

.sidebar-container {
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.02);
  overflow-y: auto;
  overflow-x: hidden;
  
  &::-webkit-scrollbar {
    width: 0;
    background: transparent;
  }
}

.content-container {
  padding: 0;
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.scrollable-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--spacing-md);
  
  @media (max-width: 768px) {
    padding-bottom: calc(var(--spacing-md) + 68px); // BottomNav height
  }
  
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
  padding: var(--spacing-xl) 0 0;
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

