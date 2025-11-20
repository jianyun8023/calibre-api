<template>
  <div class="search-wrapper glass-container">
    <div class="search-input-wrapper">
      <el-affix target=".search-wrapper">
        <el-input
            v-model="searchQuery"
            @input="fetchBooks"
            type="text"
            placeholder="书名、作者、ISBN"
            size="large"
        />
      </el-affix>
    </div>
    <h2 class="search-title">
      搜索结果：
      <strong>{{ keyword }}</strong>
    </h2>
    <el-text class="search-count"
    >共计 {{ total }} 条, 当前{{ offset }} --
      {{ offset + limit >= total ? total : offset + limit }}
    </el-text>

    <el-row :gutter="20" class="books-grid">
      <el-col v-for="book in books" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24" class="book-col-spacing">
        <BookCard :book="book" :more_info="true"/>
      </el-col>
    </el-row>
    <el-row class="pagination-row" justify="center">
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
import {ElButton, ElCol, ElInput, ElRow} from 'element-plus'
import {fetchBooks} from "@/api/api";

export default {
  name: 'Search',
  components: {ElInput, ElButton, ElRow, ElCol, BookCard},
  data() {
    return {
      searchQuery: '',
      keyword: '',
      publisher: '',
      author: '',
      books: [],
      filter: [],
      limit: 12,
      offset: 0,
      total: 0
    }
  },
  created() {
    this.initializeFromQueryParams()
  },
  // 当组件被激活时重新获取数据
  activated() {
    console.log('Search page activated, refreshing data...')
    this.initializeFromQueryParams()
    this.fetchBooks()
  },
  watch: {
    searchQuery() {
      this.updateQueryParams()
      this.fetchBooks()
    },
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
    async fetchBooks() {
      const data = await fetchBooks(this.searchQuery, this.filter, this.limit, this.offset);
      this.books = data.records
      this.total = data.total
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
      let query = {...this.$route.query, offset: this.offset, limit: this.limit}
      if (this.searchQuery) {
        query.q = this.searchQuery
      }
      if (this.publisher) {
        query.publisher = this.publisher
      }
      if (this.author) {
        query.author = this.author
      }
      this.$router.push({query: query})
    },
    initializeFromQueryParams() {
      const query = this.$route.query
      if (query.offset) {
        this.offset = parseInt(query.offset, 10)
      }
      if (query.limit) {
        this.limit = parseInt(query.limit, 10)
      }
      if (query.q) {
        this.searchQuery = query.q
        this.keyword = this.searchQuery
        this.filter = []
      }
      if (query.publisher) {
        this.publisher = query.publisher
        this.keyword = this.publisher
        this.filter[0] = 'publisher = "' + this.publisher + '"'
      }
      if (query.author) {
        this.author = query.author
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

.search-input-wrapper :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: none;
  transition: all 0.3s ease;
}

.search-input-wrapper :deep(.el-input__wrapper:hover),
.search-input-wrapper :deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.4);
}

.search-input-wrapper :deep(.el-input__inner) {
  color: var(--text-primary);
}

.search-input-wrapper :deep(.el-input__inner::placeholder) {
  color: var(--text-tertiary);
}

.search-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
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
  color: var(--text-secondary);
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
  color: var(--text-primary);
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
