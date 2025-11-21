<template>
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
                更新
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
              <el-tag 
                v-for="item in book.tags" 
                :key="item" 
                effect="dark" 
                round
                @click="searchByTag(item)"
              >
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
            <el-button color="#626aef" :xs="24" :icon="Menu" plain @click="dialogPreviewVisible = true">
              预览目录
            </el-button>
            <el-button color="#626aef" :xs="24" :icon="Coffee" plain @click="readBook">
              阅读
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


  <MetadataSearch :book="book"
                  v-model:dialogSearchVisible="dialogSearchVisible"/>


  <MetadataEdit :book="book" v-model:dialogEditVisible="dialogEditVisible"
                />
  <PreviewBook :book="book" v-model:dialog-preview-visible="dialogPreviewVisible"
               />

</template>

<script lang="ts">
import {h} from 'vue'
import {ElButton, ElCol, ElInput, ElMessage, ElNotification, ElRow} from 'element-plus'
import MetadataSearch from '@/components/MetadataSearch.vue'
import MetadataEdit from '@/components/MetadataEdit.vue'
import MetadataUpdate from '@/components/MetadataUpdate.vue'
import {Coffee, Delete, Download, Edit, Menu, Rank, Refresh, Trophy} from '@element-plus/icons-vue'
import {Book} from '@/types/book'
import PreviewBook from "@/components/PreviewBook.vue";
import {copyToClipboard, formatFileSize} from "@/utils/utils";
import {deleteBook, fetchBook} from "@/api/api";

export default {
  name: 'Detail',
  computed: {
    Coffee() {
      return Coffee
    },
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
    PreviewBook,
    MetadataUpdate,
    Trophy,
    Rank,
    MetadataEdit,
    MetadataSearch,
    ElCol,
    ElRow,
    ElButton,
    ElInput,
    ElNotification,
    ElMessage,
  },
  // setup() {
  //   const book = ref<Book>({} as Book)
  //   const route = useRoute()
  //   const fetchBook = async (id: string) => {
  //     try {
  //       const response = await fetch(`/api/book/${id}`)
  //       if (!response.ok) throw new Error('Network response was not ok')
  //       book.value = await response.json()
  //     } catch (error) {
  //       console.error('There was a problem with the fetch operation:', error)
  //     }
  //   }
  //
  //   onMounted(() => {
  //     fetchBook(route.params.id as string)
  //   })
  //
  //   provide('book', book)
  //
  //   return {
  //     book
  //   }
  // },
  props: {
    id: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      book: {} as Book,

      dialogSearchVisible: false,
      dialogEditVisible: false as boolean,
      dialogPreviewVisible: false,
      currentRow: {} as any,
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
  // 监听路由变化，当路由参数改变时重新获取数据
  watch: {
    '$route'(to, from) {
      // 当路由参数变化时，重新获取书籍数据
      if (to.params.id && to.params.id !== from.params.id) {
        this.fetchBook(to.params.id as string)
      }
    }
  },
  // 路由守卫：在同一组件内路由参数变化时调用
  beforeRouteUpdate(to, from, next) {
    // 当路由参数变化时，重新获取书籍数据
    if (to.params.id) {
      this.fetchBook(to.params.id as string)
    }
    next()
  },

  methods: {
    async fetchBook(id: string) {
      try {
        this.book = await fetchBook(id)
      } catch (error) {
        console.error('There was a problem with the fetch operation:', error)
      }
    },
    formatFileSize,
    copyToClipboard,
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
    searchByTag(tag: string) {
      this.$router.push({
        path: '/search',
        query: {
          tag: tag
        }
      })
    },
    editBook() {
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
      const data = await deleteBook(Number(bookId))
      if (data) {
        ElNotification({
          title: 'Book deleted successfully',
          message: this.book.title,
          type: 'success'
        })
        this.$router.back()
      }
    },
    readBook() {
      // this.$router.push(`/read/${this.id}`)
      window.open(`/read/${this.id}`, '_blank')
    }
  }
}
</script>

<style scoped lang="scss">
.detail-content {
  padding: var(--spacing-lg);
  max-width: 1400px;
  margin: 0 auto;
}

.detail-row {
  margin-bottom: var(--spacing-xl);
}

.cover-container {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: var(--spacing-lg);
}

.book-cover {
  width: 100%;
  max-width: 300px;
  height: auto;
  border-radius: var(--border-radius-lg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  transition: transform 0.3s ease;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.2);
  }
}

.book-info {
  padding: 0 var(--spacing-lg);
}

/* Glassmorphism Table Styles - 使用全局样式 */
/* .el-descriptions 样式已移至 index.scss */

.cell-item {
  display: flex;
  align-items: center;
  gap: 6px;
  
  .el-icon {
    font-size: 16px;
    opacity: 0.7;
  }
}

/* Tags Styling */
:deep(.el-tag) {
  background: var(--glass-bg-medium);
  border: 1px solid var(--glass-border);
  color: var(--el-text-color-primary);
  margin-right: var(--spacing-xs);
  margin-bottom: var(--spacing-xs);
  cursor: pointer;
  transition: all 0.2s ease;
  
  &:hover {
    background: var(--glass-bg-strong);
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
}

/* Button Groups - 使用更轻量的样式 */
.book-buttons {
  margin-top: var(--spacing-lg);
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
  
  .el-button {
    flex: 1;
    min-width: 140px;
    background: var(--glass-bg-medium);
    backdrop-filter: blur(8px);
    border: 1px solid var(--glass-border);
    color: var(--el-text-color-primary);
    font-weight: 500;
    transition: all 0.3s ease;
    
    &:hover:not(:disabled) {
      background: var(--el-color-primary);
      color: white;
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
      border-color: var(--el-color-primary);
    }
    
    &.delete-button {
      background: rgba(245, 108, 108, 0.1);
      border-color: rgba(245, 108, 108, 0.3);
      color: #f56c6c;
      
      &:hover {
        background: #f56c6c;
        color: white;
        border-color: #f56c6c;
        box-shadow: 0 4px 12px rgba(245, 108, 108, 0.3);
      }
    }
    
    &:disabled {
      opacity: 0.4;
      cursor: not-allowed;
      transform: none;
    }
  }
}

/* Comments Section */
.book-comments {
  background: var(--glass-bg-light);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  border: 1px solid var(--glass-border);
  box-shadow: var(--glass-shadow);
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-xl);
  margin-top: var(--spacing-xl);
  color: var(--el-text-color-primary);
}

.comments-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: var(--spacing-md);
  color: var(--el-text-color-primary);
  position: relative;
  padding-bottom: var(--spacing-sm);
  
  &::after {
    content: '';
    position: absolute;
    left: 0;
    bottom: 0;
    width: 60px;
    height: 3px;
    background: var(--primary-gradient);
    border-radius: 2px;
  }
}

.comments-text {
  font-size: 1rem;
  color: var(--el-text-color-regular);
  line-height: 1.8;
  text-indent: 2em;
  
  :deep(p) {
    margin-bottom: var(--spacing-md);
  }
}

/* Rating Component */
:deep(.el-rate) {
  .el-rate__icon {
    font-size: 20px;
    margin-right: 4px;
  }
  
  .el-rate__text {
    color: var(--el-text-color-regular);
    font-weight: 600;
  }
}

/* Mobile Responsiveness */
@media (max-width: 768px) {
  .detail-content {
    padding: var(--spacing-md);
  }
  
  .cover-container {
    padding: var(--spacing-md) 0;
  }
  
  .book-cover {
    max-width: 200px;
  }
  
  .book-info {
    padding: var(--spacing-md) 0;
  }
  
  :deep(.el-descriptions) {
    .el-descriptions__header {
      padding: var(--spacing-md);
      
      .el-descriptions__title {
        font-size: 1.25rem;
      }
      
      .el-descriptions__extra {
        flex-direction: column;
        width: 100%;
        
        .el-button {
          width: 100%;
        }
      }
    }
    
    .el-descriptions__table .el-descriptions__label {
      width: 100px;
      font-size: 0.75rem;
    }
  }
  
  .book-buttons {
    flex-direction: column;
    
    .el-button {
      width: 100%;
      min-width: unset;
    }
  }
  
  .comments-title {
    font-size: 1.25rem;
  }
  
  .comments-text {
    font-size: 0.9375rem;
    text-indent: 0;
  }
}
</style>
