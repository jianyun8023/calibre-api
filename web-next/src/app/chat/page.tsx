"use client"

import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { MemoizedMarkdown } from '@/components/markdown'
import { Send, StopCircle, User, Bot, Plus, Trash2, RefreshCw } from 'lucide-react'
import { BookCard } from '@/components/book-card'
import type { Book } from '@/types/book'
import { useState, useEffect, useRef, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { useToast } from '@/hooks/use-toast'
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from '@/components/ui/collapsible'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'

interface Conversation {
    id: string
    title: string
    updated_at: string
}

interface ChatMessage {
    id: string
    role: 'user' | 'assistant'
    content: string
    thinking?: string
    books?: Book[]
}

export default function ChatPage() {
  const { toast } = useToast()
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [currentConversation, setCurrentConversation] = useState<Conversation | null>(null)
  const [input, setInput] = useState('')
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([])
  const [bookPageMap, setBookPageMap] = useState<Record<string, number>>({})
  const [deleteDialog, setDeleteDialog] = useState<{ type: 'conversation' | 'message', id: string } | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const abortControllerRef = useRef<AbortController | null>(null)
  
  const scrollRef = useRef<HTMLDivElement>(null)
  
  const scrollToBottom = useCallback(() => {
      if (scrollRef.current) {
          scrollRef.current.scrollTop = scrollRef.current.scrollHeight
      }
  }, [])
  
  useEffect(() => {
      scrollToBottom()
  }, [chatMessages, scrollToBottom])

  // Load conversations on mount
  useEffect(() => {
    loadConversations()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Load messages when conversation changes
  useEffect(() => {
    if (currentConversation) {
      loadMessages(currentConversation.id)
    } else {
      setChatMessages([])
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentConversation])

  const loadConversations = async () => {
    try {
      const res = await fetch('/api/chat/conversations')
      if (!res.ok) throw new Error('Failed to load conversations')
      const data = await res.json()
      const convs = data.data || []
      setConversations(convs)
      
      // Select first conversation if none selected
      if (!currentConversation && convs.length > 0) {
        setCurrentConversation(convs[0])
      }
    } catch (error) {
      console.error('Load conversations error:', error)
      toast({
        title: '加载对话列表失败',
        variant: 'destructive'
      })
    }
  }

  const loadMessages = async (conversationId: string) => {
    try {
      const res = await fetch(`/api/chat/conversations/${conversationId}/messages`)
      if (!res.ok) throw new Error('Failed to load messages')
      const data = await res.json()
      const msgs = data.data || []
      
      // Convert to ChatMessage format
      const chatMsgs: ChatMessage[] = msgs.map((msg: {
        id: string;
        role: 'user' | 'assistant';
        content: string;
        thinking?: string;
        metadata?: string;
      }) => {
        const chatMsg: ChatMessage = {
          id: msg.id,
          role: msg.role,
          content: msg.content,
          thinking: msg.thinking,
        }
        
        // Parse metadata for books
        if (msg.metadata) {
          try {
            const meta = JSON.parse(msg.metadata)
            if (meta.books) {
              chatMsg.books = meta.books
            }
          } catch (e) {
            console.error('Parse metadata error:', e)
          }
        }
        
        return chatMsg
      })
      
      setChatMessages(chatMsgs)
    } catch (error) {
      console.error('Load messages error:', error)
      toast({
        title: '加载消息失败',
        variant: 'destructive'
      })
    }
  }

  const createNewConversation = async () => {
    try {
      const res = await fetch('/api/chat/conversations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: '新对话' })
      })
      if (!res.ok) throw new Error('Failed to create conversation')
      const newConv = await res.json()  // API直接返回对话对象
      
      setConversations([newConv, ...conversations])
      setCurrentConversation(newConv)
      setChatMessages([])
      
      toast({
        title: '创建成功',
        description: '已创建新对话'
      })
      
      return newConv
    } catch (error) {
      console.error('Create conversation error:', error)
      toast({
        title: '创建失败',
        variant: 'destructive'
      })
      return null
    }
  }

  const deleteConversation = async (id: string) => {
    try {
      const res = await fetch(`/api/chat/conversations/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error('Failed to delete conversation')
      
      setConversations(conversations.filter(c => c.id !== id))
      
      if (currentConversation?.id === id) {
        const remaining = conversations.filter(c => c.id !== id)
        setCurrentConversation(remaining.length > 0 ? remaining[0] : null)
      }
      
      toast({
        title: '删除成功'
      })
    } catch (error) {
      console.error('Delete conversation error:', error)
      toast({
        title: '删除失败',
        variant: 'destructive'
      })
    }
  }

  const deleteMessage = async (id: string) => {
    try {
      const res = await fetch(`/api/chat/messages/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error('Failed to delete message')
      
      setChatMessages(chatMessages.filter(m => m.id !== id))
      
      toast({
        title: '删除成功'
      })
    } catch (error) {
      console.error('Delete message error:', error)
      toast({
        title: '删除失败',
        variant: 'destructive'
      })
    }
  }

  const handleSend = async () => {
    if (!input.trim() || isLoading) return
    
    let conv = currentConversation
    
    // If no conversation, create one first
    if (!conv) {
      const newConv = await createNewConversation()
      if (!newConv) return
      conv = newConv
      setCurrentConversation(newConv)
    }
    
    const userMessage = input
    setInput('')
    setIsLoading(true)
    
    // Add user message to UI
    const userMsg: ChatMessage = {
      id: `temp-${Date.now()}`,
      role: 'user',
      content: userMessage
    }
    setChatMessages(prev => [...prev, userMsg])
    
    // Prepare AI message
    const aiMsg: ChatMessage = {
      id: `ai-${Date.now()}`,
      role: 'assistant',
      content: ''
    }
    setChatMessages(prev => [...prev, aiMsg])
    
    try {
      // Create AbortController for cancellation
      const controller = new AbortController()
      abortControllerRef.current = controller
      
      if (!conv) {
        throw new Error('No conversation available')
      }
      
      const res = await fetch('/api/chat/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          conversationId: conv.id,
          messages: [{ role: 'user', content: userMessage }]
        }),
        signal: controller.signal
      })
      
      if (!res.ok) throw new Error('Failed to send message')
      
      const reader = res.body?.getReader()
      if (!reader) throw new Error('No response body')
      
      const decoder = new TextDecoder()
      let buffer = ''
      
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.substring(6)
          if (data === '[DONE]') continue
          
          try {
            const event = JSON.parse(data)
            
            if (event.type === 'conversation') {
              // Update conversation ID
              setCurrentConversation(prev => prev ? { ...prev, id: event.conversationId } : null)
            } else if (event.type === 'text-delta') {
              // Update content
              setChatMessages(prev => {
                const last = prev[prev.length - 1]
                if (last.role === 'assistant') {
                  return [...prev.slice(0, -1), { ...last, content: last.content + event.delta }]
                }
                return prev
              })
            } else if (event.type === 'data-books') {
              // Add books
              setChatMessages(prev => {
                const last = prev[prev.length - 1]
                if (last.role === 'assistant') {
                  return [...prev.slice(0, -1), { ...last, books: event.data.books }]
                }
                return prev
              })
            } else if (event.type === 'data-thinking') {
              // Add thinking
              setChatMessages(prev => {
                const last = prev[prev.length - 1]
                if (last.role === 'assistant') {
                  return [...prev.slice(0, -1), { ...last, thinking: event.thinking }]
                }
                return prev
              })
            }
          } catch (e) {
            console.error('Parse SSE event error:', e)
          }
        }
      }
      
      // Reload conversations to update title
      await loadConversations()
    } catch (error: unknown) {
      if ((error as {name?: string}).name === 'AbortError') {
        toast({ title: '已停止生成' })
      } else {
        console.error('Send message error:', error)
        toast({
          title: '发送失败',
          description: (error as Error).message,
          variant: 'destructive'
        })
      }
      // Remove AI message if error
      setChatMessages(prev => prev.slice(0, -1))
    } finally {
      setIsLoading(false)
      abortControllerRef.current = null
    }
  }

  const stopGeneration = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const summarizeBook = (book: Book) => {
    setInput(`总结一下《${book.title}》这本书的内容`)
  }

  const changeBooks = (messageId: string, totalBooks: number) => {
    const currentPage = bookPageMap[messageId] || 0
    const nextPage = (currentPage + 1) * 8 >= totalBooks ? 0 : currentPage + 1
    setBookPageMap({ ...bookPageMap, [messageId]: nextPage })
  }

  const getVisibleBooks = (books: Book[], messageId: string) => {
    const page = bookPageMap[messageId] || 0
    return books.slice(page * 8, (page + 1) * 8)
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

  return (
    <div className="flex h-[calc(100vh-8rem)] gap-4">
      {/* Sidebar: Conversations (fixed width, scrollable) */}
      <div className="w-64 flex-col gap-2 border-r pr-4 hidden md:flex">
        <Button 
          className="w-full justify-start gap-2" 
          variant="secondary"
          onClick={createNewConversation}
        >
            <Plus className="h-4 w-4" /> 新建对话
        </Button>
        
        <ScrollArea className="flex-1">
          <div className="space-y-2">
            {conversations.map((conv) => (
              <div
                key={conv.id}
                className={cn(
                  "w-full flex items-center justify-between p-3 rounded-lg cursor-pointer transition-colors hover:bg-accent",
                  currentConversation?.id === conv.id && "bg-accent border-l-4 border-primary"
                )}
                onClick={() => setCurrentConversation(conv)}
              >
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-sm truncate">{conv.title}</div>
                  <div className="text-xs text-muted-foreground">{formatTime(conv.updated_at)}</div>
                </div>
                {currentConversation?.id === conv.id && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0"
                    onClick={(e) => {
                      e.stopPropagation()
                      setDeleteDialog({ type: 'conversation', id: conv.id })
                    }}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                )}
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>
      
      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col min-w-0 bg-background/50 rounded-lg border shadow-sm h-full">
        {!currentConversation && chatMessages.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center text-muted-foreground">
            <Bot className="h-16 w-16 mb-4" />
            <p>选择或创建一个对话开始聊天</p>
          </div>
        ) : (
          <>
            <ScrollArea className="flex-1 p-4" ref={scrollRef}>
              <div className="space-y-6 max-w-3xl mx-auto">
                {chatMessages.map((msg) => (
                  <div key={msg.id} className={cn("flex gap-3", msg.role === 'user' && "flex-row-reverse")}>
                    <div className={cn(
                      "shrink-0 w-8 h-8 rounded-full flex items-center justify-center",
                      msg.role === 'user' ? "bg-primary text-primary-foreground" : "bg-muted"
                    )}>
                      {msg.role === 'user' ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
                    </div>
                    
                    <div className="flex-1 space-y-2">
                      {/* Thinking process (collapsible) */}
                      {msg.thinking && (
                        <Collapsible>
                          <CollapsibleTrigger asChild>
                            <Button variant="outline" size="sm" className="w-full justify-start">
                              💭 思考过程
                            </Button>
                          </CollapsibleTrigger>
                          <CollapsibleContent>
                            <Card className="p-3 mt-2 bg-muted/50">
                              <pre className="text-xs whitespace-pre-wrap font-mono">{msg.thinking}</pre>
                            </Card>
                          </CollapsibleContent>
                        </Collapsible>
                      )}
                      
                      {/* Message content */}
                      <Card className={cn(
                        "p-4 group relative",
                        msg.role === 'user' ? "bg-primary text-primary-foreground" : "bg-muted/50"
                      )}>
                        <MemoizedMarkdown content={msg.content} />
                        
                        {/* Delete button */}
                        <Button
                          variant="ghost"
                          size="icon"
                          className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity h-6 w-6"
                          onClick={() => setDeleteDialog({ type: 'message', id: msg.id })}
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </Card>
                      
                      {/* Book cards */}
                      {msg.books && msg.books.length > 0 && (
                        <div className="space-y-2">
                          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                            {getVisibleBooks(msg.books, msg.id).map((book) => (
                              <div key={book.id} className="relative group">
                                <BookCard book={book} />
                                <Button
                                  variant="secondary"
                                  size="sm"
                                  className="w-full mt-2"
                                  onClick={() => summarizeBook(book)}
                                >
                                  总结此书
                                </Button>
                              </div>
                            ))}
                          </div>
                          
                          {/* Change books button */}
                          {msg.books.length > 8 && (
                            <div className="flex justify-center">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => changeBooks(msg.id, msg.books?.length || 0)}
                              >
                                <RefreshCw className="h-4 w-4 mr-2" /> 换一换
                              </Button>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
                
                {/* Loading state */}
                {isLoading && (
                  <div className="flex gap-3">
                    <div className="shrink-0 w-8 h-8 rounded-full bg-muted flex items-center justify-center">
                      <Bot className="h-4 w-4" />
                    </div>
                    <Card className="p-4 bg-muted/50 flex items-center gap-3">
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
                      <span className="text-sm text-muted-foreground">AI 正在思考...</span>
                      <Button variant="ghost" size="sm" onClick={stopGeneration}>
                        <StopCircle className="h-4 w-4 mr-2" /> 停止生成
                      </Button>
                    </Card>
                  </div>
                )}
              </div>
            </ScrollArea>
            
            {/* Input area */}
            <div className="border-t p-4 bg-background">
              <div className="flex gap-2 max-w-3xl mx-auto">
                <Input
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="输入消息... (Shift+Enter 换行，Enter 发送)"
                  disabled={isLoading}
                  className="flex-1"
                />
                {isLoading ? (
                  <Button onClick={stopGeneration} variant="destructive">
                    <StopCircle className="h-4 w-4 mr-2" /> 停止
                  </Button>
                ) : (
                  <Button onClick={handleSend}>
                    <Send className="h-4 w-4 mr-2" /> 发送
                  </Button>
                )}
              </div>
              <div className="text-xs text-muted-foreground text-center mt-2">
                Shift+Enter 换行，Enter 发送
              </div>
            </div>
          </>
        )}
      </div>
      
      {/* Delete confirmation dialog */}
      <AlertDialog open={!!deleteDialog} onOpenChange={() => setDeleteDialog(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteDialog?.type === 'conversation' 
                ? '确定要删除这个对话吗？此操作不可撤销。'
                : '确定要删除这条消息吗？此操作不可撤销。'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (deleteDialog) {
                  if (deleteDialog.type === 'conversation') {
                    deleteConversation(deleteDialog.id)
                  } else {
                    deleteMessage(deleteDialog.id)
                  }
                }
                setDeleteDialog(null)
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
