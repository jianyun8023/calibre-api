<template>
  <div class="chat-container">
    <!-- 对话列表侧边栏 -->
    <el-aside width="250px" class="conversation-sidebar">
      <el-button @click="createNewConversation" type="primary" class="new-chat-btn" style="width: 100%; margin-bottom: 12px;">
        <el-icon><ChatDotRound /></el-icon>
        新建对话
      </el-button>
      
      <el-scrollbar class="conversation-list" height="calc(100vh - 120px)">
        <div 
          v-for="conv in conversations" 
          :key="conv.id"
          :class="['conversation-item', { active: currentConversation?.id === conv.id }]"
          @click="selectConversation(conv)"
        >
          <div class="conv-content">
            <div class="conv-title">{{ conv.title }}</div>
            <div class="conv-time">{{ formatTime(conv.updated_at) }}</div>
          </div>
          <el-button 
            v-if="currentConversation?.id === conv.id"
            class="delete-btn" 
            type="danger" 
            link 
            :icon="Delete"
            @click.stop="handleDeleteConversation(conv.id)"
          />
        </div>
      </el-scrollbar>
    </el-aside>
    
    <!-- 消息区域 -->
    <el-main class="chat-main">
      <div v-if="!currentConversation" class="empty-state">
        <el-icon :size="64" color="var(--el-text-color-secondary)"><ChatDotRound /></el-icon>
        <p>选择或创建一个对话开始聊天</p>
      </div>
      
      <template v-else>
        <el-scrollbar ref="messageScrollRef" class="message-list">
          <div v-for="msg in messages" :key="msg.id" :class="['message', msg.role]">
            <!-- 思考过程（可折叠） -->
            <el-collapse v-if="msg.thinking" class="thinking-section">
              <el-collapse-item title="💭 思考过程" name="thinking">
                <pre class="thinking-content">{{ msg.thinking }}</pre>
              </el-collapse-item>
            </el-collapse>
            
            <div class="message-wrapper">
              <!-- 消息内容 -->
              <div :key="msg.content.length" class="message-content markdown-body" v-html="msg.renderedContent || renderMarkdown(msg.content)"></div>
              
              <!-- 消息操作栏 -->
              <div class="message-actions">
                <el-button 
                  type="danger" 
                  link 
                  size="small" 
                  :icon="Delete"
                  @click="handleDeleteMessage(msg.id)"
                />
              </div>
            </div>
            
            <!-- 书籍卡片 -->
            <div v-if="extractBooks(msg).length > 0" class="book-cards">
              <el-card 
                v-for="book in visibleBooks(msg)" 
                :key="book.id" 
                class="book-card"
                shadow="hover"
                @click="goToBook(book.id)"
              >
                <img :src="`/api/get/cover/${book.id}.jpg`" class="book-cover" />
                <div class="book-info">
                  <h4>{{ book.title }}</h4>
                  <p class="authors">{{ book.authors?.join(', ') }}</p>
                  <el-tag v-if="book.reason" size="small" type="info" style="margin-bottom: 8px;">{{ book.reason }}</el-tag>
                  <el-button 
                    type="primary" 
                    size="small" 
                    link 
                    @click.stop="summarizeBook(book)"
                  >
                    总结此书
                  </el-button>
                </div>
              </el-card>
              
              <!-- 换一换按钮 -->
              <div v-if="extractBooks(msg).length > 8" class="refresh-books">
                <el-button link type="primary" @click="changeBooks(msg)">
                  <el-icon><Refresh /></el-icon> 换一换
                </el-button>
              </div>
            </div>
          </div>
          
          <!-- 加载中 -->
          <div v-if="sending" class="message assistant loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>AI 正在思考...</span>
            <el-button type="danger" link size="small" @click="stopGeneration">
              <el-icon><VideoPause /></el-icon> 停止生成
            </el-button>
          </div>
        </el-scrollbar>
        
        <!-- 输入框 -->
        <div class="input-area">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="3"
            placeholder="输入消息... (Shift+Enter 换行，Enter 发送)"
            @keydown.enter.exact.prevent="sendMessage"
            :disabled="sending"
          />
          <div class="input-actions">
            <el-button v-if="sending" @click="stopGeneration" type="danger" plain>
              <el-icon><VideoPause /></el-icon> 停止
            </el-button>
            <el-button v-else @click="sendMessage" type="primary" :disabled="!inputMessage.trim()">
              发送
            </el-button>
          </div>
        </div>
      </template>
    </el-main>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, triggerRef } from 'vue'
import { useRouter } from 'vue-router'
import { ChatDotRound, Loading, Refresh, Delete, VideoPause } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css' // 引入代码高亮样式

import type { Conversation, Message } from '@/api/chat'
import { createConversation, listConversations, getMessages, deleteConversation, deleteMessage } from '@/api/chat'

const router = useRouter()
const conversations = ref<Conversation[]>([])
const currentConversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])
const inputMessage = ref('')
const sending = ref(false)
const messageScrollRef = ref()
const bookPageMap = ref<Record<string, number>>({})
let abortController: AbortController | null = null

// 初始化 Markdown 解析器
const md = new MarkdownIt({
  html: true,
  breaks: true,
  linkify: true,
  typographer: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre class="hljs"><code>' +
               hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
               '</code></pre>';
      } catch (__) {}
    }
    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>';
  }
})

onMounted(async () => {
  await loadConversations()
})

const loadConversations = async () => {
  try {
    conversations.value = await listConversations()
    if (conversations.value.length > 0 && !currentConversation.value) {
      await selectConversation(conversations.value[0])
    }
  } catch (error) {
    ElMessage.error('加载对话列表失败')
  }
}

const createNewConversation = async () => {
  try {
    const conv = await createConversation({ title: '新对话' })
    conversations.value.unshift(conv)
    await selectConversation(conv)
  } catch (error) {
    console.error('Create conversation failed:', error)
    ElMessage.error(`创建对话失败: ${error instanceof Error ? error.message : String(error)}`)
  }
}

const selectConversation = async (conv: Conversation) => {
  currentConversation.value = conv
  try {
    const msgs = await getMessages(conv.id)
    messages.value = msgs || []
    scrollToBottom()
  } catch (error) {
    console.error('Load messages failed:', error)
    ElMessage.error(`加载消息失败: ${error instanceof Error ? error.message : String(error)}`)
    messages.value = []
  }
}

const handleDeleteConversation = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个对话吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteConversation(id)
    conversations.value = conversations.value.filter(c => c.id !== id)
    if (currentConversation.value?.id === id) {
      currentConversation.value = null
      messages.value = []
      if (conversations.value.length > 0) {
        selectConversation(conversations.value[0])
      }
    }
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleDeleteMessage = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这条消息吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteMessage(id)
    messages.value = messages.value.filter(m => m.id !== id)
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const summarizeBook = (book: any) => {
  inputMessage.value = `总结一下《${book.title}》这本书的内容`
  sendMessage()
}

const stopGeneration = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
    sending.value = false
    ElMessage.info('已停止生成')
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || sending.value || !currentConversation.value) return
  
  const userMessage: Message = {
    id: Date.now().toString(),
    conversation_id: currentConversation.value.id,
    role: 'user',
    content: inputMessage.value,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)
  
  const input = inputMessage.value
  inputMessage.value = ''
  sending.value = true
  scrollToBottom()
  
  // 初始化 AbortController
  abortController = new AbortController()
  
  // AI 消息（流式更新）
  const aiMessage: Message = {
    id: (Date.now() + 1).toString(),
    conversation_id: currentConversation.value.id,
    role: 'assistant',
    content: '',
    created_at: new Date().toISOString(),
    renderedContent: ''  // 初始化为空字符串，确保 Vue 响应式追踪
  }
  messages.value.push(aiMessage)
  // 获取响应式对象
  const reactiveAiMessage = messages.value[messages.value.length - 1]
  
  try {
    // 使用 SSE 接收流式响应
    const response = await fetch(`/api/chat/conversations/${currentConversation.value.id}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: input }),
      signal: abortController.signal
    })
    
    if (!response.ok) throw new Error('发送失败')
    
    const reader = response.body!.getReader()
    const decoder = new TextDecoder()
    
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      
      const chunk = decoder.decode(value)
      const lines = chunk.split('\n')
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.substring(6)
          if (data === '[DONE]') break
          
          // 尝试解析为 JSON (处理元数据)
          if (data.startsWith('{')) {
            try {
              const jsonData = JSON.parse(data)
              if (jsonData.type === 'metadata') {
                reactiveAiMessage.metadata = JSON.stringify({ books: jsonData.books })
                continue
              }
            } catch (e) {
              // 忽略解析错误，当作普通文本
            }
          }
          
          // 追加内容并立即渲染
          reactiveAiMessage.content += data
          reactiveAiMessage.renderedContent = md.render(reactiveAiMessage.content)
          scrollToBottom()
        }
      }
    }
  } catch (error: any) {
    if (error.name === 'AbortError') {
      console.log('Generation aborted')
    } else {
      ElMessage.error('发送消息失败')
      messages.value.pop() // 移除失败的 AI 消息
    }
  } finally {
    sending.value = false
    abortController = null
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messageScrollRef.value) {
      const scrollbar = messageScrollRef.value
      scrollbar.setScrollTop(scrollbar.wrapRef.scrollHeight)
    }
  })
}

const renderMarkdown = (text: string) => {
  return md.render(text || '')
}

const extractBooks = (message: Message) => {
  if (!message.metadata) return []
  try {
    const meta = JSON.parse(message.metadata)
    return meta.books || []
  } catch {
    return []
  }
}

const visibleBooks = (message: Message) => {
  const allBooks = extractBooks(message)
  const page = bookPageMap.value[message.id] || 0
  const start = page * 8
  return allBooks.slice(start, start + 8)
}

const changeBooks = (message: Message) => {
  const allBooks = extractBooks(message)
  const page = bookPageMap.value[message.id] || 0
  const nextStart = (page + 1) * 8
  
  if (nextStart < allBooks.length) {
    bookPageMap.value[message.id] = page + 1
  } else {
    bookPageMap.value[message.id] = 0
  }
}

const goToBook = (id: number) => {
  router.push(`/detail/${id}`)
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return date.toLocaleDateString()
}
</script>

<style scoped>
.chat-container {
  display: flex;
  height: 100%;
  background: var(--el-bg-color);
  overflow: hidden; /* 防止整体滚动 */
}

.conversation-sidebar {
  background: var(--glass-bg);
  border-right: 1px solid var(--el-border-color);
  padding: 16px;
  flex-shrink: 0;
}

.conversation-item {
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: var(--el-bg-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.conversation-item:hover {
  background: var(--glass-bg-medium);
}

.conversation-item.active {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.conv-content {
  flex: 1;
  overflow: hidden;
}

.conv-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.conversation-item:hover .delete-btn {
  opacity: 1;
}

.chat-main {
  display: flex;
  flex-direction: column;
  padding: 0;
  flex: 1;
  height: 100%;
  overflow: hidden; /* 关键：确保内部滚动 */
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--el-text-color-secondary);
}

.message-list {
  flex: 1;
  padding: 20px;
  overflow-y: auto; /* 允许滚动 */
}

.message {
  margin-bottom: 20px;
  max-width: 80%;
}

.message.user {
  margin-left: auto;
}

.message-wrapper {
  position: relative;
  group: true; /* For hover effect */
}

.message.user .message-content {
  background: var(--el-color-primary);
  color: white;
  padding: 12px 16px;
  border-radius: 12px 12px 0 12px;
}

.message.assistant .message-content {
  background: var(--glass-bg);
  padding: 12px 16px;
  border-radius: 12px 12px 12px 0;
}

.message-actions {
  position: absolute;
  top: 0;
  right: -30px;
  opacity: 0;
  transition: opacity 0.2s;
}

.message:hover .message-actions {
  opacity: 1;
}

.message.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
}

.thinking-section {
  margin-bottom: 12px;
  background: var(--glass-bg-medium);
  border-radius: 8px;
}

.thinking-content {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  margin: 0;
}

.book-cards {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.book-card {
  width: 200px;
  cursor: pointer;
  transition: transform 0.2s;
}

.book-card:hover {
  transform: translateY(-4px);
}

.book-cover {
  width: 100%;
  height: 260px;
  object-fit: cover;
  border-radius: 4px;
}

.book-info h4 {
  font-size: 14px;
  margin: 8px 0 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.authors {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin: 0 0 8px 0;
}

.refresh-books {
  width: 100%;
  display: flex;
  justify-content: center;
  margin-top: 8px;
}

.input-area {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
  flex-shrink: 0; /* 防止被压缩 */
  align-items: flex-end;
}

.input-area .el-textarea {
  flex: 1;
}

.input-actions {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
}

/* Markdown Styles */
:deep(.markdown-body) {
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
}

:deep(.markdown-body pre) {
  background: #1e1e1e;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 8px 0;
}

:deep(.markdown-body code) {
  font-family: 'Menlo', 'Monaco', monospace;
  font-size: 12px;
}

:deep(.markdown-body p) {
  margin-bottom: 8px;
}

:deep(.markdown-body ul), :deep(.markdown-body ol) {
  padding-left: 20px;
  margin-bottom: 8px;
}
</style>
