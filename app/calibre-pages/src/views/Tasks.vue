<template>
  <el-container class="tasks-container">
    <el-header class="page-header">
      <div class="header-content">
        <h2 class="page-title">任务管理</h2>
        <p class="page-subtitle">管理后台同步任务与索引状态</p>
      </div>
    </el-header>

    <el-main>
      <el-row :gutter="20">
        <!-- Task Controls -->
        <el-col :span="24" :lg="8">
          <el-card class="control-card glass-card">
            <template #header>
              <div class="card-header">
                <span>Qdrant 数据同步</span>
                <el-tag size="small" type="success" effect="dark">语义搜索</el-tag>
              </div>
            </template>
            <div class="control-actions">
              <el-button type="primary" @click="startTask('qdrant_sync', 'incremental')" :loading="loading.qdrant_incremental">
                增量同步
              </el-button>
              <el-button type="warning" @click="startTask('qdrant_sync', 'full')" :loading="loading.qdrant_full">
                全量重建
              </el-button>
            </div>
            <p class="control-desc">
              同步书籍元数据并生成嵌入向量存入 Qdrant。用于支持自然语言语义搜索和相关书籍推荐。
            </p>
          </el-card>
        </el-col>


        <!-- Task List -->
        <el-col :span="24" :lg="16">
          <el-card class="list-card glass-card">
            <template #header>
              <div class="card-header">
                <span>任务列表</span>
                <el-button circle size="small" @click="fetchTasks">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </div>
            </template>
            
            <el-table :data="tasks" style="width: 100%" v-loading="loading.list">
              <el-table-column prop="type" label="类型" width="140">
                <template #default="scope">
                  <el-tag v-if="scope.row.type === 'qdrant_sync'" type="success">Qdrant Sync</el-tag>
                  <el-tag v-else type="info">{{ scope.row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="mode" label="模式" width="100">
                <template #default="scope">
                  <el-tag size="small" :type="scope.row.mode === 'full' ? 'warning' : 'info'">
                    {{ scope.row.mode === 'full' ? '全量' : '增量' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="state" label="状态" width="100">
                <template #default="scope">
                  <el-tag v-if="scope.row.state === 'running'" type="primary" effect="dark">运行中</el-tag>
                  <el-tag v-else-if="scope.row.state === 'completed'" type="success" effect="plain">完成</el-tag>
                  <el-tag v-else-if="scope.row.state === 'error'" type="danger" effect="plain">错误</el-tag>
                  <el-tag v-else type="info" effect="plain">{{ scope.row.state }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="进度" min-width="150">
                <template #default="scope">
                  <div class="progress-wrapper">
                    <el-progress 
                      :percentage="Math.round(scope.row.progress || 0)" 
                      :status="scope.row.state === 'error' ? 'exception' : (scope.row.state === 'completed' ? 'success' : '')"
                    />
                    <span class="progress-msg">{{ scope.row.message }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="80" fixed="right">
                <template #default="scope">
                  <el-button 
                    v-if="scope.row.state === 'running'" 
                    type="danger" 
                    link 
                    size="small" 
                    @click="stopTask(scope.row.id)"
                  >
                    停止
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface Task {
  id: string
  type: string
  mode: string
  state: string
  progress: number
  message: string
  target_index?: string
  start_time: string
  end_time: string
  error?: string
}

const tasks = ref<Task[]>([])
const loading = ref({
  list: false,
  qdrant_full: false,
  qdrant_incremental: false
})

let pollTimer: number | null = null

const fetchTasks = async () => {
  try {
    const res = await fetch('/api/tasks')
    const data = await res.json()
    if (data.code === 200) {
      tasks.value = data.data
    }
  } catch (e) {
    console.error(e)
  }
}

const startTask = async (type: string, mode: string) => {
  const key = `qdrant_${mode}` as keyof typeof loading.value
  loading.value[key] = true

  
  try {
    const res = await fetch('/api/tasks/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type, mode })
    })
    const data = await res.json()
    
    if (data.code === 200) {
      ElMessage.success('任务已启动')
      fetchTasks()
    } else {
      ElMessage.error(data.message || '启动失败')
    }
  } catch (e) {
    ElMessage.error('请求失败')
  } finally {
    loading.value[key] = false
  }
}

const stopTask = async (id: string) => {
  try {
    const res = await fetch(`/api/tasks/${id}/stop`, { method: 'POST' })
    const data = await res.json()
    if (data.code === 200) {
      ElMessage.success('已发送停止请求')
      fetchTasks()
    } else {
      ElMessage.error(data.message || '停止失败')
    }
  } catch (e) {
    ElMessage.error('请求失败')
  }
}

onMounted(() => {
  fetchTasks()
  pollTimer = setInterval(fetchTasks, 2000) as unknown as number
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped lang="scss">
.tasks-container {
  height: 100%;
  background-color: var(--bg-color);
}

.page-header {
  background: var(--glass-bg);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--glass-border);
  padding: 0 var(--spacing-xl);
  height: 80px;
  display: flex;
  align-items: center;

  .page-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .page-subtitle {
    margin: var(--spacing-xs) 0 0;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }
}

.glass-card {
  background: var(--glass-bg);
  backdrop-filter: blur(10px);
  border: 1px solid var(--glass-border);
  border-radius: var(--border-radius-lg);
  
  :deep(.el-card__header) {
    border-bottom: 1px solid var(--glass-border);
    padding: var(--spacing-md) var(--spacing-lg);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: var(--text-primary);
}

.control-actions {
  display: flex;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.control-desc {
  color: var(--text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
  margin: 0;
}

.progress-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
  
  .progress-msg {
    font-size: 0.75rem;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.mono-text {
  font-family: monospace;
  font-size: 0.85em;
  background: rgba(0, 0, 0, 0.2);
  padding: 2px 4px;
  border-radius: 4px;
}
</style>
