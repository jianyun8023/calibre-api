<template>
  <el-container>
    <el-main>
      <h1>设置</h1>
      <el-container class="mt-8 w-full md:w-2/3">
        <h2 class="text-xl font-bold mb-4">Settings</h2>
        <el-table :data="settings" style="width: 100%">
          <el-table-column prop="name" label="Setting" width="180"></el-table-column>
          <el-table-column prop="description" label="Value"></el-table-column>
          <el-table-column label="Action" width="180">
            <el-button type="primary" @click="updateIndex">Update Index</el-button>
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
          description: 'Click to update the search index'
        }
      ]
    }
  },
  methods: {
    async updateIndex() {
      try {
        const response = await fetch('/api/index/update', {method: 'POST'})
        if (response.ok) {
          ElNotification({
            title: 'Index updated successfully.',
            message: h('i', {style: 'color: teal'}, '共计' + response.data + '本书'),
            type: 'success',
          })
        } else {
          ElNotification({
            title: 'Failed to update index.',
            message: h('i', {style: 'color: red'}, response.json().data),
            type: 'error',
          })
        }
      } catch (error) {
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
