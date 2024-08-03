<template>
  <div class="affix-container">
  <el-affix target=".affix-container">
    <el-input v-model="searchQuery" @keyup.enter="redirectToSearch" placeholder="书名、作者、ISBN">
      <template #append>
        <el-button size="large" @click="redirectToSearch">搜索</el-button>
      </template>
    </el-input>
  </el-affix>
  </div>
</template>

<script>
import { ElButton, ElInput } from 'element-plus'

export default {
  name: 'SearchBar',
  components: { ElButton, ElInput },
  data() {
    return {
      searchQuery: ''
    }
  },
  methods: {
    redirectToSearch() {
      if (this.searchQuery) {
        this.$router.push({ path: '/search', query: { q: this.searchQuery } })
      }
    }
  }
}
</script>
<style scoped>

</style>
