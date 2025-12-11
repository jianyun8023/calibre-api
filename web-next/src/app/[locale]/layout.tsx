import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from '@/i18n/config';
import "../globals.css";
import { ThemeProvider } from "@/components/theme-provider"
import { SettingsProvider } from "@/contexts/settings-context"
import { Toaster } from "@/components/ui/sonner"
import { AppSidebar } from "@/components/app-sidebar"
import { AppHeader } from "@/components/app-header"

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
});

export const metadata: Metadata = {
  title: "Calibre API",
  description: "Modern Calibre Web Interface",
};

// Disable static generation for now
// export function generateStaticParams() {
//   return locales.map((locale) => ({ locale }));
// }

export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  
  // Validate that the incoming `locale` parameter is valid
  const isValidLocale = locales.some(l => l === locale);
  if (!isValidLocale) {
    notFound();
  }

  // Enable static rendering
  const { setRequestLocale } = await import('next-intl/server');
  setRequestLocale(locale);

  // Providing all messages to the client side is the easiest way to get started
  const messages = await getMessages();

  return (
    <html lang={locale} suppressHydrationWarning className={inter.variable}>
      <body className="antialiased min-h-screen bg-background font-sans">
        <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <SettingsProvider>
              <NextIntlClientProvider messages={messages}>
                <div className="flex h-screen flex-col md:flex-row overflow-hidden">
                  {/* Desktop Sidebar */}
                  <AppSidebar className="hidden md:block h-full sticky top-0 shrink-0" />
                  
                  <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                    <AppHeader className="shrink-0" />
                    <main className="flex-1 overflow-auto p-4 md:p-6 lg:p-8">
                      {children}
                    </main>
                  </div>
                </div>
                <Toaster />
              </NextIntlClientProvider>
            </SettingsProvider>
          </ThemeProvider>
      </body>
    </html>
  );
}
