<template>
  <div class="search-wrapper glass-container">
    <el-affix target=".search-wrapper">
      <el-input
          v-model="searchQuery"
          @keyup.enter="redirectToSearch"
          placeholder="书名、作者、ISBN"
          class="search-input"
          size="large"
      >
        <template #append>
          <el-button size="large" @click="redirectToSearch">搜索</el-button>
        </template>
      </el-input>
    </el-affix>
  </div>
</template>

<script setup lang="ts">
import {ElButton, ElInput} from 'element-plus'
import {ref} from 'vue'
import {useRouter} from 'vue-router'

const searchQuery = ref('')
const router = useRouter()

const redirectToSearch = () => {
  if (searchQuery.value) {
    router.push({path: '/search', query: {q: searchQuery.value}})
  }
}
</script>
<style scoped lang="scss">
.search-wrapper {
  margin-bottom: var(--spacing-lg);
}

.search-input {
  :deep(.el-input__wrapper) {
    background: rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(8px);
    border: 1px solid var(--glass-border);
    transition: var(--transition-normal);
    color: var(--el-text-color-primary);
    
    &:hover {
      background: rgba(255, 255, 255, 0.2);
    }
    
    &.is-focus {
      background: rgba(255, 255, 255, 0.22);
      border-color: rgba(255, 255, 255, 0.4);
    }
  }
  
  :deep(.el-input__inner) {
    color: var(--el-text-color-primary);
    
    &::placeholder {
      color: var(--el-text-color-placeholder);
    }
  }
  
  :deep(.el-input-group__append) {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(8px);
    border: 1px solid var(--glass-border);
    border-left: none;
    
    .el-button {
      background: transparent;
      border: none;
      color: var(--el-text-color-primary);
      font-weight: 500;
      transition: var(--transition-normal);
      
      &:hover {
        background: rgba(255, 255, 255, 0.1);
        transform: scale(1.05);
      }
    }
  }
}
</style>
