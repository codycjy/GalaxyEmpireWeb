import { auth } from '@/auth';
import Providers from '@/components/providers';
import { Toaster } from '@/components/ui/sonner';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import NextTopLoader from 'nextjs-toploader';
import './globals.css';

export const metadata: Metadata = {
  title: 'Next Shadcn',
  description: 'Basic dashboard with Next.js and Shadcn'
};

const inter = Inter({
  subsets: ['latin'],
  display: 'swap'
});

export default async function RootLayout({
  children
}: {
  children: React.ReactNode;
}) {
  const session = await auth();  // 获取 session
  
  return (
    <html
      lang="en"
      className={`${inter.className}`}
      suppressHydrationWarning={true}
    >
      <body className={'overflow-hidden'}>
        <Providers session={session}>  {/* 传递 session 到 Providers */}
          <NextTopLoader showSpinner={false} />
          <Toaster />
          {children}
        </Providers>
      </body>
    </html>
  );
}
