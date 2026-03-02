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
    { name: t('tasks'), href: `/${locale}/tasks`, icon: ListTodo },
    { name: t('publishers'), href: `/${locale}/publisher`, icon: Building2 },
    { name: t('settings'), href: `/${locale}/settings`, icon: Settings },
  ]

  return (
    <div className={cn("pb-12 w-64 border-r bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60", className)}>
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
          <h2 className="mb-2 px-4 text-lg font-semibold tracking-tight flex items-center gap-2">
            <BookOpen className="h-5 w-5 text-primary" />
            Calibre API
          </h2>
          <div className="space-y-1">
            {sidebarItems.map((item) => (
              <Button
                key={item.href}
                variant={pathname === item.href ? "secondary" : "ghost"}
                className="w-full justify-start"
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

