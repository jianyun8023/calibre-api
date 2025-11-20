<template>
  <!-- 移除 SearchBar，统一使用 Header 搜索 -->
  <el-container class="publisher-wrapper">
    <section class="glass-container">
      <el-row>
        <el-col :span="24" class="col-bottom">
          <h2 class="section-title">出版社</h2>
        </el-col>
        <el-col v-for="publisher in publishers" :key="publisher" :span="6" :lg="6" :sm="12" :xs="24">
          <el-tag class="publisher-tag" @click="searchByPublisher(publisher)" effect="dark">
            {{ publisher }}
          </el-tag>
        </el-col>

        <el-col :span="24" class="col-top">
          <el-pagination justify="center"
                         size="small"
                         background
                         layout="prev, pager, next"
                         :total="allPublishers.length"
                         :page-size="pageSize"
                         v-model:current-page="currentPage"
                         class="pagination-wrapper"
                         @change="handleCurrentChange"
          />
        </el-col>
      </el-row>
    </section>
  </el-container>
</template>

<script lang="ts">
import {ElButton, ElCard, ElCol, ElContainer, ElInput, ElRow} from 'element-plus'
// SearchBar 已移除，统一使用 Header 搜索
import BookCard from '@/components/BookCard.vue'

export default {
  name: 'Publishers',
  components: {
    BookCard,
    // SearchBar 已移除
    ElContainer,
    ElRow,
    ElCol,
    ElInput,
    ElButton,
    ElCard
  },
  data() {
    return {
      publishers: [] as string[],
      allPublishers: [] as string[],
      pageSize: 36,
      currentPage: 1,
    }
  },
  computed: {},
  created() {
    this.initializeFromQueryParams()
    this.fetchPublishers()
  },
  // 当组件被激活时重新获取数据
  activated() {
    console.log('Publisher page activated, refreshing data...')
    this.initializeFromQueryParams()
    this.fetchPublishers()
  },
  watch: {
    pageSize() {
      this.updateQueryParams()
    },
    currentPage() {
      this.updateQueryParams()
      this.handleCurrentChange(this.currentPage)
    }
  },
  methods: {
    async fetchPublishers() {
      const response = await fetch('/api/publisher')
      const publishers = await response.json()
      this.allPublishers = publishers.data
      this.publishers = publishers.data.slice(0, this.pageSize)
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
      this.publishers = this.allPublishers.slice((val - 1) * this.pageSize, val * this.pageSize)
    },
    updateQueryParams() {
      this.$router.push({query: {...this.$route.query, page: this.currentPage, size: this.pageSize}})
    },
    initializeFromQueryParams() {
      const query = this.$route.query
      if (query.page) {
        this.currentPage = parseInt(query.page as string, 10)
      }
      if (query.size) {
        this.pageSize = parseInt(query.size as string, 10)
      }
    }
  }
}
</script>

<style scoped>
.publisher-wrapper {
  margin-top: var(--spacing-lg);
  width: 100%;
  max-width: 1400px;
  margin-left: auto;
  margin-right: auto;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-lg);
}

.section-title::after {
  content: '';
  display: block;
  width: 60px;
  height: 3px;
  background: var(--primary-gradient);
  margin-top: 8px;
  border-radius: 2px;
}

.publisher-tag {
  background: rgba(255, 255, 255, 0.12) !important;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border) !important;
  box-shadow: var(--glass-shadow);
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  margin: var(--spacing-xs);
  cursor: pointer;
  color: var(--text-primary) !important;
  transition: all 0.3s ease;
}

.publisher-tag:hover {
  transform: translateY(-4px);
  box-shadow: var(--glass-shadow-hover);
  background: rgba(255, 255, 255, 0.2) !important;
}

.col-bottom {
  margin-bottom: var(--spacing-lg);
}

.col-top {
  margin-top: var(--spacing-xl);
}

.pagination-wrapper :deep(.el-pager li) {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  color: var(--text-primary);
  margin: 0 4px;
}

.pagination-wrapper :deep(.el-pager li.is-active) {
  background: var(--primary-gradient) !important;
  color: white !important;
}

.pagination-wrapper :deep(.el-pager li:hover) {
  background: rgba(255, 255, 255, 0.15);
}

.pagination-wrapper :deep(.btn-prev),
.pagination-wrapper :deep(.btn-next) {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  color: var(--text-primary);
}

.pagination-wrapper :deep(.btn-prev:hover),
.pagination-wrapper :deep(.btn-next:hover) {
  background: rgba(255, 255, 255, 0.15);
}
</style>
