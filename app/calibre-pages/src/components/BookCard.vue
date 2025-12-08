<template>
  <div class="book-card-wrapper" @mousemove="handleMouseMove" @mouseleave="resetTilt" :style="cardStyle">
    <el-card class="book-card glass-card" @click="redirectToDetail(book.id)">
      <div class="card-content">
        <div class="cover-container">
          <el-image 
            class="book-cover" 
            :src="proxy_image ? ('/api/proxy/cover/' + book.cover) : book.cover" 
            alt="book cover"
            loading="lazy"
            fit="cover"
          >
            <template #placeholder>
              <div class="image-placeholder">
                <el-icon><Picture /></el-icon>
              </div>
            </template>
            <template #error>
              <div class="image-error">
                <el-icon><Warning /></el-icon>
              </div>
            </template>
          </el-image>
          <div class="cover-overlay"></div>
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
          <div class="info-item score" v-if="book.score !== undefined">
            <el-tag size="small" effect="dark" :color="getScoreColor(book.score)" style="border: none;">
              相似度: {{ (book.score * 100).toFixed(1) }}%
            </el-tag>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElCard, ElIcon, ElImage, ElTag } from 'element-plus'
import { User, Calendar, OfficeBuilding, Postcard, Picture, Warning } from '@element-plus/icons-vue'
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

// 3D Tilt Effect State
const tiltX = ref(0)
const tiltY = ref(0)

const cardStyle = computed(() => ({
  transform: `perspective(1000px) rotateX(${tiltX.value}deg) rotateY(${tiltY.value}deg)`,
  transition: tiltX.value === 0 && tiltY.value === 0 ? 'transform 0.5s ease' : 'none'
}))

const handleMouseMove = (e: MouseEvent) => {
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  
  // Calculate rotation (max 5 degrees)
  tiltY.value = ((x / rect.width) - 0.5) * 10
  tiltX.value = ((y / rect.height) - 0.5) * -10
}

const resetTilt = () => {
  tiltX.value = 0
  tiltY.value = 0
}

const redirectToDetail = (id: number) => {
  router.push(`/detail/${id}`);
};

const truncateText = (text: string, maxLength: number = 20) => {
  if (!text) return '';
  return text.length > maxLength ? text.substring(0, maxLength - 3) + '...' : text;
};

const getScoreColor = (score: number) => {
  // Return HSL color based on score
  return score > 0.7 
    ? `hsl(142, 71%, 45%)` // Green
    : `hsl(217, 91%, 60%)` // Blue
}
</script>

<style scoped lang="scss">
.book-card-wrapper {
  height: 100%;
  /* Ensure 3D context */
  transform-style: preserve-3d; 
  will-change: transform;
}

.book-card {
  width: 100%;
  min-height: 160px;
  cursor: pointer;
  border: 1px solid var(--glass-border);
  overflow: hidden;
  background: var(--glass-bg-strong); // Fallback
  border-radius: var(--border-radius-md);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  
  // 支持 backdrop-filter 的设备使用真 Glass
  @supports (backdrop-filter: blur(12px)) {
    background: var(--glass-bg);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  
  &:hover {
    box-shadow: var(--glass-shadow-hover);
    border-color: rgba(255, 255, 255, 0.3);
    background: var(--glass-bg-medium);
    
    @supports (backdrop-filter: blur(12px)) {
      background: var(--glass-bg-medium);
    }
    
    .book-cover {
      transform: scale(1.05) rotate(2deg);
    }
    
    .cover-overlay {
      opacity: 1;
    }
    
    .title {
      color: var(--accent-color);
    }
  }
  
  // 低性能设备降级
  @media (prefers-reduced-motion: reduce) {
    transition: none;
    
    &:hover {
      .book-cover {
        transform: none;
      }
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
  position: relative;
  z-index: 1;
}

.cover-container {
  flex-shrink: 0;
  width: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  border-radius: var(--border-radius-sm);
  overflow: hidden;
  background: var(--surface3);
}

.book-cover {
  width: 100%;
  height: auto;
  aspect-ratio: 2/3;
  object-fit: cover;
  transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  display: block;
}

.image-placeholder, .image-error {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background: var(--surface3);
  color: var(--text2);
  font-size: 24px;
}

.cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    to bottom,
    transparent 0%,
    rgba(0, 0, 0, 0.2) 100%
  );
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.info-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.info-item {
  line-height: 1.5;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text2);
  
  .el-icon {
    flex-shrink: 0;
    font-size: 14px;
    opacity: 0.7;
  }
}

.title {
  font-size: 16px; /* Spec: 16px */
  font-weight: 600;
  color: var(--text1);
  white-space: normal;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;  
  overflow: hidden;
  margin-bottom: 4px;
  transition: color 0.2s ease;
}

.author, .publisher, .isbn, .pubdate {
  font-size: 12px; /* Spec: 12px */
  font-weight: 400;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .book-card {
    min-height: auto;
  }

  .card-content {
    min-height: 140px;
    padding: 10px;
    gap: 12px;
  }
  
  .cover-container {
    width: 70px;
  }
  
  .title {
    font-size: 14px;
  }
}
</style>
