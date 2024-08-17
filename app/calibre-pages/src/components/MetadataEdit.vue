<template>
  <el-form
      v-loading="loading"
      :model="form"
      label-width="70px"
      class="book-form"
      style="max-width: 600px"
  >
    <el-form-item label="书名">
      <el-col :span="18">
        <el-input v-model="form.title" placeholder="请输入书名">
        </el-input>
      </el-col>
    </el-form-item>
    <el-form-item label="作者">
      <el-col :span="18">
        <el-select
            v-model="form.authors"
            multiple
            filterable
            placeholder="Select"
            allow-create
            style="width: 240px"
        >
          <el-option
              v-for="item in book.authors"
              :key="item"
              :label="item"
              :value="item"
          />
        </el-select>
      </el-col>
    </el-form-item>
    <el-form-item label="出版社">
      <el-col :span="18">
        <el-input v-model="form.publisher" placeholder="请输入出版社"></el-input>
      </el-col>

    </el-form-item>
    <el-form-item label="出版日期">
      <el-col :span="18">
        <el-date-picker v-model="form.pubdate" type="date" placeholder="请选择出版日期"/>
      </el-col>

    </el-form-item>
    <el-form-item label="ISBN">
      <el-col :span="18">
        <el-input v-model="form.isbn" placeholder="请输入ISBN"></el-input>
      </el-col>

    </el-form-item>
    <el-form-item label="评分">
      <el-col :span="18">
        <el-rate
            :value="form.rating / 2"
            @input="(val: number) => (form.rating = val * 2)"
            show-score
            text-color="#ff9900"
            :max="5"
            allow-half
            :score-template="`${form.rating}分`"
        >
        </el-rate>
      </el-col>
    </el-form-item>
    <el-form-item label="标签">
      <el-col :span="18">
        <el-select
            v-model="form.tags"
            multiple
            filterable
            placeholder="Select"
            allow-create
            style="width: 240px"
        >
          <el-option
              v-for="item in book.tags"
              :key="item"
              :label="item"
              :value="item"
          />
        </el-select>
      </el-col>

    </el-form-item>
    <el-form-item label="简介">
      <el-input :rows="6" v-model="form.comments" placeholder="请输入简介" type="textarea">
      </el-input>
    </el-form-item>
    <el-form-item class="align-right">
      <el-button type="info" @click="updateMetadata" :loading="loading">更新</el-button>
    </el-form-item>
  </el-form>
</template>
<script setup lang="ts">
import {h, reactive, ref, watch} from 'vue';
import {ElButton, ElInput, ElNotification} from 'element-plus';
import {Book} from '@/types/book';

const props = defineProps<{
  book: Book;
}>();

const emit = defineEmits(['update:book']);


const form = reactive({
  title: props.book.title,
  authors: props.book.authors,
  publisher: props.book.publisher,
  pubdate: new Date(props.book.pubdate).toISOString().split('T')[0],
  isbn: props.book.isbn,
  comments: props.book.comments,
  tags: props.book.tags,
  rating: props.book.rating,
});

watch(() => props.book, (newVal) => {
  console.log(newVal)
  form.title = newVal.title;
  form.authors = newVal.authors;
  form.publisher = newVal.publisher;
  form.pubdate = newVal.pubdate;
  form.isbn = newVal.isbn;
  form.comments = newVal.comments;
  form.tags = newVal.tags;
  form.rating = newVal.rating;
  console.log(form)
});

const loading = ref(false);

const updateMetadata = async () => {
  loading.value = true;
  console.log(form);

  try {
    const response = await fetch(`/api/book/${props.book.id}/update`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(form),
    });
    if (response.ok) {
      setTimeout(() => {
        ElNotification({
          title: '书籍更新成功',
          message: form.title,
          type: 'success',
        });
        loading.value = false;
        window.location.reload();
      }, 1000);
    } else {
      ElNotification({
        title: '书籍更新失败',
        message: h('i', {style: 'color: red'}, form.title),
        type: 'error',
      });
      loading.value = false;
    }
  } catch (e: any) {
    ElNotification({
      title: '书籍更新失败',
      message: h('i', {style: 'color: red'}, e.message),
      type: 'error',
    });
    loading.value = false;
  }
};
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

.el-input,
.el-date-picker {
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
