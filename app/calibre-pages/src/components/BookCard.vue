<template>
  <el-card class="book-card" @click="redirectToDetail(book.id)">
    <el-row type="flex" align="middle">
      <el-col :span="6" class="cover-container">
        <img class="book-cover" :src="proxy_image?('/api/proxy/cover/' + book.cover) :book.cover" alt="book cover"/>
      </el-col>
      <el-col :span="18" class="info-container">
        <div class="info-item title">{{ truncateText(book.title) }}</div>
        <div class="info-item author" v-if="book.authors && book.authors.length">
          {{ truncateText(book.authors.join(', ')) }}
        </div>
        <div class="info-item publisher" v-if="more_info">{{ book.publisher }}</div>
        <div class="info-item publisher" v-if="more_info">{{ book.isbn }}</div>
        <div class="info-item pubdate" v-if="book.pubdate">
          {{ new Date(book.pubdate).toLocaleDateString() }}
        </div>
      </el-col>
    </el-row>
  </el-card>
</template>

<script setup lang="ts">
import {ElCard, ElCol, ElRow} from 'element-plus'
import {useRouter} from 'vue-router';


import {Book} from '@/types/book'

const props = defineProps<{
  book: Book;
  more_info: boolean;
  proxy_image: boolean;
}>();

const router = useRouter();

const redirectToDetail = (id: number) => {
  router.push(`/detail/${id}`);
};

const truncateText = (title: string) => {
  if (!title) return '';
  return title.length > 20 ? title.substring(0, 16) + '...' : title;
};
</script>
<style scoped>
.book-card {
  width: 100%;
  height: 150px; /* 固定卡片高度 */
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}

.cover-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-cover {
  width: 96px;
  height: 139px; /* 固定卡片高度 */
}

.info-container {
  height: 139px; /* 固定卡片高度 */
  width: 192px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.info-item {
  line-height: 1.5;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  margin-left: 20px;
}

.title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.author,
.publisher,
.pubdate {
  font-size: 14px;
  color: #666;
}

.el-card {
  cursor: pointer;
  transition: box-shadow 0.3s;
}

.el-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}
</style>
