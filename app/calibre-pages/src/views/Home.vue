<template>
  <div class="home-page">
    <SearchBar/>
    
    <section class="glass-container fade-in">
      <div class="section-header">
        <h2 class="section-title">最近更新</h2>
        <el-link href="/books" class="section-link">
          更多
          <el-icon><More/></el-icon>
        </el-link>
      </div>
      
      <!-- 加载骨架屏 -->
      <el-row v-if="loadingRecent" :gutter="20">
        <el-col v-for="n in 4" :key="'skeleton-recent-' + n" :span="6" :lg="6" :sm="12" :xs="24">
          <el-skeleton :rows="6" animated />
        </el-col>
      </el-row>
      
      <!-- 实际内容 -->
      <el-row v-else :gutter="20">
        <el-col v-for="book in recentBooks" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24">
          <BookCard :book="book"/>
        </el-col>
      </el-row>
    </section>

    <section class="glass-container fade-in">
      <div class="section-header">
        <h2 class="section-title">便便看看</h2>
        <el-button link class="refresh-button" @click="randomSomeBooks">
          换换
          <el-icon><Refresh/></el-icon>
        </el-button>
      </div>
      
      <!-- 加载骨架屏 -->
      <el-row v-if="loadingRandom" :gutter="20">
        <el-col v-for="n in 4" :key="'skeleton-random-' + n" :span="6" :lg="6" :sm="12" :xs="24">
          <el-skeleton :rows="6" animated />
        </el-col>
      </el-row>
      
      <!-- 实际内容 -->
      <el-row v-else :gutter="20">
        <el-col v-for="book in randomBooks" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24">
          <BookCard :book="book"/>
        </el-col>
      </el-row>
    </section>

    <section class="glass-container fade-in">
      <div class="section-header">
        <h2 class="section-title">出版社</h2>
        <el-link href="/publisher" class="section-link">
          更多
          <el-icon><More/></el-icon>
        </el-link>
      </div>
      <div class="publishers-grid">
        <el-tag 
          v-for="publisher in publishers" 
          :key="publisher" 
          @click="searchByPublisher(publisher)" 
          effect="light"
          class="publisher-tag"
        >
          {{ publisher }}
        </el-tag>
      </div>
      <el-col :span="24" class="pagination-wrapper">
        <el-pagination
          justify="center"
          size="small"
          background
          layout="prev, pager, next"
          :total="allPublishers.length"
          :page-size="publisherPage"
          @change="handleCurrentChange"
        />
      </el-col>
    </section>
  </div>
</template>

<script lang="ts">
import {ElButton, ElCard, ElCol, ElContainer, ElInput, ElLink, ElRow} from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'
import BookCard from '@/components/BookCard.vue'
import {Book} from '@/types/book'
import {fetchPublishers, fetchRandomBooks, fetchRecentBooks} from "@/api/api";

export default {
  name: 'Home',
  components: {
    BookCard,
    SearchBar,
    ElContainer,
    ElRow,
    ElCol,
    ElInput,
    ElButton,
    ElLink,
    ElCard
  },
  data() {
    return {
      // 加载状态
      loadingRecent: true,
      loadingRandom: true,
      loadingPublishers: true,
      
      recentBooks: [] as Book[],
      randomBooks: [] as Book[],
      publishers: [] as string[],
      allPublishers: [] as string[],
      publisherPage: 8,
    }
  },
  created() {
    this.fetchRecentBooks()
    this.fetchPublishers()
    this.randomSomeBooks()
  },
  // 当组件被激活时重新获取数据（解决从其他页面返回时内容为空的问题）
  activated() {
    console.log('Home page activated, refreshing data...')
    this.fetchRecentBooks()
    this.fetchPublishers()
    this.randomSomeBooks()
  },
  // 监听路由变化
  watch: {
    '$route'(to, from) {
      // 当路由变化到 Home 页面时，重新获取数据
      if (to.path === '/' || to.name === 'Home') {
        console.log('Route changed to Home, refreshing data...')
        this.fetchRecentBooks()
        this.fetchPublishers()
        this.randomSomeBooks()
      }
    }
  },
  methods: {
    async fetchRecentBooks() {
      this.loadingRecent = true
      try {
        this.recentBooks = await fetchRecentBooks(12, 0).then(res => res.records)
      } finally {
        this.loadingRecent = false
      }
    },
    async fetchPublishers() {
      this.loadingPublishers = true
      try {
        this.allPublishers = await fetchPublishers()
        this.publishers = this.allPublishers.slice(0, this.publisherPage)
      } finally {
        this.loadingPublishers = false
      }
    },
    async randomSomeBooks() {
      this.loadingRandom = true
      try {
        this.randomBooks = await fetchRandomBooks()
      } finally {
        this.loadingRandom = false
      }
    },
    searchByPublisher(publisher: string) {
      this.$router.push({
        path: '/search',
        query: {
          publisher: publisher
        }
      })
    },
    handleCurrentChange(val: number) {
      console.log(`当前页: ${val}`)
      this.publishers = this.allPublishers.slice((val - 1) * this.publisherPage, val * this.publisherPage)
    }
  },

}
</script>

<style scoped lang="scss">
.home-page {
  animation: fadeIn 0.5s ease-out;
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

.section-link,
.refresh-button {
  color: var(--text-primary);
  font-weight: 500;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 4px;
  
  &:hover {
    transform: translateX(4px);
    color: var(--text-primary);
  }
}

.refresh-button {
  &:hover {
    transform: rotate(90deg) scale(1.1);
  }
}

.publishers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.publisher-tag {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(8px);
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  cursor: pointer;
  transition: var(--transition-normal);
  padding: var(--spacing-sm) var(--spacing-md);
  
  &:hover {
    background: rgba(255, 255, 255, 0.18);
    transform: translateY(-2px);
    box-shadow: var(--glass-shadow);
  }
}

.pagination-wrapper {
  margin-top: var(--spacing-lg);
  display: flex;
  justify-content: center;
  
  :deep(.el-pagination) {
    .btn-prev,
    .btn-next,
    .el-pager li {
      background: rgba(255, 255, 255, 0.1);
      backdrop-filter: blur(8px);
      border: 1px solid var(--glass-border);
      color: var(--text-primary);
      transition: var(--transition-fast);
      
      &:hover {
        background: rgba(255, 255, 255, 0.2);
      }
      
      &.is-active {
        background: var(--primary-gradient);
        color: white;
        border-color: transparent;
      }
    }
  }
}

@media (max-width: 768px) {
  .section-title {
    font-size: 1.5rem;
  }
  
  .publishers-grid {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  }
}
</style>
