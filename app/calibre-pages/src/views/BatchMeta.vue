<template>
  <div class="flex justify-center mb-8">
    <el-row>
      <el-col :span="1">
        <el-text>查询：</el-text>
      </el-col>
      <el-col :span="2">
        <el-select
            v-model="filterType"
            placeholder="类型"
            @change="fetchBooks"
            placement="bottom-start"
        >
          <el-option label="出版社" value="publisher"/>
          <el-option label="作者" value="author"/>
          <el-option label="ISBN" value="isbn"/>
        </el-select>
      </el-col>
      <el-col :span="8" :offset="1">
        <el-input
            v-model="keyword"
            @input="fetchBooks"
            type="text"
            placeholder="书名、作者、ISBN"
            class=""
        />
      </el-col>
    </el-row>
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
    <el-table
        ref="multipleTable"
        row-key="id"
        :data="books"
        :border="true"
        highlight-current-row
        :show-overflow-tooltip="true"
        stripe
        style="width: 100%"
        :row-class-name="tableRowClassName"
        @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55"/>
      <el-table-column prop="id" label="ID" width="100"/>
      <el-table-column prop="title" label="标题" width="180"/>
      <el-table-column prop="authors" label="作者" width="180"/>
      <el-table-column prop="isbn" label="ISBN"/>
      <el-table-column prop="publisher" label="出版社"/>
      <el-table-column
          prop="pubdate"
          label="出版日期"
          :formatter="(row: Book) => new Date(row.pubdate).toLocaleDateString()"
      >
      </el-table-column>
    </el-table>
    <div style="margin-top: 20px">
      <el-button @click="toggleSelection">选择有ISBN的书籍</el-button>
      <el-button @click="exclusionPackage">排除套装</el-button>
      <el-button @click="clearSelection">清除选择</el-button>
      <el-button @click="updateMetaData">更新书籍元数据</el-button>
    </div>
  </el-row>
  <el-row class="mt-4" justify="center">
    <el-select
        class="w-20"
        v-model="limit"
        placeholder="分页"
        placement="bottom-start"
    >
      <el-option label="10" value="10"/>
      <el-option label="20" value="20"/>
      <el-option label="50" value="50"/>
    </el-select>
    <el-button @click="prevPage" :disabled="offset === 0">
      <el-icon>
        <ArrowLeftBold/>
      </el-icon>
      上一页
    </el-button>
    <el-button @click="nextPage" :disabled="offset + limit >= estimatedTotalHits"
    >下一页
      <el-icon>
        <ArrowRightBold/>
      </el-icon>
    </el-button>
  </el-row>

  <el-dialog v-model="metaUpdateDialogVisible" :title="'更新 ' + metaUpdate.index + '/' + metaUpdate.total " width="500"
             center :close-on-click-modal="false" :close-on-press-escape="false">
    <el-row>
      <el-col :span="12">
        <el-text>当前书籍：</el-text>
        <BookCard :book="metaUpdate.currentBook" :more_info="true"/>
      </el-col>
      <el-col :span="12" v-loading="metaUpdate.updating == 0">
        <el-text>新元数据：</el-text>
        <BookCard v-if="metaUpdate.updating == 1 || metaUpdate.updating == 2"
                  :book="mapMetaBookToBook(metaUpdate.newMeta)" :proxy_image="true" :more_info="true"/>
        <el-text v-if="metaUpdate.updating == -1">更新失败</el-text>
        <el-text v-if="metaUpdate.updating == 3">更新完成，成功数量 {{ metaUpdate.successCount }}/
          {{ metaUpdate.total }}
        </el-text>
      </el-col>

    </el-row>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="metaUpdateDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="metaUpdateDialogVisible = false">
          Confirm
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script lang="ts">
import {Book, mapMetaBookToBook, MetaBook} from '@/types/book'
import BookCard from '@/components/BookCard.vue'
import {ElButton, ElCol, ElInput, ElRow, ElTable} from 'element-plus'

export default {
  name: 'BatchMeta',
  components: {ElInput, ElButton, ElRow, ElCol, BookCard},
  data() {
    return {
      filterType: 'publisher' as string,
      keyword: '' as string,
      books: [] as Book[],
      multipleSelection: [] as Book[],
      filter: [] as string[],
      limit: 12 as number,
      offset: 0 as number,
      estimatedTotalHits: 0 as number,
      metaUpdateDialogVisible: false,
      metaUpdate: {
        currentBook: {} as Book,
        total: 0 as number,
        index: 0 as number,
        successCount: 0 as number,
        updating: 0,
        newMeta: {} as MetaBook,
      },
    }
  },
  created() {
    this.initializeFromQueryParams()
  },
  watch: {
    keyword() {
      this.updateQueryParams()
      this.fetchBooks()
    },
    filterType() {
      this.updateQueryParams()
      this.fetchBooks()
    },
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
    mapMetaBookToBook,
    async fetchBooks() {
      if (this.filterType === 'publisher') {
        this.filter[0] = 'publisher = "' + this.keyword + '"'
      }
      if (this.filterType === 'author') {
        this.filter[0] = 'authors = "' + this.keyword + '"'
      }
      if (this.filterType === 'isbn') {
        this.filter[0] = 'isbn = "' + this.keyword + '"'
      }

      const response = await fetch('/api/search?q=', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          Filter: this.filter,
          Limit: this.limit as number,
          Offset: this.offset
        })
      })
      const data = await response.json()
      this.books = data.hits
      this.estimatedTotalHits = data.estimatedTotalHits
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
        // 套装 共\d+册 全\d+册
        if (row.isbn && (row.title.includes('套装') || row.title.includes('册'))) {
          ;(this.$refs.multipleTable as any).toggleRowSelection(row, false)
        }

        // if (row.isbn && row.title.includes('套装')) {
        //   ;(this.$refs.multipleTable as any).toggleRowSelection(row, false)
        // }
      })
    },
    toggleSelection() {
      this.books.forEach((row) => {
        if (row.isbn) {
          ;(this.$refs.multipleTable as any).toggleRowSelection(row, true)
        }
      })
    },
    handleSelectionChange(val: Book[]) {
      console.log(val)
      this.multipleSelection = val
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
          const response = await fetch('/api/metadata/isbn/' + book.isbn, {
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

<style scoped>
.el-table .warning-row {
  --el-table-tr-bg-color: var(--el-color-success-light-9);
}

.w-20 {
  width: 70px;
  margin-right: 10px;
}
</style>
