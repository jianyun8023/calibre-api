"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
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

const sidebarItems = [
  { name: "Home", href: "/", icon: Home },
  { name: "All Books", href: "/books", icon: Library },
  { name: "Search", href: "/search", icon: Search },
  { name: "Metadata Manager", href: "/metadata/manager", icon: FileEdit },
  { name: "Chat Agent", href: "/chat", icon: MessageSquare },
  { name: "Tasks", href: "/tasks", icon: ListTodo },
  { name: "Publishers", href: "/publisher", icon: Building2 },
  { name: "Settings", href: "/settings", icon: Settings },
]

export function AppSidebar({ className }: { className?: string }) {
  const pathname = usePathname()

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

