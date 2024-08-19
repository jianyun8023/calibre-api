<template>

  <el-dialog
      :model-value="dialogSearchVisible"
      @update:model-value="val => emit('update:dialogSearchVisible', val)"
      title="搜索元数据"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
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
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCloseSearch">取消</el-button>
        <el-button type="primary" @click="nextUpdate"> 确认</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
      v-model="dialogUpdateVisible"
      title="更新元数据"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <MetadataUpdate :book="book" :new-book="selectRow" :update-metadata-flag="triggerUpdate"/>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogUpdateVisible = false">取消</el-button>
        <el-button type="primary" @click="triggerUpdate = true">更新</el-button>
      </div>
    </template>
  </el-dialog>


</template>
<script setup lang="ts">
import {defineEmits, inject, onMounted, ref, Ref, watch} from 'vue'
import {ElButton, ElNotification} from 'element-plus'
import {Search} from '@element-plus/icons-vue'
import {Book, MetaBook} from "@/types/book";
import MetadataUpdate from "@/components/MetadataUpdate.vue";



const props = defineProps<{
  book: Book;
  dialogSearchVisible: boolean;
}>();

const isPhone = window.innerWidth < 768
const dialogUpdateVisible = ref(false)
const querySearchLoading = ref(false)
const selectRow = ref({})
const triggerUpdate = ref(false)
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

const emit = defineEmits(['update:dialogSearchVisible'])

const handleCurrentChange = (val: MetaBook) => {
  selectRow.value = val
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

const handleCloseSearch = () => {
  emit('update:dialogSearchVisible', false)
}

const nextUpdate = () => {
  emit('update:dialogSearchVisible', false)
  dialogUpdateVisible.value = true
  console.log(selectRow)
}

function initData(d: Book) {
  options.value = []
  console.log(d)
  if (d.isbn) {
    let isbn = d.isbn.replace(/-/g, '')
    query.value = isbn
    options.value.push({
      value: isbn,
      label: isbn
    })
    options.value.push({
      value: d.title,
      label: d.title
    })
  } else if (d.title) {
    query.value = d.title
    options.value.push({
      value: d.title,
      label: d.title
    })
  }
  if (d.authors) {
    options.value.push({
      value: d.title + ' ' + d.authors,
      label: d.title + ' ' + d.authors
    })
  }
  if (options.value.length > 0) {
    query.value = options.value[0].value
  }else {
    query.value = ''
  }
}


onMounted(() => {
  if (props.book) {
    initData(props.book)
  }
})

watch(() => props.book, (newVal) => {
  if (props.book) {
    initData(props.book)
  }
})

</script>

<style scoped></style>