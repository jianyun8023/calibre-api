<template>

  <el-form v-loading="loading" :model="form" label-width="120px" class="book-form"  style="max-width: 600px">
    <el-form-item label="书名" >
      <el-col :span="18">
        <el-input v-model="form.title" placeholder="请输入书名">
        </el-input>
      </el-col>
      <el-col :span="6" >
      <el-radio-group class="align-right" v-model="titleNew" aria-label="label position" placeholder="源" :disabled="!book.title || book.title === newBook.title">
        <el-radio-button value="1">新</el-radio-button>
        <el-radio-button value="2">旧</el-radio-button>
      </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="作者">
      <el-col :span="18">
        <el-input v-model="form.authors" placeholder="请输入作者">
        </el-input>
      </el-col>
      <el-col :span="6" >
      <el-radio-group class="align-right" v-model="authorsNew" aria-label="label position" placeholder="源"  :disabled="!book.authors || arraysEqual(book.authors,newBook.author)">
        <el-radio-button value="1">新</el-radio-button>
        <el-radio-button value="2">旧</el-radio-button>
      </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="出版社">

      <el-col :span="18">
        <el-input v-model="form.publisher" placeholder="请输入出版社">
        </el-input>
      </el-col>
      <el-col :span="6" >
        <el-radio-group class="align-right" v-model="publisherNew" aria-label="label position" placeholder="源" :disabled="!book.publisher || book.publisher === newBook.publisher">
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="出版日期">
      <el-col :span="18">
        <el-date-picker
            v-model="form.pubdate"
            type="date"
            placeholder="请选择出版日期"
        />
      </el-col>
      <el-col :span="6" >
      <el-radio-group class="align-right" v-model="pubdateNew" aria-label="label position" placeholder="源" >
        <el-radio-button value="1">新</el-radio-button>
        <el-radio-button value="2">旧</el-radio-button>
      </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="ISBN">

      <el-col :span="18">
        <el-input v-model="form.isbn" placeholder="请输入ISBN">
        </el-input>
      </el-col>
      <el-col :span="6" >
        <el-radio-group class="align-right" v-model="isbnNew" aria-label="label position" placeholder="源" :disabled="!book.isbn || book.isbn === newBook.isbn13">
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="简介">
      <el-radio-group class="radio-group" v-if="book.comments && book.comments !== newBook.summary" v-model="commentsNew"
                      aria-label="label position" placeholder="源">
        <el-radio-button value="1">新</el-radio-button>
        <el-radio-button value="2">旧</el-radio-button>
      </el-radio-group>
      <el-input :rows="6" v-model="form.comments" placeholder="请输入简介" type="textarea">
      </el-input>
    </el-form-item>


  </el-form>
</template>
<script>
import {ElButton, ElInput, ElNotification} from 'element-plus'
import {h} from "vue";

export default {
  name: 'MetadataEdit',
  components: {ElButton, ElInput},
  props: {
    book: {
      type: Object,
      default: () => ({})
    },
    newBook: {
      type: Object,
      default: () => ({})
    },
    updateMetadataFlag: {
      type: Boolean,
      default: false
    },
  },
  data() {
    return {
      form: {
        title: '',
        authors: [],
        publisher: '',
        pubdate: new Date(),
        isbn: '',
        comments: '',
      },
      options: [],
      tableData: [],
      titleNew: '1',
      authorsNew: '1',
      publisherNew: '1',
      pubdateNew: '1',
      isbnNew: '1',
      commentsNew: '1',
      loading: false,
    }
  },
  created() {

    // this.newBook.summary
    // 清理html标签中的class
    this.newBook.summary = this.newBook.summary.replace(/class=".*?"/g, '')
    this.form.comments = this.newBook.summary
    this.form.title = this.newBook.title
    this.form.publisher = this.newBook.publisher
    this.form.isbn = this.newBook.isbn13

    //parse date string to Date object
    this.form.pubdate = new Date(this.newBook.pubdate)
    this.form.authors = this.newBook.author
  },
  watch: {
    titleNew(val) {
      if (val === '1') {
        this.form.title = this.newBook.title
      } else {
        this.form.title = this.book.title
      }
    },
    authorsNew(val) {
      if (val === '1') {
        this.form.authors = this.newBook.author
      } else {
        this.form.authors = this.book.authors
      }
    },
    publisherNew(val) {
      if (val === '1') {
        this.form.publisher = this.newBook.publisher
      } else {
        this.form.publisher = this.book.publisher
      }
    },
    pubdateNew(val) {
      if (val === '1') {
        // 2022-4-1
        console.log("新" + this.newBook.pubdate)
        console.log("新" + new Date(this.newBook.pubdate))
        this.form.pubdate = new Date(this.newBook.pubdate)
      } else {
        console.log("旧" + this.book.pubdate)
        console.log("旧" + new Date(this.book.pubdate))
        this.form.pubdate = new Date(this.book.pubdate)
      }
    },
    isbnNew(val) {
      if (val === '1') {
        this.form.isbn = this.newBook.isbn13
      } else {
        this.form.isbn = this.book.isbn
      }
    },
    commentsNew(val) {
      if (val === '1') {
        this.form.comments = this.newBook.summary
      } else {
        this.form.comments = this.book.comments
      }
    },
    updateMetadataFlag(val) {
      if (val) {
        this.updateMetadata()
      }
    }
  },
  methods: {
    async updateMetadata() {
      this.loading = true
      console.log(this.form)

      try {
        const response = await fetch(`/api/book/${this.book.id}/update`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.form)
        })
        if (response.ok) {

          setTimeout(() => {
            ElNotification({
              title: '书籍更新成功',
              message: this.form.title,
              type: 'success',
            })
            this.loading = false
            //刷新页面
            window.location.reload()
          }, 1000)
        } else {
          ElNotification({
            title: '书籍更新失败',
            message: h('i', {style: 'color: red'}, this.form.title),
            type: 'error',
          })
          this.loading = false
        }
      } catch (e) {
        ElNotification({
          title: '书籍更新失败',
          message: h('i', {style: 'color: red'}, e.message),
          type: 'error',
        })
        this.loading = false
      }


    },
    arraysEqual(arr1, arr2) {
      if (arr1.length !== arr2.length) return false;
      for (let i = 0; i < arr1.length; i++) {
        if (arr1[i] !== arr2[i]) return false;
      }
      return true;
    },
  }
}

</script>

<style scoped>
.book-form {
  max-width: 600px;
  margin: auto;
  padding: 10px;
}
.radio-group {
  display: flex;
  justify-content: space-between;
  margin-top: 0;
}
.el-input, .el-date-picker {
  width: 100%;
}


.el-button {
  margin-right: 10px;
}
.align-right {
  display: flex;
  justify-content: flex-end;
}
</style>