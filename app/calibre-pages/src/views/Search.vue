<template>
  <div class="flex justify-center mb-8">

    <div class="affix-container">
      <el-affix target=".affix-container">
        <el-input
            v-model="searchQuery"
            @input="fetchBooks"
            type="text"
            placeholder="书名、作者、ISBN"
            class=""
        />
      </el-affix>
    </div>


  </div>
  <h2 class="text-xl font-bold mb-4">
    搜索结果：
    <strong style="margin-left: 10px">{{ keyword }}</strong>
  </h2>
  <el-text
    >共计 {{ estimatedTotalHits }} 条, 当前{{ offset }} --
    {{ offset + limit >= estimatedTotalHits ? estimatedTotalHits : offset + limit }}
  </el-text>

  <el-row :gutter="20">
    <el-col v-for="book in books" :key="book.id" :span="6" :lg="6" :sm="8" :xs="24">
      <BookCard :book="book" :more_info="true" />
    </el-col>
  </el-row>
  <el-row class="mt-4" justify="center">
    <el-button @click="prevPage" :disabled="offset === 0"><el-icon><ArrowLeftBold /></el-icon>上一页</el-button>
    <el-button @click="nextPage" :disabled="offset + limit >= estimatedTotalHits">下一页<el-icon><ArrowRightBold /></el-icon></el-button>
  </el-row>
</template>

<script>
import BookCard from '@/components/BookCard.vue'
import {ElButton, ElCol, ElInput, ElRow} from 'element-plus'

export default {
  name: 'Search',
  components: {ElInput, ElButton, ElRow, ElCol, BookCard },
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
      estimatedTotalHits: 0
    }
  },
  created() {
    console.log(this.$route)
    if (this.$route.query.q){
      this.searchQuery = this.$route.query.q
    }
    if (this.$route.query.publisher) {
      this.publisher = this.$route.query.publisher
    }
    if (this.$route.query.author) {
      this.author = this.$route.query.author
    }
  },
  watch: {
    searchQuery() {
      this.keyword = this.searchQuery
      this.filter = []
    },
    publisher() {
      this.keyword = this.publisher
      this.filter[0] = 'publisher = "' + this.publisher + '"'
    },
    author() {
      this.keyword = this.author
      this.filter[0] = 'authors = "' + this.author + '"'
    }
  },

  methods: {
    async fetchBooks() {
      const response = await fetch('/api/search?q=' + this.searchQuery, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          Filter: this.filter,
          Limit: this.limit,
          Offset: this.offset
        })
      })
      const data = await response.json()
      this.books = data.hits
      this.estimatedTotalHits = data.estimatedTotalHits

      // Clear the query parameter from the URL
      const url = new URL(window.location)
      url.searchParams.delete('query')
      window.history.replaceState({}, '', url)
    },
    redirectToDetail(id) {
      window.location.href = `/detail/${id}`
    },
    redirectToHome() {
      window.location.href = '/'
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
    }
  },
  mounted() {
    this.fetchBooks()
  }
}
</script>

<style scoped></style>
