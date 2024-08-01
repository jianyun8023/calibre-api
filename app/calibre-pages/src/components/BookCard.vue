<template>
  <el-card @click="redirectToDetail(book.id)">
    <el-row>
      <el-col :span="12">
        <el-image
            class="book-cover"
            style="width: 60%; height: fit-content"
            :src="book.cover"
            fit="cover"
        />
      </el-col>
      <el-col :span="12">
        <el-row>
          <el-text tag="b" class="mx-1" size="large">{{ truncateText(book.title) }}</el-text>
        </el-row>
        <el-row align="middle">
          <el-text class="mx-1" v-if="book.authors && book.authors.length">
            {{ truncateText(book.authors.join(', ')) }}
          </el-text>
        </el-row>
        <el-row align="middle">
          <el-text class="mx-1">
            {{ book.publisher }}
          </el-text>
        </el-row>
        <el-row align="middle" v-if="more_info">
          <el-text class="mx-1" v-if="book.pubdate">
            {{ new Date(book.pubdate).toLocaleDateString() }}
          </el-text>
        </el-row>
      </el-col>
    </el-row>
  </el-card>
</template>

<script>
import {ElButton, ElCard, ElCol, ElInput, ElRow} from 'element-plus'

export default {
  name: 'BookCard',
  components: {ElRow, ElCard, ElCol, ElButton, ElInput},
  props: {
    book: {
      type: Object,
      required: true
    },
    more_info: {
      type: Boolean,
      default: false
    }
  },
  methods: {
    redirectToDetail(id) {
      this.$router.push(`/detail/${id}`)
    },
    truncateText(title) {
      if (!title) return ''
      return title.length > 20 ? title.substring(0, 16) + '...' : title
    }
  }
}
</script>
<style scoped>
.book-cover {
  width: 100px; /* Adjust the width as needed */
  height: 140px; /* Adjust the height as needed */
}

.mx-1 {
  margin-bottom: 5px;
}
</style>
