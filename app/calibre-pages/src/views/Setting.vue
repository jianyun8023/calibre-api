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
