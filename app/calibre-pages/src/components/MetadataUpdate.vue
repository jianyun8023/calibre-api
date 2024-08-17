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
          <template #suffix v-if="titleNew === '1' && newBook.sub_title">
            <el-checkbox v-model="useSubTitle">Full</el-checkbox>
          </template>
        </el-input>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="titleNew"
            aria-label="label position"
            placeholder="源"
            :disabled="!book.title || book.title === newBook.title"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="作者">
      <el-col :span="18">
        <el-checkbox-group v-model="form.authors">
          <el-checkbox v-for="author in authors" :label="author" :value="author"/>
        </el-checkbox-group>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="authorsNew"
            aria-label="label position"
            placeholder="源"
            :disabled="!book.authors || arraysEqual(book.authors, newBook.author)"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="出版社">
      <el-col :span="18">
        <el-input v-model="form.publisher" placeholder="请输入出版社"></el-input>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="publisherNew"
            aria-label="label position"
            placeholder="源"
            :disabled="!book.publisher || book.publisher === newBook.publisher"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="出版日期">
      <el-col :span="18">
        <el-date-picker v-model="form.pubdate" type="date" placeholder="请选择出版日期"/>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="pubdateNew"
            aria-label="label position"
            placeholder="源"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="ISBN">
      <el-col :span="18">
        <el-input v-model="form.isbn" placeholder="请输入ISBN"></el-input>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="isbnNew"
            aria-label="label position"
            placeholder="源"
            :disabled="!book.isbn || book.isbn === newBook.isbn13"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
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
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="ratingNew"
            aria-label="label position"
            placeholder="源"
            :disabled="book.rating === 0"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="标签">
      <el-col :span="18">
        <el-checkbox-group v-model="form.tags">
          <el-checkbox v-for="tag in tags" :label="tag" :value="tag"/>
        </el-checkbox-group>
      </el-col>
      <el-col :span="6">
        <el-radio-group
            class="align-right"
            v-model="tagsNew"
            aria-label="label position"
            placeholder="源"
            :disabled="!book.tags"
        >
          <el-radio-button value="1">新</el-radio-button>
          <el-radio-button value="2">旧</el-radio-button>
        </el-radio-group>
      </el-col>
    </el-form-item>
    <el-form-item label="简介">
      <el-radio-group
          class="radio-group"
          v-if="book.comments && book.comments !== newBook.summary"
          v-model="commentsNew"
          aria-label="label position"
          placeholder="源"
      >
        <el-radio-button value="1">新</el-radio-button>
        <el-radio-button value="2">旧</el-radio-button>
      </el-radio-group>
      <el-input :rows="6" v-model="form.comments" placeholder="请输入简介" type="textarea">
      </el-input>
    </el-form-item>
  </el-form>
</template>
<script setup lang="ts">
import {ElNotification} from 'element-plus'
import {h, reactive, ref, watch} from 'vue'
import {Book} from '@/types/book'

const props = defineProps<{
  book: Book;
  newBook: any;
  updateMetadataFlag: boolean;
}>();

const form = reactive({
  title: '',
  authors: [] as string[],
  publisher: '',
  pubdate: new Date(),
  isbn: '',
  comments: '',
  tags: [] as string[],
  rating: 0
});

const options = ref([]);
const tableData = ref([]);
const useSubTitle = ref(true);
const titleNew = ref('1');
const authorsNew = ref('1');
const authors = ref([] as string[]);
const publisherNew = ref('1');
const pubdateNew = ref('1');
const isbnNew = ref('1');
const commentsNew = ref('1');
const ratingNew = ref('1');
const tagsNew = ref('1');
const tags = ref([] as string[]);
const loading = ref(false);
const colors = ['#99A9BF', '#F7BA2A', '#F7BA2A', '#FF9900'];

const joinTitle = (row: any) => {
  if (row.sub_title) {
    return row.title + '：' + row.sub_title;
  } else {
    return row.title;
  }
};

const parseDateString = (dateString: string) => {
  const dateParts = dateString.split('-');
  const year = parseInt(dateParts[0], 10);
  const month = parseInt(dateParts[1], 10) - 1; // JavaScript months are 0-based
  const day = dateParts.length === 3 ? parseInt(dateParts[2], 10) : 1; // Default to the first day of the month if day is not provided
  return new Date(year, month, day);
};

const arraysEqual = (arr1: string[], arr2: string[]) => {
  if (!arr1) return false;
  if (!arr2) return false;
  if (arr1 === arr2) return true;
  if (arr1.length !== arr2.length) return false;
  for (let i = 0; i < arr1.length; i++) {
    if (arr1[i] !== arr2[i]) return false;
  }
  return true;
};

const updateMetadata = async () => {
  loading.value = true;
  console.log(form);

  try {
    const response = await fetch(`/api/book/${props.book.id}/update`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(form)
    });
    if (response.ok) {
      setTimeout(() => {
        ElNotification({
          title: '书籍更新成功',
          message: form.title,
          type: 'success'
        });
        loading.value = false;
        //刷新页面
        window.location.reload();
      }, 1000);
    } else {
      ElNotification({
        title: '书籍更新失败',
        message: h('i', {style: 'color: red'}, form.title),
        type: 'error'
      });
      loading.value = false;
      // this.updateMetadataFlag = false
    }
  } catch (e: any) {
    ElNotification({
      title: '书籍更新失败',
      message: h('i', {style: 'color: red'}, e.message),
      type: 'error'
    });
    loading.value = false;
    // this.updateMetadataFlag = false
  }
};

watch(() => props.updateMetadataFlag, (val) => {
  if (val) {
    updateMetadata();
  }
});

watch(() => useSubTitle, (val) => {
  if (titleNew.value === '1') {
    form.title = val ? joinTitle(props.newBook) : props.newBook.title;
  } else {
    form.title = props.book.title;
  }
});

watch(() => titleNew.value, (val) => {
  if (val === '1') {
    form.title = useSubTitle.value ? joinTitle(props.newBook) : props.newBook.title;
  } else {
    form.title = props.book.title;
  }
});

watch(() => authorsNew.value, (val) => {
  if (val === '1') {
    form.authors = props.newBook.author;
    authors.value = props.newBook.author;
  } else {
    form.authors = props.book.authors;
    authors.value = props.book.authors;
  }
});

watch(() => publisherNew.value, (val) => {
  if (val === '1') {
    form.publisher = props.newBook.publisher;
  } else {
    form.publisher = props.book.publisher;
  }
});

watch(() => pubdateNew.value, (val) => {
  if (val === '1') {
    form.pubdate = parseDateString(props.newBook.pubdate);
  } else {
    form.pubdate = new Date(props.book.pubdate);
  }
});

watch(() => isbnNew.value, (val) => {
  if (val === '1') {
    form.isbn = props.newBook.isbn13;
  } else {
    form.isbn = props.book.isbn;
  }
});

watch(() => commentsNew.value, (val) => {
  if (val === '1') {
    form.comments = props.newBook.summary;
  } else {
    form.comments = props.book.comments;
  }
});

watch(() => ratingNew.value, (val) => {
  if (val === '1') {
    form.rating = Number(props.newBook.rating.average);
  } else {
    form.rating = props.book.rating;
  }
});

watch(() => tagsNew.value, (val) => {
  if (val === '1') {
    tags.value = props.newBook.tags.map((tag: any) => tag.name);
    form.tags = tags.value;
  } else {
    form.tags = props.book.tags;
    tags.value = props.book.tags;
  }
});

function setFormData() {
  props.newBook.summary = props.newBook.summary?.replace(/class=".*?"/g, '');
  form.comments = props.newBook.summary;
  form.title = useSubTitle.value ? joinTitle(props.newBook) : props.newBook.title;
  form.publisher = props.newBook.publisher;
  form.isbn = props.newBook.isbn13;
  form.pubdate = props.newBook.pubdate ? parseDateString(props.newBook.pubdate) : new Date(0);
  form.authors = props.newBook.author;
  authors.value = props.newBook.author;
  tags.value = props.newBook.tags?.map((tag: any) => tag.name);
  form.tags = tags.value;
  form.rating = Number(props.newBook.rating?.average);
}

setFormData();

watch(() => props.newBook, () => {
  setFormData();
});

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
