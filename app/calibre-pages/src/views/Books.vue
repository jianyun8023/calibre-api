<template>
  <div class="books-page">
    <el-container class="books-wrapper">
      <section class="glass-container">
        <el-row class="mb-4">
          <el-col :span="24">
            <div class="section-header">
              <h2 class="section-title">全部书籍</h2>
              <span class="book-count">共 {{ total }} 本书籍</span>
            </div>
          </el-col>
        </el-row>
        
        <!-- Loading State -->
        <el-row v-if="loading" :gutter="20" class="book-grid">
          <el-col v-for="n in 12" :key="n" :span="6" :lg="6" :md="12" :sm="24" :xs="24">
            <el-skeleton :rows="4" animated />
          </el-col>
        </el-row>
        
        <!-- Book Grid -->
        <el-row v-else-if="recentBooks.length > 0" :gutter="20" class="book-grid">
          <el-col 
            v-for="book in recentBooks" 
            :key="book.id" 
            :span="6" 
            :lg="6" 
            :md="12" 
            :sm="24" 
            :xs="24"
            class="book-col-spacing"
          >
            <BookCard :book="book" :more_info="true" />
          </el-col>
        </el-row>
        
        <!-- Empty State -->
        <el-empty v-else description="暂无书籍" />
        
        <!-- Pagination -->
        <el-row class="pagination-row" justify="center">
          <el-button 
            class="glass-button" 
            @click="prevPage" 
            :disabled="offset === 0 || loading"
          >
            <el-icon><ArrowLeftBold /></el-icon>
            上一页
          </el-button>
          <span class="page-info">
            第 {{ currentPage }} / {{ totalPages }} 页
          </span>
          <el-button 
            class="glass-button" 
            @click="nextPage" 
            :disabled="offset + limit >= total || loading"
          >
            下一页
            <el-icon><ArrowRightBold /></el-icon>
          </el-button>
        </el-row>
      </section>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onActivated, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeftBold, ArrowRightBold } from '@element-plus/icons-vue'
import BookCard from '@/components/BookCard.vue'
import { Book } from '@/types/book'
import { fetchRecentBooks } from "@/api/api"

const router = useRouter()
const route = useRoute()

// State
const recentBooks = ref<Book[]>([])
const limit = ref(12)
const offset = ref(0)
const total = ref(0)
const loading = ref(false)

// Computed
const totalPages = computed(() => Math.ceil(total.value / limit.value))
const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1)

// Methods
const fetchBooks = async () => {
  loading.value = true
  try {
    const data = await fetchRecentBooks(limit.value, offset.value)
    recentBooks.value = data.records
    total.value = data.total
  } finally {
    loading.value = false
  }
}

const prevPage = () => {
  if (offset.value > 0) {
    offset.value -= limit.value
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

const nextPage = () => {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

const updateQueryParams = () => {
  router.push({ 
    query: { 
      ...route.query, 
      offset: offset.value, 
      limit: limit.value 
    } 
  })
}

const initializeFromQueryParams = () => {
  const query = route.query
  if (query.offset) {
    offset.value = parseInt(query.offset as string, 10)
  }
  if (query.limit) {
    limit.value = parseInt(query.limit as string, 10)
  }
}

// Watchers
watch(offset, () => {
  updateQueryParams()
  fetchBooks()
})

watch(limit, () => {
  updateQueryParams()
  fetchBooks()
})

// Lifecycle
onMounted(() => {
  initializeFromQueryParams()
  fetchBooks()
})

onActivated(() => {
  console.log('Books page activated, refreshing data...')
  initializeFromQueryParams()
  fetchBooks()
})
</script>

<style scoped lang="scss">
.books-page {
  animation: fadeIn 0.5s ease-out;
}

.books-wrapper {
  margin-top: var(--spacing-lg);
  width: 100%;
  max-width: 1400px;
  margin-left: auto;
  margin-right: auto;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.book-count {
  font-size: 0.875rem;
  color: var(--el-text-color-regular);
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(4px);
  border-radius: var(--border-radius-sm);
  border: 1px solid var(--glass-border);
}

.book-grid {
  min-height: 400px;
}

.pagination-row {
  margin-top: var(--spacing-xl);
  gap: var(--spacing-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-info {
  color: var(--el-text-color-primary);
  font-size: 0.875rem;
  padding: 0 var(--spacing-md);
  font-weight: 500;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .section-title {
    font-size: 1.5rem;
  }
  
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .pagination-row {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .page-info {
    order: -1;
  }
}
</style>
