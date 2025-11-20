<template>
  <el-card class="book-card glass-card" @click="redirectToDetail(book.id)">
    <div class="card-content">
      <div class="cover-container">
        <img 
          class="book-cover" 
          :src="proxy_image ? ('/api/proxy/cover/' + book.cover) : book.cover" 
          alt="book cover"
          loading="lazy"
        />
      </div>
      <div class="info-container">
        <div class="info-item title">{{ book.title }}</div>
        <div class="info-item author" v-if="book.authors && book.authors.length">
          <el-icon class="icon-author"><User /></el-icon>
          {{ truncateText(book.authors.join(', '), 25) }}
        </div>
        <div class="info-item publisher" v-if="more_info && book.publisher">
          <el-icon class="icon-publisher"><OfficeBuilding /></el-icon>
          {{ truncateText(book.publisher, 20) }}
        </div>
        <div class="info-item isbn" v-if="more_info && book.isbn">
          <el-icon class="icon-isbn"><Postcard /></el-icon>
          {{ book.isbn }}
        </div>
        <div class="info-item pubdate" v-if="book.pubdate">
          <el-icon class="icon-date"><Calendar /></el-icon>
          {{ new Date(book.pubdate).toLocaleDateString() }}
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ElCard, ElIcon } from 'element-plus'
import { User, Calendar, OfficeBuilding, Postcard } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router';
import { Book } from '@/types/book'

const props = defineProps({
  book: {
    type: Object as () => Book,
    required: true
  },
  more_info: {
    type: Boolean,
    default: false
  },
  proxy_image: {
    type: Boolean,
    default: false
  }
});

const router = useRouter();

const redirectToDetail = (id: number) => {
  router.push(`/detail/${id}`);
};

const truncateText = (text: string, maxLength: number = 20) => {
  if (!text) return '';
  return text.length > maxLength ? text.substring(0, maxLength - 3) + '...' : text;
};
</script>

<style scoped lang="scss">
.book-card {
  width: 100%;
  min-height: 160px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--glass-border);
  overflow: hidden;
  
  &:hover {
    transform: translateY(-4px) scale(1.02);
    box-shadow: 0 12px 24px rgba(0, 0, 0, 0.15);
    border-color: rgba(255, 255, 255, 0.3);
    
    .book-cover {
      transform: scale(1.05);
    }
    
    .title {
      color: var(--el-color-primary);
    }
  }
  
  :deep(.el-card__body) {
    padding: 0;
    height: 100%;
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

.book-cover {
  width: 100%;
  max-width: 80px;
  height: auto;
  aspect-ratio: 96 / 139;
  border-radius: var(--border-radius-sm);
  object-fit: cover;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
  transition: transform 0.3s ease;
}

.info-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  min-width: 0; /* 防止文本溢出 */
}

.info-item {
  line-height: 1.5;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  display: flex;
  align-items: center;
  gap: 6px;
  
  .el-icon {
    flex-shrink: 0;
    font-size: 14px;
    opacity: 0.7;
  }
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: normal;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;  
  overflow: hidden;
  margin-bottom: 4px;
  transition: color 0.2s ease;
}

.author,
.publisher,
.isbn,
.pubdate {
  font-size: 13px;
  color: var(--el-text-color-regular);
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
  
  .book-cover {
    max-width: 70px;
  }
  
  .title {
    font-size: 14px;
  }
  
  .author,
  .publisher,
  .isbn,
  .pubdate {
    font-size: 12px;
  }
  
  .info-item .el-icon {
    font-size: 12px;
  }
}
</style>
