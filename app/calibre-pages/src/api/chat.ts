export interface Conversation {
    id: string
    title: string
    created_at: string
    updated_at: string
}

export interface Message {
    id: string
    conversation_id: string
    role: 'user' | 'assistant' | 'system'
    content: string
    thinking?: string
    metadata?: string
    created_at: string
    updateKey?: number  // 用于强制 Vue 重新渲染 v-html 内容
    renderedContent?: string  // 预渲染的 HTML 内容，用于流式更新时的正确显示
}

export interface Book {
    id: number
    title: string
    authors?: string[]
    reason?: string
    score?: number
}

export async function createConversation(data: { title: string }): Promise<Conversation> {
    const response = await fetch('/api/chat/conversations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
    if (!response.ok) throw new Error('Failed to create conversation')
    return response.json()
}

export async function listConversations(): Promise<Conversation[]> {
    const response = await fetch('/api/chat/conversations')
    if (!response.ok) throw new Error('Failed to list conversations')
    return response.json()
}

export async function getConversation(id: string): Promise<Conversation> {
    const response = await fetch(`/api/chat/conversations/${id}`)
    if (!response.ok) throw new Error('Failed to get conversation')
    return response.json()
}

export async function getMessages(conversationId: string): Promise<Message[]> {
    const response = await fetch(`/api/chat/conversations/${conversationId}/messages`)
    if (!response.ok) throw new Error('Failed to get messages')
    return response.json()
}

export const deleteConversation = async (id: string): Promise<void> => {
    const response = await fetch(`/api/chat/conversations/${id}`, {
        method: 'DELETE'
    })
    if (!response.ok) throw new Error('Failed to delete conversation')
}

export const deleteMessage = async (id: string): Promise<void> => {
    const response = await fetch(`/api/chat/messages/${id}`, {
        method: 'DELETE'
    })
    if (!response.ok) throw new Error('Failed to delete message')
}

// SSE 流式响应由组件内部处理（见 Chat.vue）
