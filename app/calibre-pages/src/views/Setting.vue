<template>
  <el-row :gutter="10">
    <h2 class="text-xl">设置</h2>
    <el-table :data="settings" style="width: 100%" stripe>
      <el-table-column prop="name" label="Setting"></el-table-column>
      <el-table-column prop="description" label="Value"></el-table-column>
      <el-table-column fixed="right" label="Action">
        <template #default="scope">
          <el-button
              v-loading="scope.row.loading"
              element-loading-background="rgba(122, 122, 122, 0.8)"
              type="primary"
              size="default"
              @click="scope.row.func(scope.row)"
          >
            {{ scope.row.operator }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-row>
</template>

<script lang="ts">
import {h} from 'vue'
import {ElButton, ElContainer, ElMain, ElNotification, ElRow, ElTable, ElTableColumn} from 'element-plus'

export default {
  name: 'Setting',
  components: {
    ElRow,
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
          name: '更新索引',
          description: '更新MeiliSearch索引',
          loading: false,
          func: this.updateIndex,
          operator: '执行'
        },
        {
          name: '切换主备索引',
          description: '切换主备索引',
          loading: false,
          func: this.switchIndex,
          operator: '执行'
        },
        {
          name: '批量管理',
          description: '批量管理书籍',
          loading: false,
          func: this.redirectToManagerPage,
          operator: '前往'
        }
      ]
    }
  },
  methods: {
    async switchIndex(config: { loading: boolean }) {
      config.loading = true
      try {
        const response = await fetch('/api/index/switch', {method: 'POST'})
        config.loading = false
        if (response.ok) {
          const responseData = await response.json()

          if (responseData.code === 200) {
            ElNotification({
              title: 'Index switched successfully.',
              message: h('i', {style: 'color: teal'}, 'Index switched successfully.'),
              type: 'success'
            })
          } else {
            ElNotification({
              title: 'Failed to update index.',
              message: h('i', {style: 'color: red'}, 'Error: ' + responseData.error),
              type: 'error'
            })
          }


        } else {
          ElNotification({
            title: 'Failed to update index.',
            message: h('i', {style: 'color: red'}, 'Error: ' + response.statusText),
            type: 'error'
          })
        }
      } catch (error) {
        config.loading = false
        ElNotification({
          title: 'Failed to update index.',
          message: h('i', {style: 'color: red'}, 'Error: ' + (error as Error).message),
          type: 'error'
        })
      }
    },
    async updateIndex(config: { loading: boolean }) {
      config.loading = true
      try {
        const response = await fetch('/api/index/update', {method: 'POST'})
        config.loading = false
        if (response.ok) {
          const responseData = await response.json()

          if (responseData.code === 200) {
            ElNotification({
              title: 'Index update successfully.',
              message: h('i', {style: 'color: teal'}, 'Index updated successfully.'),
              type: 'success'
            })
          } else {
            ElNotification({
              title: 'Failed to update index.',
              message: h('i', {style: 'color: red'}, 'Error: ' + responseData.error),
              type: 'error'
            })
          }
        } else {
          ElNotification({
            title: 'Failed to update index.',
            message: h('i', {style: 'color: red'}, 'Error: ' + response.statusText),
            type: 'error'
          })
        }
      } catch (error) {
        config.loading = false
        ElNotification({
          title: 'Failed to update index.',
          message: h('i', {style: 'color: red'}, 'Error: ' + (error as Error).message),
          type: 'error'
        })
      }
    },
    redirectToManagerPage(config: { loading: boolean }) {
      this.$router.push('/metadata/manager')
    },
  }
}
</script>

<style scoped></style>
