<template>
  <el-row>
    <SearchBar />
  </el-row>
  <el-row>
    <el-row class="full-width-row">
      <h2>最近更新</h2>
      <el-link href="/books" class="text-right">更多</el-link>
    </el-row>
    <el-row :gutter="20">
      <el-col v-for="book in recentBooks" :key="book.id" :span="6">
        <BookCard :book="book" />
      </el-col>
    </el-row>
  </el-row>

  <el-row>
    <h2>文学</h2>
    <el-row :gutter="20">
      <!-- Repeat the book item structure here -->
    </el-row>
  </el-row>
  <el-row>
    <h2>社会文化</h2>
    <el-row :gutter="20">
      <!-- Repeat the book item structure here -->
    </el-row>
  </el-row>
  <el-row>
    <h2>历史</h2>
    <el-row :gutter="20">
      <!-- Repeat the book item structure here -->
    </el-row>
  </el-row>
</template>

<script>
import { ElButton, ElCard, ElCol, ElContainer, ElInput, ElLink, ElRow } from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'
import BookCard from '@/components/BookCard.vue'

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
      recentBooks: []
    }
  },
  created() {
    this.fetchRecentBooks()
  },
  methods: {
    async fetchRecentBooks() {
      const response = await fetch('/api/recently?limit=12')
      const books = await response.json()
      this.recentBooks = books.hits
    }
  }
}
</script>

<style scoped>
.full-width-row {
  display: flex;
  justify-content: space-between;
  width: 100%;
}

.text-right {
  margin-left: auto;
}
</style>
