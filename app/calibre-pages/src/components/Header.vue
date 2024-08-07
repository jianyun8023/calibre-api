<template>
  <el-row justify="start">
    <el-link href="/">
      <el-text class="site-title">书海拾贝</el-text>
    </el-link>
    <!--    <h2 style="font-size: xx-large;">Treasure in Book Ocean</h2>-->
  </el-row>
</template>

<script lang="ts">
export default {
  name: 'SiteHeader'
}
</script>

<style scoped>
.site-title {
  font-size: xx-large;
  color: #ef475d;
  font-family: Inter, 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}
</style>
