<template>
  <el-menu 
    class="sidebar-menu glass-panel" 
    default-active="1" 
    @select="handleSelect" 
    :router="true"
  >
    <div class="menu-header">
      <el-text class="menu-title">导航</el-text>
    </div>
    
    <el-menu-item index="/" class="menu-item-enhanced">
      <el-icon><House/></el-icon>
      <span>首页</span>
      <el-tag size="small" class="menu-badge" effect="plain">Hot</el-tag>
    </el-menu-item>
    
    <el-menu-item index="/books" class="menu-item-enhanced">
      <el-icon><Files/></el-icon>
      <span>书库</span>
    </el-menu-item>
    
    <el-menu-item index="/search" class="menu-item-enhanced">
      <el-icon><Search/></el-icon>
      <span>搜索</span>
    </el-menu-item>
    
    <el-menu-item index="/publisher" class="menu-item-enhanced">
      <el-icon><Shop/></el-icon>
      <span>出版社</span>
    </el-menu-item>
    
    <el-divider class="menu-divider" />
    
    <el-menu-item index="/setting" class="menu-item-enhanced">
      <el-icon><Setting/></el-icon>
      <span>设置</span>
    </el-menu-item>
  </el-menu>
</template>

<script setup lang="ts">
import { House, Files, Search, Setting, Shop } from '@element-plus/icons-vue'

const handleSelect = (key: string, keyPath: string) => {
  console.log(key, keyPath);
};
</script>

<style scoped lang="scss">
.sidebar-menu {
  width: 100%;
  height: 100%;
  background: transparent;
  border: none;
  padding: var(--spacing-lg) 0;
  
  .menu-header {
    padding: 0 var(--spacing-lg) var(--spacing-md);
    
    .menu-title {
      font-size: 0.75rem;
      font-weight: 600;
      color: var(--el-text-color-regular);
      text-transform: uppercase;
      letter-spacing: 1px;
    }
  }
  
  .menu-divider {
    margin: var(--spacing-md) var(--spacing-lg);
    border-color: var(--glass-border);
  }
  
  :deep(.el-menu-item) {
    color: var(--el-text-color-primary);
    transition: var(--transition-normal);
    margin: var(--spacing-xs) var(--spacing-md);
    border-radius: var(--border-radius-md);
    height: 52px;
    line-height: 52px;
    position: relative;
    overflow: hidden;
    
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      width: 3px;
      height: 0;
      background: var(--el-color-primary);
      border-radius: 0 2px 2px 0;
      transition: var(--transition-normal);
    }
    
    &:hover {
      background: rgba(255, 255, 255, 0.08);
      transform: translateX(4px);
      color: var(--el-color-primary);
      
      &::before {
        height: 60%;
      }
      
      .el-icon {
        transform: scale(1.1);
      }
    }
    
    &.is-active {
      background: rgba(59, 130, 246, 0.15);
      color: var(--el-color-primary);
      font-weight: 600;
      
      &::before {
        height: 80%;
      }
      
      .el-icon {
        color: var(--el-color-primary);
      }
    }
    
    .el-icon {
      font-size: 20px;
      transition: var(--transition-normal);
    }
    
    span {
      margin-left: var(--spacing-sm);
    }
  }
  
  .menu-badge {
    margin-left: auto;
    background: rgba(59, 130, 246, 0.2);
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
    font-size: 0.65rem;
    padding: 0 6px;
    height: 18px;
    line-height: 18px;
  }
}
</style>
