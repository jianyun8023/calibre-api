<template>
  <div class="books-page">
    <el-row>
      <SearchBar />
    </el-row>
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
            class="book-col"
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
import SearchBar from '@/components/SearchBar.vue'
import BookCard from '@/components/BookCard.vue'
import { Book } from '@/types/book'
import { fetchRecentBooks } from "@/api/api";

export default {
  name: 'Books',
  components: {
    BookCard,
    SearchBar,
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

.section-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  position: relative;
  
  &::after {
    content: '';
    position: absolute;
    left: 0;
    bottom: -8px;
    width: 60px;
    height: 3px;
    background: var(--primary-gradient);
    border-radius: 2px;
  }
}

.book-count {
  font-size: 0.875rem;
  color: var(--text-secondary);
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
  color: var(--text-primary);
  font-size: 0.875rem;
  padding: 0 var(--spacing-md);
  font-weight: 500;
}

.glass-button {
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-sm) var(--spacing-lg);
  color: var(--text-primary);
  transition: all 0.3s ease;
  
  &:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.18);
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  }
  
  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    transform: none;
  }
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
