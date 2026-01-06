import { getRequestConfig } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from './config';

export default getRequestConfig(async ({ requestLocale }) => {
  // Typically corresponds to the `[locale]` segment
  const requested = await requestLocale;
  
  // Validate that the incoming locale is valid
  const isValidLocale = locales.some(l => l === requested);
  if (!isValidLocale) notFound();

  return {
    locale: requested as string,
    messages: (await import(`../../messages/${requested}.json`)).default
  };
});
