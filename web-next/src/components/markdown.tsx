import { memo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import 'highlight.js/styles/github-dark.css' // Import highlight.js styles

export const MemoizedMarkdown = memo(({ content }: { content: string }) => (
  <div className="prose dark:prose-invert max-w-none break-words">
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={[rehypeHighlight]}
      components={{
          // Customize components if needed
          p: ({children}) => <p className="mb-2 last:mb-0">{children}</p>
      }}
    >
      {content}
    </ReactMarkdown>
  </div>
))

MemoizedMarkdown.displayName = 'MemoizedMarkdown'

