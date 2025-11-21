<template>
  <div class="search-wrapper glass-container">
    <div class="search-input-wrapper">
      <el-affix target=".search-wrapper">
        <div class="search-bar">
          <el-input
              v-model="searchQuery"
              @input="handleInput"
              @keyup.enter="fetchBooks"
              type="text"
              :placeholder="searchMode === 'semantic' ? '描述您想找的书籍情节、主题...' : '书名、作者、ISBN'"
              size="large"
              class="search-input"
          >
            <template #append>
              <el-button @click="fetchBooks">
                <el-icon><SearchIcon /></el-icon>
              </el-button>
            </template>
          </el-input>
          
          <el-radio-group v-model="searchMode" @change="handleModeChange" class="mode-switch">
            <el-radio-button label="keyword">关键词搜索</el-radio-button>
            <el-radio-button label="semantic">语义搜索</el-radio-button>
          </el-radio-group>
        </div>
      </el-affix>
    </div>
    <h2 class="search-title">
      {{ searchMode === 'semantic' ? '语义匹配结果：' : '搜索结果：' }}
      <strong>{{ keyword }}</strong>
    </h2>
    <el-text class="search-count" v-if="searchMode === 'keyword'"
    >共计 {{ total }} 条, 当前{{ offset }} --
      {{ offset + limit >= total ? total : offset + limit }}
    </el-text>
    <el-text class="search-count" v-else>
      找到最相关的结果
    </el-text>

    <el-row :gutter="20" class="books-grid">
      <el-col v-for="book in books" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24" class="book-col-spacing">
        <BookCard :book="book" :more_info="true"/>
      </el-col>
    </el-row>
    <el-row class="pagination-row" justify="center" v-if="searchMode === 'keyword'">
      <el-button class="glass-button" @click="prevPage" :disabled="offset === 0">
        <el-icon>
          <ArrowLeftBold/>
        </el-icon>
        上一页
      </el-button>
      <el-button class="glass-button" @click="nextPage" :disabled="offset + limit >= total"
      >下一页
        <el-icon>
          <ArrowRightBold/>
        </el-icon>
      </el-button>
    </el-row>
  </div>
</template>

<script lang="ts">
import BookCard from '@/components/BookCard.vue'
import {ElButton, ElCol, ElInput, ElRow, ElRadioGroup, ElRadioButton, ElIcon} from 'element-plus'
import { Search as SearchIcon, ArrowLeftBold, ArrowRightBold } from '@element-plus/icons-vue'
import {fetchBooks, searchSemantic} from "@/api/api";
import type { Book } from '@/types/book';

export default {
  name: 'Search',
  components: {ElInput, ElButton, ElRow, ElCol, BookCard, ElRadioGroup, ElRadioButton, ElIcon, SearchIcon, ArrowLeftBold, ArrowRightBold},
  data() {
    return {
      searchQuery: '',
      keyword: '',
      publisher: '',
      author: '',
      books: [] as Book[],
      filter: [] as string[],
      limit: 12,
      offset: 0,
      total: 0,
      searchMode: 'keyword', // 'keyword' | 'semantic'
      debounceTimer: null as number | null
    }
  },
  created() {
    this.initializeFromQueryParams()
  },
  // 当组件被激活时重新获取数据
  activated() {
    console.log('Search page activated, refreshing data...')
    this.initializeFromQueryParams()
    if (this.searchQuery) {
      this.fetchBooks()
    }
  },
  watch: {
    publisher() {
      this.updateQueryParams()
      this.fetchBooks()
    },
    author() {
      this.updateQueryParams()
      this.fetchBooks()
    },
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
    handleInput() {
      if (this.debounceTimer) clearTimeout(this.debounceTimer)
      this.debounceTimer = setTimeout(() => {
        this.updateQueryParams()
        this.fetchBooks()
      }, 500) as unknown as number
    },
    
    handleModeChange() {
      this.offset = 0
      this.books = []
      this.updateQueryParams()
      if (this.searchQuery) {
        this.fetchBooks()
      }
    },

    async fetchBooks() {
      if (!this.searchQuery && !this.publisher && !this.author) {
        this.books = []
        return
      }

      this.keyword = this.searchQuery || this.publisher || this.author
      
      try {
        if (this.searchMode === 'semantic' && this.searchQuery) {
          const data = await searchSemantic(this.searchQuery, this.limit)
          // Semantic search now returns the same structure as keyword search
          this.books = data.records
          this.total = data.total
        } else {
          const data = await fetchBooks(this.searchQuery, this.filter, this.limit, this.offset);
          this.books = data.records
          this.total = data.total
        }
      } catch (e) {
        console.error("Search failed:", e)
        this.books = []
      }
    },

    prevPage() {
      if (this.offset > 0) {
        this.offset -= this.limit
        this.fetchBooks()
      }
    },
    nextPage() {
      if (this.offset + this.limit < this.total) {
        this.offset += this.limit
        this.fetchBooks()
      }
    },
    updateQueryParams() {
      const query: Record<string, string | number> = { offset: this.offset, limit: this.limit, mode: this.searchMode }
      if (this.searchQuery) {
        query.q = this.searchQuery
      }
      if (this.publisher) {
        query.publisher = this.publisher
      }
      if (this.author) {
        query.author = this.author
      }
      this.$router.push({query: query as any})
    },
    initializeFromQueryParams() {
      const query = this.$route.query
      if (query.offset) {
        this.offset = parseInt(query.offset as string, 10) || 0
      }
      if (query.limit) {
        this.limit = parseInt(query.limit as string, 10) || 12
      }
      if (query.mode) {
        this.searchMode = (query.mode as string) || 'keyword'
      }
      if (query.q) {
        this.searchQuery = (query.q as string) || ''
        this.keyword = this.searchQuery
        this.filter = []
      }
      if (query.publisher) {
        this.publisher = (query.publisher as string) || ''
        this.keyword = this.publisher
        this.filter[0] = 'publisher = "' + this.publisher + '"'
      }
      if (query.author) {
        this.author = (query.author as string) || ''
        this.keyword = this.author
        this.filter[0] = 'authors = "' + this.author + '"'
      }
    }
  },
  mounted() {
    this.fetchBooks()
  }
}
</script>

<style scoped>
.search-wrapper {
  max-width: 1400px;
  margin: var(--spacing-lg) auto;
  padding: var(--spacing-xl);
}

.search-input-wrapper {
  margin-bottom: var(--spacing-xl);
}

.search-bar {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.search-input {
  flex: 1;
}

.mode-switch {
  flex-shrink: 0;
}

.search-input-wrapper :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: none;
  transition: all 0.3s ease;
}

.search-input-wrapper :deep(.el-input-group__append) {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--glass-border);
  color: var(--el-text-color-primary);
  box-shadow: none;
}

.search-input-wrapper :deep(.el-input__wrapper:hover),
.search-input-wrapper :deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.4);
}

.search-input-wrapper :deep(.el-input__inner) {
  color: var(--el-text-color-primary);
}

.search-input-wrapper :deep(.el-input__inner::placeholder) {
  color: var(--el-text-color-placeholder);
}

.search-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: var(--spacing-md);
}

.search-title strong {
  margin-left: var(--spacing-sm);
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.search-count {
  color: var(--el-text-color-regular);
  margin-bottom: var(--spacing-lg);
  display: block;
}

.books-grid {
  margin-top: var(--spacing-lg);
}

.pagination-row {
  margin-top: var(--spacing-xl);
  gap: var(--spacing-md);
}

.glass-button {
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-sm) var(--spacing-lg);
  color: var(--el-text-color-primary);
  transition: all 0.3s ease;
}

.glass-button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.18);
  transform: translateY(-2px);
}

.glass-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
