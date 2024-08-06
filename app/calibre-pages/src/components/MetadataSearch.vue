<template>

  <div v-loading="querySearchLoading">
    <el-row >
      <el-col :span="12">
        <el-autocomplete
            v-model="query"
            :fetch-suggestions="querySearch"
            clearable
            class="w-50"
            placeholder="Please Input"
            @select="handleSelect"
        >
          <template #append>
            <el-icon @click="searchMetadata">
              <Search/>
            </el-icon>
          </template>
        </el-autocomplete>

      </el-col>
    </el-row>
    <el-table :data="tableData" height="350" style="width: 100%" highlight-current-row
              @current-change="handleCurrentChange">
      <el-table-column label="封面" width="180">
        <template #default="scope">
          <el-image
              style="width: 100px; height: 150px"
              :src="'/api/proxy/cover/' + scope.row.image"
              fit="cover"/>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" width="180"/>
      <el-table-column prop="author" label="作者" width="180"/>
      <el-table-column prop="publisher" label="出版社"/>
      <el-table-column prop="pubdate" label="发布日期"/>
      <el-table-column prop="isbn13" label="ISBN"/>
    </el-table>
  </div>
</template>
<script>
import {ElButton, ElInput, ElNotification} from 'element-plus'

export default {
  name: 'MetadataSearch',
  components: {ElButton, ElInput},
  props: {
    book: {
      type: Object,
      default: () => ({})
    },
  },
  data() {
    return {
      querySearchLoading: false,
      selectRow: {},
      query: '',
      options: [],
      tableData: [],
      form: {
        name: '',
        region: ''
      },
    }
  },
  emits: ['current-metadata'],
  created() {
    if (this.book.isbn) {
      let isbn = this.book.isbn.replace(/-/g, '');
      this.query = isbn
      this.options.push({
        value: isbn,
        label: isbn
      })
      this.options.push({
        value: this.book.title,
        label: this.book.title
      })
    } else if (this.book.title) {
      this.query = this.book.title
      this.options.push({
        value: this.book.title,
        label: this.book.title
      })
    }
    if (this.book.authors) {
      this.options.push({
        value: this.book.title + " " + this.book.authors,
        label: this.book.title + " " + this.book.authors
      })
    }
  },
  methods: {
    async searchMetadata() {
      this.querySearchLoading = true;
      if (this.query) {
        try {
          const response = await fetch('/api/metadata/search?query=' + this.query, {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json'
            },
          })
          const data = await response.json()
          if (!data.success) {
            ElNotification({
              title: '未找到相关书籍',
              message: this.query,
              type: 'warning',
            })
          }
          console.log(data)
          this.tableData = data.books
        } catch (e) {
          ElNotification({
            title: '搜索失败',
            message: 'Error: ' + e,
            type: 'error',
          })
          console.log(e)

        }
      }
      this.querySearchLoading = false;


    },
    handleCurrentChange(val) {
      this.$emit('current-metadata', val)
    },
    handleSelect(item) {
      console.log(item)
      this.query = item.value
    },
    querySearch(queryString, cb) {
      cb(this.options)
    },
  }
}
</script>

<style scoped>

</style>