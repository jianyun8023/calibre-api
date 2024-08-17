<template>
  <div v-loading="querySearchLoading">
    <el-row>
      <el-col :span="12">
        <el-autocomplete
            v-model="query"
            :fetch-suggestions="querySearch"
            value-key="value"
            clearable
            class="w-50"
            placeholder="Please Input"
            @select="handleSelect"
        >
          <template #append>
            <el-button @click="searchMetadata" :icon="Search" type="success">搜索</el-button>
          </template>
        </el-autocomplete>
      </el-col>
    </el-row>
    <el-table
        :data="tableData"
        height="350"
        style="width: 100%"
        highlight-current-row
        @current-change="handleCurrentChange"
        :fit="false"
    >
      <el-table-column label="封面" width="180">
        <template #default="scope">
          <el-image
              style="width: 100px; height: 150px"
              :src="'/api/proxy/cover/' + scope.row.image"
              fit="cover"
          />
        </template>
      </el-table-column>
      <el-table-column prop="title" :formatter="joinTitle" label="标题" width="180"/>
      <el-table-column prop="author" label="作者" width="180"/>
      <el-table-column prop="publisher" label="出版社"/>
      <el-table-column prop="pubdate" label="发布日期"/>
      <el-table-column prop="isbn13" label="ISBN"/>
    </el-table>
  </div>
</template>
<script setup lang="ts">
import {defineEmits, onMounted, ref} from 'vue'
import {ElButton, ElNotification} from 'element-plus'
import {Search} from '@element-plus/icons-vue'
import {Book, MetaBook} from "@/types/book";

const props = defineProps<{
  book: Book;
}>();

const querySearchLoading = ref(false)
const selectRow = ref({})
const query = ref('')
type Option = { value: string; label: string }
const options = ref<Option[]>([])
const tableData = ref([] as MetaBook[])

// const book = ref({
//   isbn: '',
//   title: '',
//   authors: ''
// })

const searchMetadata = async () => {
  querySearchLoading.value = true
  if (query.value) {
    try {
      const response = await fetch('/api/metadata/search?query=' + query.value, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      })
      const data = await response.json()
      if (!data.success) {
        ElNotification({
          title: '未找到相关书籍',
          message: query.value,
          type: 'warning'
        })
      }
      console.log(data)
      tableData.value = data.books
    } catch (e) {
      ElNotification({
        title: '搜索失败',
        message: 'Error: ' + e,
        type: 'error'
      })
      console.log(e)
    }
  }
  querySearchLoading.value = false
}

const emit = defineEmits(['current-metadata'])

const handleCurrentChange = (val: MetaBook) => {
  emit('current-metadata', val)
}

const joinTitle = (row: MetaBook) => {
  if (row.sub_title) {
    return row.title + '：' + row.sub_title
  } else {
    return row.title
  }
}

const handleSelect = (item: { value: string; label: string }) => {
  console.log(item)
  query.value = item.value
}

const querySearch = (queryString: string, cb: (options: Option[]) => void) => {
  cb(options.value)
}


onMounted(() => {
  if (props.book.isbn) {
    let isbn = props.book.isbn.replace(/-/g, '')
    query.value = isbn
    options.value.push({
      value: isbn,
      label: isbn
    })
    options.value.push({
      value: props.book.title,
      label: props.book.title
    })
  } else if (props.book.title) {
    query.value = props.book.title
    options.value.push({
      value: props.book.title,
      label: props.book.title
    })
  }
  if (props.book.authors) {
    options.value.push({
      value: props.book.title + ' ' + props.book.authors,
      label: props.book.title + ' ' + props.book.authors
    })
  }
})


</script>

<style scoped></style>