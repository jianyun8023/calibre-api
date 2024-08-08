<template>
  <el-row class="detail-header">
    <SearchBar/>
  </el-row>
  <article class="detail-content">
    <el-row class="detail-row">
      <el-col :span="8" class="cover-container" :xs="24">
        <img class="book-cover" :src="book.cover" alt="book cover"/>
      </el-col>
      <el-col :span="16" :xs="24">
        <div class="book-info">
          <el-descriptions :title="book.title" :column="1" size="large" border>
            <template #extra>
              <el-button type="primary" plain @click="dialogSearchVisible = true" :icon="Refresh">
                更新元数据
              </el-button>
              <el-button type="primary" plain @click="editBook" :icon="Edit">
                编辑
              </el-button>
            </template>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <Box/>
                  </el-icon>
                  ID
                </div>
              </template>
              <el-button text bg @click="copyToClipboard(book.id)">{{ book.id }}📋</el-button>
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <user/>
                  </el-icon>
                  Authors
                </div>
              </template>
              <el-tag
                  class="tag-spacing"
                  v-for="item in book.authors"
                  :key="item"
                  effect="dark"
                  @click="searchByAuthor(item)"
              >
                {{ item }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <Discount/>
                  </el-icon>
                  Publisher
                </div>
              </template>
              <span @click="searchByPublisher">{{ book.publisher }}</span>
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon class="el-icon">
                    <Key/>
                  </el-icon>
                  ISBN
                </div>
              </template>
              {{ book.isbn }}
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <Timer/>
                  </el-icon>
                  Published Date
                </div>
              </template>
              <span class="tag-spacing">{{ new Date(book.pubdate).toLocaleDateString() }}</span>
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <Trophy/>
                  </el-icon>
                  Rating
                </div>
              </template>
              <el-rate
                  :value="book.rating / 2"
                  @input="(val: number) => (book.rating = val * 2)"
                  show-score
                  text-color="#ff9900"
                  :max="5"
                  allow-half
                  :score-template="`${book.rating}分`"
              >
              </el-rate>
            </el-descriptions-item>
            <el-descriptions-item v-if="book.tags && book.tags.length">
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <CollectionTag/>
                  </el-icon>
                  Tags
                </div>
              </template>
              <el-tag v-for="item in book.tags" :key="item" effect="dark" round>
                {{ item }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item>
              <template #label>
                <div class="cell-item">
                  <el-icon>
                    <Document/>
                  </el-icon>
                  File Size
                </div>
              </template>
              {{ formatFileSize(book.size) }}
            </el-descriptions-item>
          </el-descriptions>
          <el-row class="book-buttons">
            <el-button color="#626aef" :xs="24" :icon="Menu" plain @click="showBookMenu">
              预览目录
            </el-button>
          </el-row>
          <el-row class="book-buttons">
            <el-button
                color="#626aef"
                :xs="24"
                :icon="Download"
                plain
                :disabled="!book.file_path"
                @click="redirectToDownload(book.file_path)"
            >
              下载书籍
            </el-button>
            <el-popconfirm title="确定删除?" @confirm="deleteBook(book.id)">
              <template #reference>
                <el-button :icon="Delete" :xs="24" class="delete-button">删除书籍</el-button>
              </template>
            </el-popconfirm>
          </el-row>
        </div>
      </el-col>
    </el-row>
    <el-row>
      <article v-if="book.comments" class="book-comments">
        <h2 class="comments-title">简介</h2>
        <p class="comments-text" v-html="book.comments"></p>
      </article>
    </el-row>
  </article>

  <el-dialog
      v-model="dialogSearchVisible"
      title="搜索元数据"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <MetadataSearch :book="book" @current-metadata="handleCurrentMeta"/>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogSearchVisible = false">取消</el-button>
        <el-button type="primary" @click="handleClose"> 确认</el-button>
      </div>
    </template>
  </el-dialog>
  <el-dialog
      v-model="dialogEditVisible"
      title="更新元数据"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <MetadataEdit :book="book" :new-book="currentRow" :update-metadata-flag="triggerUpdate"/>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogEditVisible = false">取消</el-button>
        <el-button type="primary" @click="triggerUpdate = true">更新</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
      v-model="dialogMenuVisible"
      title="查看目录"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <el-row class="margin-top" v-loading="menuLoding">
      <el-tree
          style="max-width: 600px"
          :data="bookMenu"
          :props="defaultProps"
          @node-click="handleNodeClick"
      />
    </el-row>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogMenuVisible = false">OK</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script lang="ts">
import {h} from 'vue'
import {ElButton, ElCol, ElInput, ElMessage, ElNotification, ElRow} from 'element-plus'
import SearchBar from '@/components/SearchBar.vue'
import MetadataSearch from '@/components/MetadataSearch.vue'
import MetadataEdit from '@/components/MetadataEdit.vue'
import {Delete, Download, Edit, Menu, Rank, Refresh, Trophy} from '@element-plus/icons-vue'
import {Book} from '@/types/book'

export default {
  name: 'Detail',
  computed: {
    Refresh() {
      return Refresh
    },
    Edit() {
      return Edit
    },
    Delete() {
      return Delete
    },
    Download() {
      return Download
    },
    Menu() {
      return Menu
    }
  },
  components: {
    Trophy,
    Rank,
    MetadataEdit,
    MetadataSearch,
    ElCol,
    SearchBar,
    ElRow,
    ElButton,
    ElInput,
    ElNotification,
    ElMessage
  },
  props: {
    id: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      book: {} as Book,
      bookMenu: {} as any,
      menuLoding: false as boolean,
      defaultProps: {
        children: 'points',
        label: 'text'
      },
      dialogSearchVisible: false,
      dialogEditVisible: false as boolean,
      dialogMenuVisible: false,
      currentRow: {} as any,
      triggerUpdate: false as boolean,
      isPhone: document.documentElement.clientWidth < 993
    }
  },
  created() {
    this.fetchBook((this.$route as any).params.id)
  },
  mounted() {
    window.addEventListener('resize', () => {
      this.isPhone = document.documentElement.clientWidth < 993 // 小于993视为平板及手机
      console.log('isPhone: ' + this.isPhone)
    })
  },

  methods: {
    async fetchBook(id: string) {
      try {
        const response = await fetch(`/api/book/${id}`)
        if (!response.ok) throw new Error('Network response was not ok')
        this.book = await response.json()
      } catch (error) {
        console.error('There was a problem with the fetch operation:', error)
      }
    },
    formatFileSize(size: number) {
      if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
      return (size / 1024 / 1024).toFixed(2) + ' MB'
    },

    async showBookMenu() {
      this.dialogMenuVisible = true

      this.menuLoding = true
      try {
        const response = await fetch(`/api/read/${this.book.id}/toc`)
        if (!response.ok) throw new Error('Network response was not ok')
        const data = await response.json()
        this.bookMenu = data.points
        this.menuLoding = false
        if (!data.points) {
          ElNotification({
            title: 'ID copied ' + text,
            message: 'ID copied to clipboard',
            type: 'warning'
          })
          this.dialogMenuVisible = false
        }
        console.log(data.points)
      } catch (error) {
        this.menuLoding = false
        console.error('There was a problem with the fetch operation:', error)
      }
    },

    handleNodeClick(data: any) {
      console.log(data)
      window.open("/api" + data.content.src, '_blank')
    },
    copyToClipboard(text: string) {
      navigator.clipboard
          .writeText(text)
          .then(() => {
            ElNotification({
              title: 'ID copied ' + text,
              message: h('i', {style: 'color: teal'}, 'ID copied to clipboard'),
              type: 'success'
            })
          })
          .catch((err) => {
            ElNotification({
              title: 'ID copied ' + text,
              message: h('i', {style: 'color: red'}, 'Oops, Could not copy text.'),
              type: 'error'
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
    searchByAuthor(author: string) {
      this.$router.push({
        path: '/search',
        query: {
          author: author
        }
      })
    },
    handleCurrentMeta(currentMeta: any) {
      this.currentRow = currentMeta
      console.log(this.currentRow)
    },
    handleClose() {
      this.dialogSearchVisible = false
      console.log(this.currentRow)
      this.dialogEditVisible = true
    },
    editBook() {
      this.currentRow = {}
      this.dialogEditVisible = true
    },
    redirectToHome() {
      this.$router.push('/')
    },
    redirectToDownload(url: string) {
      window.location.href = url
    },
    joinTags(tags: string[]) {
      if (tags.length === 0) return ''
      return tags.join(', ')
    },
    async deleteBook(bookId: string) {
      const response = await fetch(`/api/book/${bookId}/delete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      })
      if (response.ok) {
        ElNotification({
          title: 'Book deleted successfully',
          message: this.book.title,
          type: 'success'
        })
        this.$router.back()
      } else {
        ElNotification({
          title: '删除书籍失败',
          message: h('i', {style: 'color: red'}, this.book.title),
          type: 'error'
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
  padding-top: 80px;
}

.book-cover {
  width: 60%; /* 固定宽度 */
  height: auto; /* 固定高度 */
}

@media (max-width: 768px) {
  .book-cover {
    width: 60%; /* 手机上宽度60% */
  }

  .detail-content {
    padding: 20px 0;
  }

  .book-info {
    padding-top: 30px;
    padding-left: 30px;
  }

  .cover-container {
    padding-top: 10px;
  }
}

.book-info {
  padding-left: 20px;
}

.info-item {
  margin-bottom: 10px;
}

.tag-spacing {
  margin-right: 10px;
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

.el-descriptions {
  margin-top: 20px;
}

.cell-item {
  display: flex;
  align-items: center;
}

.el-icon {
  padding-right: 5px;
}

.margin-top {
  margin-top: 20px;
}
</style>
