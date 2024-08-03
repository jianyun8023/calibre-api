<template>
  <el-menu default-active="1" @select="handleSelect" :router="true">
    <el-menu-item index="/">
      <el-icon><House /></el-icon>
      <span>Home</span>
    </el-menu-item>
    <el-menu-item index="/books">
      <el-icon><Files /></el-icon>
      <span>Books</span>
    </el-menu-item>
    <el-menu-item index="/search">
      <el-icon><Search /></el-icon>
      <span>Search</span>
    </el-menu-item>
    <el-menu-item index="/setting">
      <el-icon><Setting /></el-icon>
      <span>Setting</span>
    </el-menu-item>
  </el-menu>
</template>

<script>
export default {
  name: 'Sidebar',
  methods: {
    handleSelect(key, keyPath) {
      console.log(key, keyPath)
    }
  }
}
</script>

<style scoped></style>
