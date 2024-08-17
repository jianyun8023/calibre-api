<template>
  <div class="affix-container">
    <el-affix target=".affix-container">
      <el-input
          v-model="searchQuery"
          @keyup.enter="redirectToSearch"
          placeholder="书名、作者、ISBN"
      >
        <template #append>
          <el-button size="large" @click="redirectToSearch">搜索</el-button>
        </template>
      </el-input>
    </el-affix>
  </div>
</template>

<script setup lang="ts">
import {ElButton, ElInput} from 'element-plus'
import {ref} from 'vue'
import {useRouter} from 'vue-router'

const searchQuery = ref('')
const router = useRouter()

const redirectToSearch = () => {
  if (searchQuery.value) {
    router.push({path: '/search', query: {q: searchQuery.value}})
  }
}
</script>
<style scoped></style>
