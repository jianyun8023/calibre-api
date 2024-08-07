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
        <el-col v-for="book in recentBooks" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24">
          <BookCard :book="book" :more_info="true" />
        </el-col>
      </el-row>
      <el-row class="mt-4" justify="center">
        <el-button @click="prevPage" :disabled="offset === 0">
          <el-icon>
            <ArrowLeftBold />
          </el-icon>
          上一页
        </el-button>
        <el-button @click="nextPage" :disabled="offset + limit >= estimatedTotalHits"
          >下一页
          <el-icon>
            <ArrowRightBold />
          </el-icon>
        </el-button>
      </el-row>
    </section>
  </el-container>
</template>

<script lang="ts">
import { ElButton, ElCard, ElCol, ElContainer, ElInput, ElRow } from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'
import BookCard from '@/components/BookCard.vue'
import { Book } from '@/types/book'

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
      recentBooks: [] as Book[],
      limit: 12 as number,
      offset: 0 as number,
      estimatedTotalHits: 0
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.estimatedTotalHits / this.limit)
    }
  },
  created() {
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

<style scoped></style>
