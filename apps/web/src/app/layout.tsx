import type { Metadata } from 'next';
import './globals.css';
import { GeistSans } from 'geist/font/sans';
import { Providers } from '@/components/providers';
import { Header } from '@/components/layout/header';
import { MusicPlayer } from '@/components/music-player';
import { ModerationToast } from '@/components/moderation-toast';
import { LenisProvider } from '@/components/lenis-provider';

export const metadata: Metadata = {
  title: 'Furry 同好社区',
  description: '毛毛们的温暖家园，分享兽设、创作与生活',
  keywords: ['furry', 'fursuit', '兽迷', '兽设', '同好社区'],
  authors: [{ name: 'Furry 同好社区' }],
  openGraph: {
    type: 'website',
    locale: 'zh_CN',
    url: 'https://furry.example.com',
    siteName: 'Furry 同好社区',
    title: 'Furry 同好社区',
    description: '毛毛们的温暖家园',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Furry 同好社区',
    description: '毛毛们的温暖家园',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning className={GeistSans.variable}>
      <body className={GeistSans.className}>
        <LenisProvider>
          <Providers>
            <Header />
            {children}
            <MusicPlayer />
            <ModerationToast />
          </Providers>
        </LenisProvider>
      </body>
    </html>
  );
}
