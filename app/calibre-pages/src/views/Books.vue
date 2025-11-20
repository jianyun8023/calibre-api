<template>
  <div class="books-page">
    <!-- 移除 SearchBar，统一在Header中搜索 -->
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

<script lang="ts">
import { ElButton, ElCard, ElCol, ElContainer, ElInput, ElRow, ElSkeleton, ElEmpty } from 'element-plus'
import { ArrowLeftBold, ArrowRightBold } from '@element-plus/icons-vue'
// SearchBar 已移除，仅在搜索页保留
import BookCard from '@/components/BookCard.vue'
import { Book } from '@/types/book'
import { fetchRecentBooks } from "@/api/api";

export default {
  name: 'Books',
  components: {
    BookCard,
    // SearchBar 已移除
    ElContainer,
    ElRow,
    ElCol,
    ElInput,
    ElButton,
    ElCard,
    ElSkeleton,
    ElEmpty,
    ArrowLeftBold,
    ArrowRightBold
  },
  data() {
    return {
      searchQuery: '',
      recentBooks: [] as Book[],
      limit: 12 as number,
      offset: 0 as number,
      total: 0,
      loading: false
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.total / this.limit)
    },
    currentPage() {
      return Math.floor(this.offset / this.limit) + 1
    }
  },
  created() {
    this.initializeFromQueryParams()
    this.fetchBooks()
  },
  // 当组件被激活时重新获取数据
  activated() {
    console.log('Books page activated, refreshing data...')
    this.initializeFromQueryParams()
    this.fetchBooks()
  },
  watch: {
    offset() {
      this.updateQueryParams()
      this.fetchBooks()
    },
    limit() {
      this.updateQueryParams()
      this.fetchBooks()
    }
  },
  methods: {
    async fetchBooks() {
      this.loading = true
      try {
        const data = await fetchRecentBooks(this.limit, this.offset)
        this.recentBooks = data.records
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    prevPage() {
      if (this.offset > 0) {
        this.offset -= this.limit
        window.scrollTo({ top: 0, behavior: 'smooth' })
      }
    },
    nextPage() {
      if (this.offset + this.limit < this.total) {
        this.offset += this.limit
        window.scrollTo({ top: 0, behavior: 'smooth' })
      }
    },
    updateQueryParams() {
      this.$router.push({ query: { ...this.$route.query, offset: this.offset, limit: this.limit } })
    },
    initializeFromQueryParams() {
      const query = this.$route.query
      if (query.offset) {
        this.offset = parseInt(query.offset as string, 10)
      }
      if (query.limit) {
        this.limit = parseInt(query.limit as string, 10)
      }
    }
  }
}
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

/* .section-title 样式已移至 index.scss */

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

.book-col {
  margin-bottom: var(--spacing-lg);
  transition: all 0.3s ease;
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

/* .glass-button 样式已移至 index.scss */

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
