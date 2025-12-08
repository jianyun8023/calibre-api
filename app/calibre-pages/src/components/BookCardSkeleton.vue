<template>
  <el-card class="book-card glass-card skeleton-card">
    <div class="card-content">
      <div class="cover-container">
        <el-skeleton :rows="0" animated>
          <template #template>
            <el-skeleton-item variant="rect" class="cover-skeleton" />
          </template>
        </el-skeleton>
      </div>
      <div class="info-container">
        <el-skeleton :rows="3" animated>
          <template #template>
            <el-skeleton-item variant="text" class="title-skeleton" />
            <el-skeleton-item variant="text" class="meta-skeleton" />
            <el-skeleton-item variant="text" class="meta-skeleton short" />
          </template>
        </el-skeleton>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ElCard, ElSkeleton, ElSkeletonItem } from 'element-plus'
</script>

<style scoped lang="scss">
.book-card {
  width: 100%;
  min-height: 160px;
  border: 1px solid var(--glass-border);
  overflow: hidden;
  background: var(--glass-bg-strong);
  border-radius: var(--border-radius-md);
  
  // 降级策略
  @supports (backdrop-filter: blur(12px)) {
    background: var(--glass-bg);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  
  :deep(.el-card__body) {
    padding: 0;
    height: 100%;
  }
  
  :deep(.el-skeleton) {
    --el-skeleton-color: var(--surface3);
    --el-skeleton-to-color: var(--surface4);
  }
}

.card-content {
  display: flex;
  align-items: stretch;
  height: 100%;
  min-height: 160px;
  padding: 12px;
  gap: 16px;
}

.cover-container {
  flex-shrink: 0;
  width: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-skeleton {
  width: 100%;
  height: 0;
  padding-bottom: 150%; // 2/3 aspect ratio
  border-radius: var(--border-radius-sm);
  background: var(--surface3);
}

.info-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
}

.title-skeleton {
  height: 20px;
  width: 90%;
  margin-bottom: 8px;
}

.meta-skeleton {
  height: 14px;
  width: 70%;
  
  &.short {
    width: 50%;
  }
}

/* 移动端优化 */
@media (max-width: 768px) {
  .card-content {
    min-height: 140px;
    padding: 10px;
    gap: 12px;
  }
  
  .cover-container {
    width: 70px;
  }
  
  .title-skeleton {
    height: 18px;
  }
  
  .meta-skeleton {
    height: 12px;
  }
}
</style>

