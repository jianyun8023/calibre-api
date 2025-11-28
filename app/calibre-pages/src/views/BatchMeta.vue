<template>
  <div class="batch-meta-wrapper glass-container">
    <div class="header-actions mb-8">
      <el-row :gutter="20" align="middle">
        <el-col :span="6" :xs="24" class="mb-4-xs">
          <div class="filter-group">
            <el-select
                v-model="filterType"
                placeholder="类型"
                @change="handleFilterTypeChange"
                placement="bottom-start"
                class="glass-select"
            >
              <el-option label="全部（按时间倒序）" value="all"/>
              <el-option label="书名" value="title"/>
              <el-option label="出版社" value="publisher"/>
              <el-option label="作者" value="author"/>
              <el-option label="ISBN" value="isbn"/>
            </el-select>
          </div>
        </el-col>
        <el-col :span="18" :xs="24">
          <el-autocomplete
              v-model="keyword"
              @input="handleKeywordInput"
              :fetch-suggestions="querySearch"
              :trigger-on-focus="false"
              clearable
              class="glass-input w-full"
              :placeholder="getPlaceholder()"
              @select="handleSearchSelect"
              :disabled="filterType === 'all'"
          >
            <template #prefix>
              <el-icon class="search-icon"><Search /></el-icon>
            </template>
            <template #default="{ item }">
              <div class="value">{{ item }}</div>
            </template>
          </el-autocomplete>
        </el-col>
      </el-row>
    </div>

    <div class="results-header mb-4">
      <h2 class="text-xl font-bold glass-text-title">
        搜索结果
        <span v-if="keyword" class="keyword-highlight">"{{ keyword }}"</span>
      </h2>
      <el-text class="glass-text-secondary">
        共计 {{ total }} 条, 当前显示 {{ offset + 1 }} - {{ Math.min(offset + limit, total) }}
      </el-text>
    </div>

    <div class="table-container glass-panel mb-6">
      <!-- 加载骨架屏 -->
      <div v-if="loadingBooks" class="skeleton-table">
        <el-skeleton :rows="10" animated />
      </div>
      
      <!-- 实际表格 -->
      <el-table
          v-else
          ref="multipleTable"
          row-key="id"
          :data="books"
          :border="false"
          highlight-current-row
          :show-overflow-tooltip="true"
          style="width: 100%"
          :row-class-name="tableRowClassName"
          @selection-change="handleSelectionChange"
          class="glass-table"
      >
        <el-table-column type="selection" width="55"/>
        <el-table-column type="expand">
          <template #default="props">
            <div class="expand-row-content">
              <el-row :gutter="24">
                <el-col :span="6" :xs="24" class="mb-4-xs">
                  <div class="book-cover-wrapper">
                    <el-image
                        class="book-cover"
                        :src="props.row.cover"
                        fit="cover"
                    >
                      <template #error>
                        <div class="image-slot">
                          <el-icon><Picture /></el-icon>
                        </div>
                      </template>
                    </el-image>
                  </div>
                </el-col>
                <el-col :span="18" :xs="24">
                  <el-descriptions :title="props.row.title" :column="1" size="large" border>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Box/></el-icon> ID
                        </div>
                      </template>
                      <el-button text bg size="small" @click="copyToClipboard(props.row.id)" class="copy-btn">
                        {{ props.row.id }} <el-icon class="ml-1"><CopyDocument /></el-icon>
                      </el-button>
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><User/></el-icon> Authors
                        </div>
                      </template>
                      <div class="tag-group">
                        <el-tag
                            v-for="item in props.row.authors"
                            :key="item"
                            effect="light"
                            class="glass-tag"
                        >
                          {{ item }}
                        </el-tag>
                      </div>
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Discount/></el-icon> Publisher
                        </div>
                      </template>
                      {{ props.row.publisher }}
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Key/></el-icon> ISBN
                        </div>
                      </template>
                      {{ props.row.isbn }}
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Timer/></el-icon> Published Date
                        </div>
                      </template>
                      {{ new Date(props.row.pubdate).toLocaleDateString() }}
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Trophy/></el-icon> Rating
                        </div>
                      </template>
                      <el-rate
                          :model-value="props.row.rating / 2"
                          disabled
                          show-score
                          text-color="var(--el-color-warning)"
                          score-template="{value}分"
                      />
                    </el-descriptions-item>
                    <el-descriptions-item>
                      <template #label>
                        <div class="cell-item">
                          <el-icon><Document/></el-icon> File Size
                        </div>
                      </template>
                      {{ formatFileSize(props.row.size) }}
                    </el-descriptions-item>
                  </el-descriptions>
                </el-col>
              </el-row>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="80" show-overflow-tooltip/>
        <el-table-column label="标题" min-width="200">
          <template #default="scope">
            <div class="title-cell">
              <el-tooltip content="查看封面" placement="top" effect="light">
                <template #content>
                  <el-image
                      style="width: 120px; height: 160px"
                      :src="scope.row.cover"
                      fit="cover"
                  />
                </template>
                <span class="book-title" @click="goToSearch(scope.row)">{{ scope.row.title }}</span>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column v-if="filterType !== 'author'" prop="authors" label="作者" width="150" show-overflow-tooltip/>
        <el-table-column v-if="filterType !== 'isbn'" prop="isbn" label="ISBN" width="140"/>
        <el-table-column v-if="filterType !== 'publisher'" prop="publisher" label="出版社" width="150" show-overflow-tooltip/>
        <el-table-column
            prop="pubdate"
            label="出版日期"
            width="120"
            :formatter="(row: Book) => new Date(row.pubdate).toLocaleDateString()"
        />
        <el-table-column fixed="right" label="操作" width="200" align="center">
          <template #default="scope">
            <div class="action-buttons">
              <el-tooltip content="预览" placement="top">
                <el-button circle size="small" class="glass-btn-icon" @click="previewBook(scope.row)">
                  <el-icon><View /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="更新元数据" placement="top">
                <el-button circle size="small" type="primary" class="glass-btn-icon" @click="updateBook(scope.row)">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="编辑" placement="top">
                <el-button circle size="small" type="warning" class="glass-btn-icon" @click="updateEditBook(scope.row)">
                  <el-icon><Edit /></el-icon>
                </el-button>
              </el-tooltip>
              <el-popconfirm title="确定删除?" @confirm="deleteBook(scope.row)">
                <template #reference>
                  <el-button circle size="small" type="danger" class="glass-btn-icon">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="actions-footer glass-panel p-4 mb-6">
      <el-row justify="space-between" align="middle" :gutter="20">
        <el-col :span="16" :xs="24" class="mb-4-xs">
          <div class="batch-actions">
            <el-button class="glass-button" @click="toggleSelection">
              <el-icon class="mr-1"><Select /></el-icon> 选择有ISBN的书籍
            </el-button>
            <el-button class="glass-button" @click="exclusionPackage">
              <el-icon class="mr-1"><Remove /></el-icon> 排除套装
            </el-button>
            <el-button class="glass-button" @click="clearSelection">
              <el-icon class="mr-1"><Close /></el-icon> 清除选择
            </el-button>
            <el-button type="primary" class="glass-button primary" @click="updateMetaData">
              <el-icon class="mr-1"><Refresh /></el-icon> 批量更新元数据
            </el-button>
            <el-popconfirm title="确定删除选中书籍?" @confirm="batchDelete">
              <template #reference>
                <el-button type="danger" class="glass-button danger">
                  <el-icon class="mr-1"><Delete /></el-icon> 批量删除
                </el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-col>
        <el-col :span="8" :xs="24" class="text-right">
          <div class="pagination-wrapper">
            <el-select
                v-model="limit"
                class="glass-select w-24 mr-2"
                size="small"
            >
              <el-option label="10条/页" :value="10"/>
              <el-option label="20条/页" :value="20"/>
              <el-option label="50条/页" :value="50"/>
              <el-option label="100条/页" :value="100"/>
            </el-select>
            <el-button-group class="glass-pagination">
              <el-button size="small" @click="prevPage" :disabled="offset === 0">
                <el-icon><ArrowLeft /></el-icon>
              </el-button>
              <el-button size="small" @click="nextPage" :disabled="offset + limit >= total">
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </el-button-group>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- Dialogs remain mostly the same structure but can be styled globally via CSS -->
    <el-dialog v-model="metaUpdateDialogVisible" :title="'批量更新进度 ' + metaUpdate.index + '/' + metaUpdate.total " width="600px"
               center :close-on-click-modal="false" :close-on-press-escape="false" class="glass-dialog">
      <el-row :gutter="20">
        <el-col :span="12">
          <div class="dialog-section-title">当前书籍</div>
          <BookCard :book="metaUpdate.currentBook" :more_info="true" :proxy_image="false"/>
        </el-col>
        <el-col :span="12" v-loading="metaUpdate.updating == 0">
          <div class="dialog-section-title">新元数据</div>
          <BookCard v-if="metaUpdate.updating == 1 || metaUpdate.updating == 2"
                    :book="mapMetaBookToBook(metaUpdate.newMeta)" :proxy_image="true" :more_info="true"/>
          <div v-if="metaUpdate.updating == -1" class="status-text error">更新失败</div>
          <div v-if="metaUpdate.updating == 3" class="status-text success">
            更新完成，成功数量 {{ metaUpdate.successCount }}/{{ metaUpdate.total }}
          </div>
        </el-col>
      </el-row>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="metaUpdateDialogVisible = false" class="glass-button">关闭</el-button>
        </div>
      </template>
    </el-dialog>

    <MetadataSearch :book="editBook" v-model:dialogSearchVisible="dialogSearchVisible"/>
    <MetadataEdit :book="editBook" v-model:dialogEditVisible="dialogEditVisible"/>
    <PreviewBook :book="editBook" v-model:dialog-preview-visible="dialogPreviewVisible"/>
  </div>
</template>

<script lang="ts">
import {Book, mapMetaBookToBook, MetaBook} from '@/types/book'
import BookCard from '@/components/BookCard.vue'
import {ElButton, ElCol, ElInput, ElNotification, ElRow, ElTable} from 'element-plus'
import MetadataEdit from "@/components/MetadataEdit.vue";
import {
  Delete, Menu, Search, Box, User, Discount, Key, Timer, Trophy, 
  CollectionTag, Document, Picture, CopyDocument, View, Edit, Refresh,
  Select, Remove, Close, ArrowLeft, ArrowRight
} from "@element-plus/icons-vue";
import {h} from "vue";
import MetadataSearch from "@/components/MetadataSearch.vue";
import PreviewBook from "@/components/PreviewBook.vue";
import {copyToClipboard, formatFileSize} from "@/utils/utils";
import {deleteBook, fetchBooks, fetchPublishers} from "@/api/api";

export default {
  name: 'BatchMeta',
  components: {
    Search, PreviewBook, MetadataSearch, MetadataEdit, ElInput, ElButton, ElRow, ElCol, BookCard,
    Box, User, Discount, Key, Timer, Trophy, CollectionTag, Document, Picture, CopyDocument,
    View, Edit, Refresh, Select, Remove, Close, ArrowLeft, ArrowRight, Delete
  },
  data() {
    return {
      // 加载状态
      loadingBooks: true,
      loadingPublishers: true,
      
      filterType: 'all',
      keyword: '',
      books: [],
      multipleSelection: [],
      filter: [],
      limit: 20,
      offset: 0,
      total: 0,
      metaUpdateDialogVisible: false,
      metaUpdate: {
        currentBook: {} as Book,
        total: 0 as number,
        index: 0 as number,
        successCount: 0 as number,
        updating: 0,
        newMeta: {} as MetaBook,
      },
      allPublishers: [] as string[],
      dialogSearchVisible: false,
      dialogEditVisible: false,
      dialogPreviewVisible: false,
      editBook: {} as Book
    }
  },
  created() {
    this.initializeFromQueryParams()
    this.fetchPublishers()
  },
  watch: {
    offset() {
      this.updateQueryParams()
      this.fetchBooks()
    },
    limit() {
      this.offset = 0
      this.updateQueryParams()
      this.fetchBooks()
    }
  },

  methods: {
    async fetchPublishers() {
      this.allPublishers = await fetchPublishers()
    },
    mapMetaBookToBook,
    formatFileSize,
    copyToClipboard,
    async fetchBooks() {
      this.loadingBooks = true
      try {
        let searchQuery = "";
        let sort = [];
        this.filter = [];
        
        if (this.filterType === 'all') {
          // 按时间倒序显示全部书籍，不需要任何过滤
          searchQuery = "";
          sort = ["last_modified:desc"];
        } else if (this.filterType === 'title') {
          // 按书名搜索
          searchQuery = this.keyword;
        } else if (this.filterType === 'publisher') {
          this.filter[0] = `publisher = "${this.keyword}"`;
        } else if (this.filterType === 'author') {
          this.filter[0] = `authors = "${this.keyword}"`;
        } else if (this.filterType === 'isbn') {
          this.filter[0] = `isbn = "${this.keyword}"`;
        }
        
        const data = await fetchBooks(searchQuery, this.filter, this.limit, this.offset, sort);

        this.books = data.records
        this.total = data.total
      } finally {
        this.loadingBooks = false
      }
    },
    
    handleFilterTypeChange() {
      // 切换类型时重置offset和keyword
      this.offset = 0;
      if (this.filterType === 'all') {
        this.keyword = '';
      }
      this.updateQueryParams();
      this.fetchBooks();
    },
    
    handleKeywordInput() {
      if (this.filterType !== 'all') {
        this.offset = 0;
        this.updateQueryParams();
        this.fetchBooks();
      }
    },
    
    getPlaceholder() {
      const placeholders = {
        'all': '显示全部书籍（按时间倒序）',
        'title': '输入书名搜索...',
        'publisher': '输入出版社名称...',
        'author': '输入作者名称...',
        'isbn': '输入ISBN号码...'
      };
      return placeholders[this.filterType] || '搜索...';
    },

    async querySearch(queryString: string, cb: (arg0: string[]) => void) {
      if (this.filterType === 'publisher') {
        const results = queryString ? this.allPublishers.filter(this.createFilter(queryString)) : this.allPublishers
        console.log(results)
        cb(results)
      } else {
        cb([])
      }
    },
    createFilter(queryString: string) {
      return (restaurant: string) => {
        return (restaurant.toLowerCase().indexOf(queryString.toLowerCase()) === 0)
      }
    },
    handleSearchSelect(item: string) {
      this.keyword = item
    },
    prevPage() {
      if (this.offset > 0) {
        this.offset -= this.limit
        this.fetchBooks()
      }
    },
    nextPage() {
      if (this.offset + this.limit < this.total) {
        this.offset += this.limit
        this.fetchBooks()
      }
    },
    updateQueryParams() {
      let query = {
        ...this.$route.query,
        offset: this.offset,
        limit: this.limit,
        keyword: this.keyword,
        filterType: this.filterType
      }
      this.$router.push({query: query})
    },
    initializeFromQueryParams() {
      const query = this.$route.query
      if (query.offset) {
        this.offset = parseInt(query.offset as string, 10)
      }
      if (query.limit) {
        this.limit = parseInt(query.limit as string, 10)
      }
      if (query.keyword) {
        this.keyword = query.keyword as string
      }
      if (query.filterType) {
        this.filterType = query.filterType as string
      }
    },
    tableRowClassName({row, rowIndex}: { row: Book; rowIndex: number }) {
      if (!row.isbn) {
        return 'warning-row'
      }
      return ''
    },
    clearSelection() {
      ;(this.$refs.multipleTable as any).clearSelection()
    },
    exclusionPackage() {
      this.multipleSelection.forEach((row) => {
        if (row.isbn && (row.title.includes('套装') || row.title.includes('册'))) {
          ;(this.$refs.multipleTable as any).toggleRowSelection(row, false)
        }
      })
    },
    async deleteBook(book: Book) {
      try {
        await deleteBook(book.id);

        ElNotification({
          title: 'Book deleted successfully',
          message: book.title,
          type: 'success'
        })
        this.fetchBooks() // Refresh list after delete
      } catch (error) {
        ElNotification({
          title: '删除书籍失败',
          message: h('i', {style: 'color: red'}, book.title),
          type: 'error'
        })
      }
    },
    updateEditBook(book: Book) {
      this.editBook = book
      this.dialogEditVisible = true
    },
    updateBook(book: Book) {
      this.editBook = book
      this.dialogSearchVisible = true
    },
    previewBook(book: Book) {
      this.editBook = book
      this.dialogPreviewVisible = true
    },
    toggleSelection() {
      this.books.forEach((row) => {
        if (row.isbn) {
          ;(this.$refs.multipleTable as any).toggleRowSelection(row, true)
        }
      })
    },
    goToSearch(book: Book) {
      const {href} = this.$router.resolve({
        path: '/search',
        query: {
          q: book.title
        }
      });
      window.open(href, "_blank");
    },
    handleSelectionChange(val: Book[]) {
      console.log(val)
      this.multipleSelection = val
    },
    async batchDelete() {
      await Promise.all(this.multipleSelection.map(book => deleteBook(book.id)));
      // 等待1s后刷新
      setTimeout(() => {
        this.fetchBooks()
      }, 1000)
    },
    async updateMetaData() {
      this.metaUpdateDialogVisible = true
      this.metaUpdate.successCount = 0
      this.metaUpdate.updating = 0
      this.metaUpdate.total = this.multipleSelection.length
      this.metaUpdate.index = 0
      for (const book of this.multipleSelection) {
        this.metaUpdate.updating = 0
        this.metaUpdate.newMeta = {} as MetaBook
        this.metaUpdate.currentBook = book
        this.metaUpdate.index = 1 + this.metaUpdate.index

        if (!book.isbn) {
          this.metaUpdate.updating = -1
          continue
        }
        try {
          let isbn = book.isbn.replace(/-/g, '')
          const response = await fetch('/api/metadata/isbn/' + isbn, {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json'
            },
          })
          const data = await response.json()
          if (data.success) {
            console.log('更新成功')
            this.metaUpdate.updating = 1
            this.metaUpdate.newMeta = data.books[0] as MetaBook


            const response = await fetch(`/api/book/${book.id}/update`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify(mapMetaBookToBook(this.metaUpdate.newMeta))
            })
            if (response.ok) {
              this.metaUpdate.updating = 2
              this.metaUpdate.successCount = 1 + this.metaUpdate.successCount
            } else {
              this.metaUpdate.updating = -1
            }
          } else {
            this.metaUpdate.updating = -1
          }

        } catch (e) {
          this.metaUpdate.updating = -1
        }
      }
      this.metaUpdate.updating = 3
    }
  },
  mounted() {
    this.fetchBooks()
  }
}
</script>

<style scoped lang="scss">
.batch-meta-wrapper {
  padding: var(--spacing-lg);
  min-height: 100%;
}

.glass-text-title {
  color: var(--el-text-color-primary);
  margin-bottom: var(--spacing-xs);
}

.glass-text-secondary {
  color: var(--el-text-color-regular);
}

.keyword-highlight {
  color: var(--el-color-primary);
}

/* Glass Table Styles - 已移至 index.scss */
/* .glass-table, .glass-button, .glass-tag, .glass-descriptions 样式已移至 index.scss */

.book-title {
  color: var(--el-text-color-primary);
  cursor: pointer;
  font-weight: 500;
  transition: color 0.2s;
  
  &:hover {
    color: var(--el-color-primary);
  }
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.glass-btn-icon {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid var(--glass-border);
  color: var(--el-text-color-primary);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  
  &.el-button--primary {
    color: var(--el-color-primary);
    background: rgba(59, 130, 246, 0.1);
    border-color: rgba(59, 130, 246, 0.3);
    
    &:hover {
      background: rgba(59, 130, 246, 0.2);
    }
  }
  
  &.el-button--warning {
    color: #e6a23c;
    background: rgba(230, 162, 60, 0.1);
    border-color: rgba(230, 162, 60, 0.3);
    
    &:hover {
      background: rgba(230, 162, 60, 0.2);
    }
  }
  
  &.el-button--danger {
    color: #f56c6c;
    background: rgba(245, 108, 108, 0.1);
    border-color: rgba(245, 108, 108, 0.3);
    
    &:hover {
      background: rgba(245, 108, 108, 0.2);
    }
  }
}

/* Batch Actions Footer */
.batch-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}



/* Expand Row Styles */
.expand-row-content {
  padding: 20px;
  background: rgba(0, 0, 0, 0.02);
}

.book-cover-wrapper {
  border-radius: var(--border-radius-md);
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  transition: transform 0.3s ease;
  
  &:hover {
    transform: scale(1.02);
  }
}

.book-cover {
  width: 100%;
  height: auto;
  display: block;
}



.cell-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}



/* Dialog Styles */
.dialog-section-title {
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--el-text-color-primary);
}

.status-text {
  margin-top: 10px;
  font-weight: 500;
  
  &.success { color: #67c23a; }
  &.error { color: #f56c6c; }
}

/* Skeleton Screen Loading State */
.skeleton-table {
  padding: var(--spacing-xl);
  min-height: 500px;
  
  :deep(.el-skeleton) {
    --el-skeleton-color: rgba(255, 255, 255, 0.1);
    --el-skeleton-to-color: rgba(255, 255, 255, 0.15);
  }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .batch-meta-wrapper {
    padding: var(--spacing-md);
  }
  
  .mb-4-xs {
    margin-bottom: 16px;
  }
  
  .batch-actions {
    justify-content: center;
    
    .el-button {
      width: 100%;
      margin-left: 0 !important;
      margin-bottom: 8px;
    }
  }
  
  .pagination-wrapper {
    justify-content: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }
}
</style>
