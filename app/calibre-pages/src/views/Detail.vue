<template>
  <el-row class="detail-header">
    <SearchBar/>
  </el-row>
  <el-row class="detail-header">
    <el-text class="book-title" v-if="book">{{ book.title }}</el-text>
    <el-text class="book-id" v-if="book">
      <strong>ID: </strong> {{ book.id }}
      <el-button plain @click="copyToClipboard(book.id)">📋</el-button>
    </el-text>
  </el-row>
  <article class="detail-content">
    <el-row class="detail-row">
      <el-col :span="8" class="cover-container">
        <img
            class="book-cover"
            :src="book.cover"
            alt="book cover"
        />
      </el-col>
<!--      <el-col :span="8">-->
<!--        <el-image class="book-cover" :src="book.cover" fit="cover"/>-->
<!--      </el-col>-->
      <el-col :span="16">
        <div class="book-info">
          <p class="info-item">
            <strong>Authors:</strong>
            <el-tag
                class="tag-spacing"
                v-for="item in book.authors"
                :key="item"
                effect="dark"
                @click="searchByAuthor(item)"
            >
              {{ item }}
            </el-tag>
          </p>
          <p class="info-item">
            <strong class="font-mono">Publisher:</strong>
            <span @click="searchByPublisher" class="tag-spacing">{{ book.publisher }}</span>
          </p>
          <p class="info-item font-serif">
            <strong>ISBN:</strong>
            <span class="tag-spacing">{{ book.isbn }}</span>
          </p>
          <p class="info-item font-serif">
            <strong>Published Date:</strong>
            <span class="tag-spacing">{{ new Date(book.pubdate).toLocaleDateString() }}</span>
          </p>
          <p class="info-item font-serif" v-if="book.tags && book.tags.length">
            <strong>Tags:</strong>
            <el-tag v-for="item in book.tags" :key="item" effect="dark" round>
              {{ item }}
            </el-tag>
          </p>
          <p class="info-item font-serif">
            <strong>File Size:</strong> {{ formatFileSize(book.size) }}
          </p>
          <el-row class="book-buttons">
            <el-button
                type="primary"
                :xs="24"
                plain
                :disabled="!book.file_path"
                @click="redirectToDownload(book.file_path)"

            >
              下载书籍
            </el-button>
            <el-popconfirm title="确定删除?" @confirm="deleteBook(book.id)">
              <template #reference>
                <el-button :xs="24" class="delete-button">删除书籍</el-button>
              </template>
            </el-popconfirm>
          </el-row>

        </div>
      </el-col>
    </el-row>
    <el-row>
      <article v-if="book.comments" class="book-comments">
        <h2 class="comments-title">简介</h2>
        <p class="comments-text">{{ book.comments }}</p>
      </article>
    </el-row>
  </article>
</template>

<script>
import {h} from 'vue'
import {ElButton, ElCol, ElInput, ElMessage, ElNotification, ElRow} from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'

export default {
  name: 'Detail',
  components: {ElCol, SearchBar, ElRow, ElButton, ElInput, ElNotification, ElMessage},
  data() {
    return {
      id: '',
      book: {}
    }
  },
  created() {
    this.fetchBook(this.$route.params.id)
  },
  // mounted() {
  //   console.log('ID:', this.$route.params.id);
  //   this.fetchBook(this.$route.params.id);
  // },
  methods: {
    async fetchBook(id) {
      try {
        const response = await fetch(`/api/book/${id}`)
        if (!response.ok) throw new Error('Network response was not ok')
        this.book = await response.json()
      } catch (error) {
        console.error('There was a problem with the fetch operation:', error)
      }
    },
    formatFileSize(size) {
      if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
      return (size / 1024 / 1024).toFixed(2) + ' MB'
    },
    copyToClipboard(text) {
      navigator.clipboard
          .writeText(text)
          .then(() => {
            ElNotification({
              title: 'ID copied ' + text,
              message: h('i', {style: 'color: teal'}, 'ID copied to clipboard'),
              type: 'success',
            })
          })
          .catch((err) => {
            ElNotification({
              title: 'ID copied ' + text,
              message: h('i', {style: 'color: red'}, 'Oops, Could not copy text.'),
              type: 'error',
            })
          })
    },
    searchByPublisher() {
      this.$router.push({
        path: '/search',
        query: {
          publisher: this.book.publisher
        }
      })
    },
    searchByAuthor(author) {
      this.$router.push({
        path: '/search',
        query: {
          author: author
        }
      })
    },
    redirectToHome() {
      this.$router.push('/')
    },
    redirectToDownload(url) {
      window.location.href = url
    },
    joinTags(tags) {
      if (tags.length === 0) return ''
      return tags.join(', ')
    },
    async deleteBook(bookId) {
      const url = `https://lib.pve.icu/cdb/delete-books/${bookId}`
      try {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          }
        })
        if (response.ok) {

          const response2 = await fetch(`/api/book/${bookId}/delete`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            }
          })
          if (response2.ok) {
            ElNotification({
              title: 'Book deleted successfully',
              message: this.book.title,
              type: 'success',
            })
          } else {
            ElNotification({
              title: '删除书籍成功，但索引删除失败',
              message: h('i', {style: 'color: red'}, this.book.title),
              type: 'error',
            })
          }
          this.$router.go(-1)
          // 这里可以添加其他逻辑，例如刷新页面或更新视图
        } else {
          ElNotification({
            title: 'Failed to delete book',
            message: h('i', {style: 'color: red'}, this.book.title),
            type: 'error',
          })
        }
      } catch (error) {
        ElNotification({
          title: 'An error occurred while deleting the book',
          message: h('i', {style: 'color: red'}, this.book.title + ' Error: ' + error.message),
          type: 'error',
        })
      }
    }
  }
}
</script>

<style scoped>

.detail-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  margin-top: 10px;
}

.book-title {
  font-size: 1.5rem;
  font-weight: bold;
  color: #333;
  margin-right: 20px;
  margin-left: 10px;
}

.book-id {
  display: flex;
  align-items: center;
  margin-left: 10px;
}


.detail-content {
  padding: 20px;
}

.detail-row {
  margin-bottom: 10px;
  margin-top: 10px;
}

.cover-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-cover {
  width: 50%; /* 固定宽度 */
  height: auto; /* 固定高度 */
}

.book-info {
  padding-left: 20px;
}

.info-item {
  margin-bottom: 10px;
}

.tag-spacing {
  margin-right: 8px;
  margin-bottom: 8px;
  margin-left: 10px;
}

.delete-button {
  color: #ff4d4f;
}

.book-comments {
  margin-top: 20px;
}

.book-buttons {
  margin-top: 40px;
}

.comments-title {
  font-size: 1.5rem;
  font-weight: bold;
  margin-bottom: 10px;
}

.comments-text {
  font-size: 1.125rem;
  color: #4a4a4a;
  text-indent: 2em;
}
</style>
