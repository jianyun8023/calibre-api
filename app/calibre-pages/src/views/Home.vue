<template>
  <el-row>
    <SearchBar/>
  </el-row>
  <el-row>
    <el-row class="full-width-row">
      <h2>最近更新</h2>
      <el-link href="/books" class="text-right"
      >更多
        <el-icon>
          <More/>
        </el-icon
        >
      </el-link>
    </el-row>
    <el-row :gutter="20">
      <el-col v-for="book in recentBooks" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24">
        <BookCard :book="book"/>
      </el-col>
    </el-row>
  </el-row>

  <el-row>
    <el-row class="full-width-row">
      <h2>便便看看</h2>
      <el-button link class="text-right" @click="randomSomeBooks">
        换换
        <el-icon>
          <Refresh/>
        </el-icon>
      </el-button>
    </el-row>
    <el-row :gutter="20">
      <el-col v-for="book in randomBooks" :key="book.id" :span="6" :lg="6" :sm="12" :xs="24">
        <BookCard :book="book"/>
      </el-col>
    </el-row>
  </el-row>

  <el-row>
    <el-row class="full-width-row">
      <el-col :span="24" class="full-width-row">
        <h2>出版社</h2>
        <el-link href="/publisher" class="text-right"
        >更多
          <el-icon>
            <More/>
          </el-icon
          >
        </el-link>
      </el-col>
      <el-col v-for="publisher in publishers" :key="publisher" :span="6" :lg="6" :sm="12" :xs="24">
        <el-tag @click="searchByPublisher(publisher)" effect="light">
          {{ publisher }}
        </el-tag>
      </el-col>

      <el-col :span="24" class="col-top">
        <el-pagination justify="center"
                       size="small"
                       background
                       layout="prev, pager, next"
                       :total="allPublishers.length"
                       :page-size="publisherPage"
                       class="mt-4"
                       @change="handleCurrentChange"
        />
      </el-col>
    </el-row>
  </el-row>
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
  methods: {
    async fetchRecentBooks() {
      this.recentBooks = await fetchRecentBooks(12, 0).then(res => res.records)
    },
    async fetchPublishers() {
      this.allPublishers = await fetchPublishers()
      this.publishers = this.allPublishers.slice(0, this.publisherPage)
    },
    async randomSomeBooks() {
      this.randomBooks = await fetchRandomBooks()
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

<style scoped>
.full-width-row {
  display: flex;
  justify-content: space-between;
  width: 100%;
}

.text-right {
  margin-left: auto;
}

.col-top {
  margin-top: 20px;
}

.col-bottom {
  margin-bottom: 20px;
}
</style>
