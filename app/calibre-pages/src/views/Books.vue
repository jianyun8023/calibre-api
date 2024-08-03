<template>
  <el-row>
    <SearchBar />
  </el-row>
  <el-container class="mt-8 w-full md:w-2/3">
    <section>
      <el-row class="mb-4">
        <el-col :span="24">
          <h2 class="text-xl font-bold">全部书籍</h2>
        </el-col>
      </el-row>
      <el-row :gutter="20">
        <el-col v-for="book in recentBooks" :key="book.id" :span="6" :lg="6" :sm="8" :xs="24">
          <BookCard :book="book" :more_info="true" />
        </el-col>
      </el-row>
      <el-row class="mt-4" justify="center">
        <el-button @click="prevPage" :disabled="offset === 0"><el-icon><ArrowLeftBold /></el-icon>上一页</el-button>
        <el-button @click="nextPage" :disabled="offset + limit >= estimatedTotalHits"
          >下一页<el-icon><ArrowRightBold /></el-icon></el-button
        >
      </el-row>
    </section>
  </el-container>
</template>

<script>
import { ElContainer, ElRow, ElCol, ElInput, ElButton, ElCard } from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'
import BookCard from '@/components/BookCard.vue'

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
    ElCard
  },
  data() {
    return {
      searchQuery: '',
      recentBooks: [],
      limit: 12,
      offset: 0,
      estimatedTotalHits: 0
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.estimatedTotalHits / this.limit)
    }
  },
  created() {
    this.fetchBooks()
  },
  methods: {
    async fetchBooks() {
      const response = await fetch(`/api/recently?limit=${this.limit}&offset=${this.offset}`)
      const books = await response.json()
      this.recentBooks = books.hits
      this.estimatedTotalHits = books.estimatedTotalHits
    },
    prevPage() {
      if (this.offset > 0) {
        this.offset -= this.limit
        this.fetchBooks()
      }
    },
    nextPage() {
      if (this.offset + this.limit < this.estimatedTotalHits) {
        this.offset += this.limit
        this.fetchBooks()
      }
    },
    redirectToSearch() {
      window.location.href = `/search?query=${encodeURIComponent(this.searchQuery)}`
    },
    redirectToDetail(id) {
      window.location.href = `/detail/${id}`
    },
    redirectToHome() {
      console.log('redirecting to home')
      window.location.href = '/'
    },
    truncateText(text) {
      if (text === undefined || text === null) {
        return ''
      }
      return text.length > 20 ? text.substring(0, 16) + '...' : text
    }
  }
}
</script>

<style scoped></style>
