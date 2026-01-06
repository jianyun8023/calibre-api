import { setRequestLocale } from 'next-intl/server';
import { Suspense } from "react"
import { Skeleton } from "@/components/ui/skeleton"
import BooksClient from './books-client';

export default async function BooksPage({
  params
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  return (
    <Suspense fallback={
      <div className="p-8">
        <Skeleton className="h-10 w-40 mb-8" />
        <div className="grid grid-cols-4 gap-4">
          {Array.from({length:8}, (_, i) => `fallback-skeleton-${i}`).map((id) => (
            <Skeleton key={id} className="h-60" />
          ))}
        </div>
      </div>
    }>
      <BooksClient locale={locale} />
    </Suspense>
  );
}
