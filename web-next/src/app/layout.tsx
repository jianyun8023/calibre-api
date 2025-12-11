// Since we have a `[locale]` layout that provides <html> and <body>,
// this root layout just passes children through
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
