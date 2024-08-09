<template>
  <el-row :gutter="20" justify="space-between" style='height: 100vh'>
    <vue-reader :url="bookUrl"
                :epubOptions='{
      allowPopups: true, // Adds `allow-popups` to sandbox-attribute
      allowScriptedContent: true, // Adds `allow-scripts` to sandbox-attribute
  }'
    />
  </el-row>
</template>
<script lang="ts">
import {VueReader} from 'vue-reader'
import {ElRow} from "element-plus";

export default {
  name: 'ReadBook',
  components: {ElRow, VueReader},
  data() {
    return {
      bookId: '',
      bookUrl: '',
    }
  },
  created() {
    this.bookId = (this.$route as any).params.id
    this.bookUrl = `/api/get/book/${this.bookId}.epub`
  },
}
</script>
<style scoped>

</style>