<template>
  <div class="read-book-wrapper glass-container">
    <!-- 加载骨架屏 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="15" animated />
      <div class="loading-text">正在加载电子书...</div>
    </div>
    
    <!-- 阅读器 -->
    <div v-else class="reader-container">
      <vue-reader 
        ref="reader" 
        :url="bookUrl"
        :location="location"
        :getRendition="getRendition"
        @update:location="locationChange"
        @book:ready="onBookReady"
      />
      <el-button 
        v-if="initPath" 
        @click="jumpToPath(initPath)" 
        class="glass-button jump-btn"
      >
        <el-icon class="mr-1"><Position /></el-icon>
        跳转到指定位置
      </el-button>
    </div>
  </div>
</template>
<script lang="ts">
import {VueReader} from 'vue-reader'
import {useStorage} from '@vueuse/core'

import {ElContainer, ElRow} from "element-plus";

export default {
  name: 'ReadBook',
  components: {ElContainer, ElRow, VueReader},
  data() {
    return {
      loading: true, // 加载状态
      bookId: '',
      bookUrl: '',
      initPath: '',
      location: useStorage('book-progress', 0, undefined, {
        serializer: {
          read: (v) => JSON.parse(v),
          write: (v) => JSON.stringify(v),
        },
      }),
      rendition: {}

    }
  },
  created() {
    // this.$refs.reader.getRendition
    this.bookId = (this.$route as any).params.id
    this.bookUrl = `/api/download/book/${this.bookId}.epub`
    if (this.$route.query.path){
      this.initPath = this.$route.query.path as string
    }
  },
  methods: {
    onBookReady() {
      this.loading = false
    },
    locationChange: (epubcifi) => {
      console.log(epubcifi)
      location.value = epubcifi
    },
    jumpToPath(path: any) {
      console.log(this.rendition)
      this.$refs.reader.setLocation(path);
    },
    getRendition(val) {
      this.rendition = val
      this.rendition.themes.default({
        '::selection': {
          background: 'orange',
        },
      })
      // this.renditionrendition.on('selected', setRenderSelection)
    }
  }
}
</script>
<style scoped lang="scss">
.read-book-wrapper {
  min-height: 90vh;
  padding: var(--spacing-lg);
}

.loading-container {
  padding: var(--spacing-xl);
  text-align: center;
  min-height: 70vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  
  :deep(.el-skeleton) {
    --el-skeleton-color: rgba(255, 255, 255, 0.1);
    --el-skeleton-to-color: rgba(255, 255, 255, 0.15);
    max-width: 800px;
    margin: 0 auto;
  }
  
  .loading-text {
    margin-top: var(--spacing-xl);
    color: var(--text2);
    font-size: 1.1rem;
    font-weight: 500;
  }
}

.reader-container {
  position: relative;
  height: 85vh;
  border-radius: var(--border-radius-md);
  overflow: hidden;
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(4px);
  border: 1px solid var(--glass-border);
}

.jump-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 100;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  color: var(--text1);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-sm);
  box-shadow: var(--glass-shadow);
  transition: var(--transition-normal);
  
  &:hover {
    background: rgba(255, 255, 255, 0.18);
    transform: translateY(-2px);
    box-shadow: var(--glass-shadow-hover);
  }
  
  .mr-1 {
    margin-right: 4px;
  }
}

// 深色模式调整
:root[data-theme="dark"] {
  .reader-container {
    background: rgba(255, 255, 255, 0.03);
  }
  
  .jump-btn {
    background: rgba(255, 255, 255, 0.08);
    
    &:hover {
      background: rgba(255, 255, 255, 0.12);
    }
  }
}
</style>