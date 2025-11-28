<template>
  <div class="bottom-nav glass-panel">
    <router-link to="/" class="nav-item" active-class="active">
      <div class="nav-icon-wrapper">
        <el-icon><House /></el-icon>
        <span class="nav-dot"></span>
      </div>
      <span class="nav-label">首页</span>
    </router-link>
    <router-link to="/books" class="nav-item" active-class="active">
      <div class="nav-icon-wrapper">
        <el-icon><Files /></el-icon>
        <span class="nav-dot"></span>
      </div>
      <span class="nav-label">书库</span>
    </router-link>
    <router-link to="/search" class="nav-item" active-class="active">
      <div class="nav-icon-wrapper">
        <el-icon><Search /></el-icon>
        <span class="nav-dot"></span>
      </div>
      <span class="nav-label">搜索</span>
    </router-link>
    <router-link to="/setting" class="nav-item" active-class="active">
      <div class="nav-icon-wrapper">
        <el-icon><Setting /></el-icon>
        <span class="nav-dot"></span>
      </div>
      <span class="nav-label">设置</span>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { House, Files, Search, Setting } from '@element-plus/icons-vue'
</script>

<style scoped lang="scss">
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 68px;
  display: flex;
  justify-content: space-around;
  align-items: center;
  z-index: 1000;
  padding-bottom: env(safe-area-inset-bottom);
  border-top: 1px solid var(--glass-border);
  background: var(--glass-bg-strong);
  /* 性能优化：移除 backdrop-filter */
  /* backdrop-filter: blur(16px); */
  /* -webkit-backdrop-filter: blur(16px); */
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.08);
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  color: var(--el-text-color-regular);
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--border-radius-md);
  transition: var(--transition-normal);
  position: relative;
  min-width: 64px;
  
  &:active {
    transform: scale(0.95);
  }
}

.nav-icon-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--border-radius-md);
  transition: var(--transition-normal);
  
  .el-icon {
    font-size: 24px;
    transition: var(--transition-normal);
  }
  
  .nav-dot {
    position: absolute;
    top: -2px;
    right: -2px;
    width: 6px;
    height: 6px;
    background: var(--el-color-primary);
    border-radius: 50%;
    opacity: 0;
    transform: scale(0);
    transition: var(--transition-normal);
  }
}

.nav-label {
  font-size: 0.7rem;
  margin-top: 4px;
  font-weight: 500;
  transition: var(--transition-normal);
}

.nav-item.active {
  color: var(--el-color-primary);
  
  .nav-icon-wrapper {
    background: rgba(59, 130, 246, 0.15);
    
    .el-icon {
      transform: translateY(-2px);
    }
    
    .nav-dot {
      opacity: 1;
      transform: scale(1);
    }
  }
  
  .nav-label {
    font-weight: 600;
  }
}

// 移动端显示
@media (max-width: 768px) {
  .bottom-nav {
    display: flex;
  }
}

@media (min-width: 769px) {
  .bottom-nav {
    display: none;
  }
}
</style>
