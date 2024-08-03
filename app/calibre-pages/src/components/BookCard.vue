<template>
  <el-card class="book-card" @click="redirectToDetail(book.id)">
    <el-row type="flex" align="middle">
      <el-col :span="6" class="cover-container">
        <img
            class="book-cover"
            :src="book.cover"
            alt="book cover"
        />
      </el-col>
      <el-col :span="18" class="info-container">
        <div class="info-item title">{{ truncateText(book.title) }}</div>
        <div class="info-item author" v-if="book.authors && book.authors.length">
          {{ truncateText(book.authors.join(', ')) }}
        </div>
        <div class="info-item publisher" v-if="more_info">{{ book.publisher }}</div>
        <div class="info-item publisher" v-if="more_info">{{ book.isbn }}</div>
        <div class="info-item pubdate" v-if="book.pubdate">
          {{ new Date(book.pubdate).toLocaleDateString() }}
        </div>
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
.book-card {
  width: 100%;
  height: 150px; /* 固定卡片高度 */
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}

.cover-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-cover {
  width: 96px;
  height: 139px; /* 固定卡片高度 */
}

.info-container {
  height: 139px; /* 固定卡片高度 */
  width: 192px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.info-item {
  line-height: 1.5;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  margin-left: 20px;
}

.title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.author, .publisher, .pubdate {
  font-size: 14px;
  color: #666;
}

.el-card {
  cursor: pointer;
  transition: box-shadow 0.3s;
}

.el-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}
</style>
