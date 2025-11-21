<template>
  <div class="setting-wrapper glass-container">
    <h2 class="setting-title">设置</h2>
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
  </div>
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
          name: '全量重建索引',
          description: '后台重建MeiliSearch索引 (耗时较长)',
          loading: false,
          func: this.rebuildIndex,
          operator: '启动任务'
        },
        {
          name: '切换主备索引',
          description: '切换搜索使用的索引集合',
          loading: false,
          func: this.switchIndex,
          operator: '切换'
        },
        {
          name: '批量管理',
          description: '批量管理书籍元数据',
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
    async rebuildIndex(config: { loading: boolean }) {
      config.loading = true
      try {
        const response = await fetch('/api/tasks/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ type: 'meilisearch_sync', mode: 'full' })
        })
        config.loading = false
        
        if (response.ok) {
          const data = await response.json()
          if (data.code === 200) {
            ElNotification({
              title: '任务已启动',
              message: h('div', null, [
                h('span', null, '全量重建任务已在后台启动。请前往 '),
                h('a', { 
                  href: '/tasks', 
                  style: 'color: var(--el-color-primary); cursor: pointer; text-decoration: underline;',
                  onClick: (e: Event) => { 
                    e.preventDefault(); 
                    this.$router.push('/tasks') 
                  } 
                }, '任务页面'),
                h('span', null, ' 查看进度。')
              ]),
              type: 'success',
              duration: 5000
            })
          } else {
            ElNotification({
              title: '启动失败',
              message: data.message || '无法启动任务',
              type: 'error'
            })
          }
        } else {
          ElNotification({
            title: '请求失败',
            message: response.statusText,
            type: 'error'
          })
        }
      } catch (error) {
        config.loading = false
        ElNotification({
          title: '错误',
          message: (error as Error).message,
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

<style scoped>
.setting-wrapper {
  margin: var(--spacing-lg) auto;
  max-width: 1200px;
  /* min-height 已由全局 .main-content 处理 */
}

.setting-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: var(--spacing-lg);
}

:deep(.el-table) {
  background: transparent !important;
  color: var(--el-text-color-primary);
}

:deep(.el-table tr) {
  background: transparent !important;
}

:deep(.el-table th) {
  background: var(--glass-bg-light) !important;
  color: var(--el-text-color-primary) !important;
  border-bottom: 1px solid var(--glass-border);
}

:deep(.el-table td) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--el-text-color-regular);
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background: var(--glass-bg-light) !important;
}

:deep(.el-table__row:hover) {
  background: var(--glass-bg-medium) !important;
}
</style>
