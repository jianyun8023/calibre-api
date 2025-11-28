<template>
  <div class="header-wrapper glass-panel">
    <el-row class="header-content" justify="space-between" align="middle">
      <!-- Logo and Title -->
      <el-col :span="8" :xs="16">
        <el-link href="/" class="logo-section" :underline="false">
          <div class="logo-icon">
            <el-icon :size="32"><Reading /></el-icon>
          </div>
          <div class="title-group">
            <el-text class="site-title">书海拾贝</el-text>
            <el-text class="site-subtitle">Your Personal Library</el-text>
          </div>
        </el-link>
      </el-col>
      
      <!-- Search Hint (Desktop only) - Click to search -->
      <el-col :span="8" class="hidden-sm-and-down">
        <div class="search-hint" @click="goToSearch">
          <el-icon><Search /></el-icon>
          <span>快速搜索书籍...</span>
          <kbd>Ctrl+K</kbd>
        </div>
      </el-col>
      
      <!-- Theme Toggle -->
      <el-col :span="8" :xs="8" class="text-right">
        <div class="theme-toggle-wrapper">
          <el-button
            circle
            class="theme-toggle-button"
            @click="toggleTheme"
            :icon="isDark ? Sunny : Moon"
            size="large"
          />
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTheme } from '../composables/useTheme'
import { Sunny, Moon, Reading, Search } from '@element-plus/icons-vue'

const { toggleTheme, isDark } = useTheme()
const router = useRouter()

// 跳转到搜索页
const goToSearch = () => {
  router.push('/search')
}

// Ctrl+K 快捷键
const handleKeydown = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    goToSearch()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped lang="scss">
.header-wrapper {
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-lg) var(--spacing-xl);
}

.header-content {
  width: 100%;
}

.text-right {
  display: flex;
  justify-content: flex-end;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  text-decoration: none;
  transition: var(--transition-normal);
  
  &:hover {
    transform: translateY(-2px);
    
    .logo-icon {
      transform: rotate(10deg) scale(1.1);
    }
  }
}

.logo-icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-gradient);
  border-radius: var(--border-radius-md);
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transition: var(--transition-normal);
}

.title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.site-title {
  font-size: 1.75rem;
  font-weight: 700;
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-family: Inter, 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  letter-spacing: 0.5px;
  line-height: 1.2;
}

.site-subtitle {
  font-size: 0.75rem;
  color: var(--el-text-color-regular);
  font-weight: 500;
  letter-spacing: 1px;
  text-transform: uppercase;
}

.search-hint {
  flex: 1;
  max-width: 400px;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  border-radius: var(--border-radius-md);
  color: var(--el-text-color-regular);
  font-size: 0.9rem;
  cursor: pointer;
  transition: var(--transition-normal);
  
  &:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: var(--el-color-primary);
  }
  
  kbd {
    padding: 2px 8px;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--glass-border);
    border-radius: 4px;
    font-size: 0.75rem;
    font-family: monospace;
    margin-left: auto;
  }
}

.theme-toggle-wrapper {
  display: flex;
  align-items: center;
}

.theme-toggle-button {
  background: var(--glass-bg-medium);
  border: 1px solid var(--glass-border);
  color: var(--el-text-color-primary);
  transition: all var(--transition-normal);
  
  &:hover {
    background: var(--el-color-primary);
    color: white;
    transform: translateY(-2px) rotate(15deg);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }
}

// 移动端适配
@media (max-width: 768px) {
  .header-wrapper {
    padding: var(--spacing-md) var(--spacing-lg);
  }
  
  .logo-icon {
    width: 48px;
    height: 48px;
    
    .el-icon {
      font-size: 24px;
    }
  }
  
  .site-title {
    font-size: 1.5rem;
  }
  
  .site-subtitle {
    font-size: 0.65rem;
  }
}
</style>
