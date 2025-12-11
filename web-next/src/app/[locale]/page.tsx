import { setRequestLocale } from 'next-intl/server';
import HomeClient from './home-client';

export default async function Home({
  params
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  return <HomeClient locale={locale} />;
}
