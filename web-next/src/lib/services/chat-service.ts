/**
 * Chat Service
 * 
 * This service handles chat and conversation operations:
 * - Managing conversations
 * - Sending and receiving messages
 * - Streaming chat responses
 * 
 * @example
 * ```typescript
 * import { chatService } from '@/lib/services/chat-service'
 * 
 * // Get conversations
 * const conversations = await chatService.getConversations()
 * 
 * // Send message
 * const message = await chatService.sendMessage(convId, 'Hello!')
 * ```
 */

import { BaseApiService } from './base-service'
import type { UnifiedApiClient } from '../api-client-v2'
import { apiClient } from '../api-client-v2'
import type { ErrorHandler } from '../error-handler'
import { errorHandler } from '../error-handler'

// ============================================================================
// Type Definitions
// ============================================================================

/**
 * Conversation
 */
export interface Conversation {
  id: string
  title: string
  created_at: string
  updated_at: string
  message_count?: number
}

/**
 * Message
 */
export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
  metadata?: Record<string, unknown>
}

/**
 * Create Conversation Request
 */
export interface CreateConversationRequest {
  title?: string
}

/**
 * Chat Stream Request
 */
export interface ChatStreamRequest {
  conversationId?: string
  messages: Array<{
    role: 'user' | 'assistant' | 'system'
    content: string
  }>
}

/**
 * Chat Response (non-streaming)
 */
export interface ChatResponse {
  message: Message
  conversation: Conversation
}

// ============================================================================
// Chat Service Class
// ============================================================================

export class ChatService extends BaseApiService {
  constructor(client: UnifiedApiClient, errorHandler: ErrorHandler) {
    super(client, errorHandler)
  }

  // ==========================================================================
  // Conversation Operations
  // ==========================================================================

  /**
   * Get all conversations
   * 
   * @returns Array of conversations
   * 
   * @example
   * ```typescript
   * const conversations = await chatService.getConversations()
   * ```
   */
  async getConversations(): Promise<Conversation[]> {
    return this.handleRequest(async () => {
      return this.client.get<Conversation[]>('/api/chat/conversations')
    })
  }

  /**
   * Create a new conversation
   * 
   * @param data - Conversation data
   * @returns Created conversation
   * 
   * @example
   * ```typescript
   * const conversation = await chatService.createConversation({ 
   *   title: 'My Chat' 
   * })
   * ```
   */
  async createConversation(data: CreateConversationRequest): Promise<Conversation> {
    return this.handleRequest(async () => {
      return this.client.post<Conversation>('/api/chat/conversations', data)
    })
  }

  /**
   * Delete a conversation
   * 
   * @param id - Conversation ID
   * 
   * @example
   * ```typescript
   * await chatService.deleteConversation('conv-123')
   * ```
   */
  async deleteConversation(id: string): Promise<void> {
    return this.handleRequest(async () => {
      return this.client.delete<void>(`/api/chat/conversations/${id}`)
    })
  }

  // ==========================================================================
  // Message Operations
  // ==========================================================================

  /**
   * Get messages for a conversation
   * 
   * @param conversationId - Conversation ID
   * @returns Array of messages
   * 
   * @example
   * ```typescript
   * const messages = await chatService.getMessages('conv-123')
   * ```
   */
  async getMessages(conversationId: string): Promise<Message[]> {
    return this.handleRequest(async () => {
      return this.client.get<Message[]>(`/api/chat/conversations/${conversationId}/messages`)
    })
  }

  /**
   * Send a message (non-streaming)
   * 
   * @param conversationId - Conversation ID
   * @param content - Message content
   * @returns Chat response with message
   * 
   * @example
   * ```typescript
   * const response = await chatService.sendMessage('conv-123', 'Hello!')
   * console.log(response.message.content)
   * ```
   */
  async sendMessage(conversationId: string, content: string): Promise<ChatResponse> {
    return this.handleRequest(async () => {
      return this.client.post<ChatResponse>('/api/chat', {
        conversation_id: conversationId,
        message: content,
        stream: false,
      })
    })
  }

  /**
   * Delete a message
   * 
   * @param id - Message ID
   * 
   * @example
   * ```typescript
   * await chatService.deleteMessage('msg-123')
   * ```
   */
  async deleteMessage(id: string): Promise<void> {
    return this.handleRequest(async () => {
      return this.client.delete<void>(`/api/chat/messages/${id}`)
    })
  }

  // ==========================================================================
  // Streaming Operations
  // ==========================================================================

  /**
   * Stream chat response
   * 
   * @param request - Chat stream request
   * @param signal - Optional AbortSignal for cancellation
   * @returns ReadableStream of chat chunks
   * 
   * @example
   * ```typescript
   * const controller = new AbortController()
   * const stream = await chatService.streamChat({
   *   conversationId: 'conv-123',
   *   messages: [{ role: 'user', content: 'Tell me about TypeScript' }]
   * }, controller.signal)
   * 
   * const reader = stream.getReader()
   * while (true) {
   *   const { done, value } = await reader.read()
   *   if (done) break
   *   console.log(new TextDecoder().decode(value))
   * }
   * ```
   */
  async streamChat(request: ChatStreamRequest, signal?: AbortSignal): Promise<ReadableStream> {
    return this.handleRequest(async () => {
      const response = await fetch('/api/chat/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
        signal,
      })

      if (!response.ok) {
        throw new Error(`Chat stream failed: ${response.status} ${response.statusText}`)
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      return response.body
    })
  }
}

// ============================================================================
// Default Service Instance
// ============================================================================

/**
 * Default chat service instance
 * 
 * @example
 * ```typescript
 * import { chatService } from '@/lib/services/chat-service'
 * 
 * const conversations = await chatService.getConversations()
 * ```
 */
export const chatService = new ChatService(apiClient, errorHandler)

