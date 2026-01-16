"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useTranslations } from 'next-intl'
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Library,
  Search,
  MessageSquare,
  Settings,
  ListTodo,
  Building2,
  Home,
  BookOpen,
  FileEdit
} from "lucide-react"

export function AppSidebar({ className }: { className?: string }) {
  const pathname = usePathname()
  const t = useTranslations('nav')

  // Extract locale from pathname
  const locale = pathname.split('/')[1] || 'zh-CN'

  const sidebarItems = [
    { name: t('home'), href: `/${locale}`, icon: Home },
    { name: t('allBooks'), href: `/${locale}/books`, icon: Library },
    { name: t('search'), href: `/${locale}/search`, icon: Search },
    { name: t('metadataManager'), href: `/${locale}/metadata/manager`, icon: FileEdit },
    { name: t('chatAgent'), href: `/${locale}/chat`, icon: MessageSquare },
    { name: t('tasks'), href: `/${locale}/tasks`, icon: ListTodo },
    { name: t('publishers'), href: `/${locale}/publisher`, icon: Building2 },
    { name: t('settings'), href: `/${locale}/settings`, icon: Settings },
  ]

  return (
    <div className={cn(
      "pb-12 w-64 border-r",
      "glass-aurora",
      "transition-all duration-300",
      className
    )}>
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
          <h2 className="mb-2 px-4 text-lg font-semibold tracking-tight flex items-center gap-2">
            <BookOpen className="h-5 w-5 text-primary" />
            <span className="text-gradient dark:text-foreground">Calibre API</span>
          </h2>
          <div className="space-y-1">
            {sidebarItems.map((item) => (
              <Button
                key={item.href}
                variant={pathname === item.href ? "secondary" : "ghost"}
                className={cn(
                  "w-full justify-start transition-all duration-200",
                  "hover:translate-x-1 hover:bg-primary/10 dark:hover:bg-white/5",
                  pathname === item.href && "bg-primary/10 dark:bg-white/10 font-medium"
                )}
                asChild
              >
                <Link href={item.href}>
                  <item.icon className="mr-2 h-4 w-4" />
                  {item.name}
                </Link>
              </Button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

