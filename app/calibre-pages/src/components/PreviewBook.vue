<script setup lang="ts">

import {ref, watch} from 'vue';
import {ElButton, ElNotification} from 'element-plus';
import {Book} from '@/types/book';

const props = defineProps<{
  book: Book;
  dialogPreviewVisible: boolean;
}>();

const emit = defineEmits(['dialogPreviewVisible']);
const isPhone = window.innerWidth < 768;

const dialogContentVisible = ref(false)
const bookMenu = ref({} as any)
const menuLoding = ref(false)
const previewLoding = ref(false)
const currentPreviewTitle = ref('')
const currentPreviewContent = ref('')
const currentPreviewUrl = ref('')

const defaultProps = {
  children: 'points',
  label: 'text'
}

watch(() => props.book, () => {
  showBookMenu();
});


const showBookMenu = async () => {

  menuLoding.value = true
  try {
    const response = await fetch(`/api/read/${props.book.id}/toc`)
    if (!response.ok) throw new Error('Network response was not ok')
    const data = await response.json()
    bookMenu.value = data.points
    menuLoding.value = false
    if (!data.points) {
      ElNotification({
        title: '目录加载失败',
        message: '无法加载目录',
        type: 'warning'
      })
    }
    console.log(data.points)
  } catch (error) {
    menuLoding.value = false
    console.error('There was a problem with the fetch operation:', error)
  }
}


const handleNodeClick = async (data: any) => {
  console.log(data)

  // this.previewLoding = true
  dialogContentVisible.value = true
  currentPreviewUrl.value = "/api" + data.content.src
  currentPreviewTitle.value = data.text
  // fetch("/api" + data.content.src)
  //     .then(response => response.text())
  //     .then(data => {
  //       this.currentPreviewContent = data
  //       this.previewLoding = false
  //     })
  //     .catch(error => {
  //       this.previewLoding = false
  //       ElMessage.error('There was a problem with the fetch operation:' + error.message)
  //     })

}

</script>

<template>
  <el-dialog
      :model-value="dialogPreviewVisible"
      @update:model-value="val => emit('dialogPreviewVisible', val)"
      title="查看目录"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <el-row class="margin-top" v-loading="menuLoding">
      <el-scrollbar height="600px">
        <el-tree
            style="max-width: 600px; width: 100%"
            :data="bookMenu"
            :props="defaultProps"
            @node-click="handleNodeClick"
        />
      </el-scrollbar>
    </el-row>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('dialogPreviewVisible', false)">OK</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
      v-model="dialogContentVisible"
      :title="'预览： ' + currentPreviewTitle"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :width="isPhone ? '100%' : '50%'"
  >
    <el-row class="margin-top" v-loading="previewLoding">
      <iframe :src="currentPreviewUrl" width="100%" height="600px">

      </iframe>
      <el-text v-html="currentPreviewContent"></el-text>
    </el-row>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogContentVisible = false">OK</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>

</style>