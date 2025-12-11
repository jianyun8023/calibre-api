import type { Metadata } from "next";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider"
import { Toaster } from "@/components/ui/sonner"
import { AppSidebar } from "@/components/app-sidebar"
import { AppHeader } from "@/components/app-header"

export const metadata: Metadata = {
  title: "Calibre API",
  description: "Modern Calibre Web Interface",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased min-h-screen bg-background font-sans">
        <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <div className="flex min-h-screen flex-col md:flex-row">
              {/* Desktop Sidebar */}
              <AppSidebar className="hidden md:block h-screen sticky top-0" />
              
              <div className="flex-1 flex flex-col min-w-0">
                <AppHeader />
                <main className="flex-1 p-4 md:p-6 lg:p-8">
                  {children}
                </main>
              </div>
            </div>
            <Toaster />
          </ThemeProvider>
      </body>
    </html>
  );
}
