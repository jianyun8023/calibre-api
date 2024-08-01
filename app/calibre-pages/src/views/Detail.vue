<template xmlns="http://www.w3.org/1999/html">
  <el-row style="margin-bottom: 10px; margin-top: 10px">
    <SearchBar/>
  </el-row>
  <el-row style="margin-bottom: 10px; margin-top: 10px">
    <el-text v-if="book">{{ book.title }}</el-text>
    <el-text class="tag-spacing" v-if="book"
    ><strong> ID:</strong> {{ book.id }}
      <el-button plain @click="copyToClipboard(book.id)">📋</el-button>
    </el-text>
  </el-row>
  <el-row style="margin-bottom: 10px; margin-top: 10px">
    <el-col :span="8">
      <el-image style="width: 60%; height: fit-content" :src="book.cover" fit="cover"/>
    </el-col>
    <el-col :span="16">
      <div class="md:ml-6 items-stretch">
        <p class="flex gap-2">
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
          <!--          <span v-for="(author, index) in book.authors" :key="author" @click="searchByAuthor(author)"-->
          <!--                :class="{'text-blue-600': index % 2 === 0, 'text-orange-600': index % 2 !== 0}"-->
          <!--                class="cursor-pointer mr-2">{{ author }} </span>-->
        </p>
        <p class="text-gray-700 mb-2 self-auto">
          <strong class="font-mono">Publisher:</strong
          ><span @click="searchByPublisher" class="tag-spacing">{{ book.publisher }}</span>
        </p>
        <p class="text-gray-700 mb-2 self-auto font-serif">
          <strong>ISBN:</strong>
          <span class="tag-spacing">{{ book.isbn }}</span>
        </p>
        <p class="text-gray-700 mb-2 self-auto" font-serif>
          <strong>Published Date:</strong>
          <span class="tag-spacing">{{ new Date(book.pubdate).toLocaleDateString() }}</span>
        </p>
        <p class="text-gray-700 mb-2 self-auto" font-serif>
          <strong v-if="book.tags && book.tags.length">Tags:</strong>
          <el-tag v-for="item in book.tags" :key="item" effect="dark" round>
            {{ item }}
          </el-tag>
        </p>
        <p class="text-gray-700 mb-2 self-auto" font-serif>
          <strong> File Size:</strong> {{ formatFileSize(book.size) }}
        </p>

        <el-button
            type="primary"
            plain
            :disabled="!book.file_path"
            @click="redirectToDownload(book.file_path)"
        >
          下载书籍
        </el-button>
        <el-popconfirm title="确定删除?" @confirm="deleteBook(book.id)">
          <template #reference>
            <el-button class="mt-4 inline-block text-gray-700 px-4 py-2">删除书籍</el-button>
          </template>
        </el-popconfirm>
      </div>
    </el-col>
  </el-row>
  <el-row>
    <article v-if="book.comments" class="mt-8">
      <h2 class="text-xl font-bold mb-4">简介</h2>
      <p class="text-gray-700 mb-2 font-serif text-prett text-lg indent-8">{{ book.comments }}</p>
    </article>
  </el-row>
</template>

<script>
import {h} from 'vue'
import {ElButton, ElInput, ElMessage, ElNotification, ElRow} from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'

export default {
  name: 'Detail',
  components: {SearchBar, ElRow, ElButton, ElInput, ElNotification, ElMessage},
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
.tag-spacing {
  margin-right: 8px; /* Adjust the value as needed */
  margin-bottom: 8px; /* Adjust the value as needed */
  margin-left: 10px;
}
</style>
