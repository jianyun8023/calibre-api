<template>
  <el-container>
    <el-main>
      <h2 class="text-xl font-bold mb-4">设置</h2>
      <el-container class="mt-8 w-full">
        <el-table :data="settings" style="width: 100%">
          <el-table-column prop="name" label="Setting" width="180"></el-table-column>
          <el-table-column prop="description" label="Value"></el-table-column>
          <el-table-column label="Action" width="180">
            <template #default="scope">
              <el-button v-loading="scope.row.loading"
                         element-loading-background="rgba(122, 122, 122, 0.8)"
                         type="primary"
                         size="large"
                         @click="updateIndex(scope.row)">
                Update Index
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-container>
    </el-main>
  </el-container>
</template>

<script>
import {h} from 'vue'

import {ElButton, ElContainer, ElMain, ElNotification, ElTable, ElTableColumn} from 'element-plus'

export default {
  name: 'Setting',
  components: {
    ElContainer,
    ElMain,
    ElTable,
    ElTableColumn,
    ElButton
  },
  data() {
    return {
      settings: [
        {
          name: 'Update Index',
          description: 'Click to update the search index',
          loading: false,
        }
      ],
    }
  },
  methods: {
    async updateIndex(config) {
      config.loading = true
      try {
        const response = await fetch('/api/index/update', {method: 'POST'})
        config.loading = false
        if (response.ok) {
          const responseData = await response.json();
          ElNotification({
            title: 'Index updated successfully.',
            message: h('i', {style: 'color: teal'}, '共计' + responseData.data + '本书'),
            type: 'success',
          })
        } else {
          ElNotification({
            title: 'Failed to update index.',
            message: h('i', {style: 'color: red'}, 'Error: ' + response.statusText),
            type: 'error',
          })
        }
      } catch (error) {
        config.loading = false
        ElNotification({
          title: 'Failed to update index.',
          message: h('i', {style: 'color: red'}, 'Error: ' + error.message),
          type: 'error',
        })
      }
    }
  }
}
</script>

<style scoped></style>
