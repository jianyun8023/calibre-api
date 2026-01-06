'use client';

import { useTranslations } from 'next-intl';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { usePathname } from 'next/navigation';

export default function NotFound() {
  const t = useTranslations('common');
  const pathname = usePathname();
  const locale = pathname.split('/')[1] || 'zh-CN';
  
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="text-muted-foreground">Page not found</p>
      <Button asChild>
        <Link href={`/${locale}`}>Go Home</Link>
      </Button>
    </div>
  );
}
