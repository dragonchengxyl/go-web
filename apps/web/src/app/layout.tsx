import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Providers } from '@/components/providers';
import { MusicPlayer } from '@/components/music-player';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: '独立游戏工作室 - Indie Game Studio',
  description: '探索独立游戏的无限可能',
  keywords: ['独立游戏', 'indie game', 'game studio', 'OST', '游戏音乐'],
  authors: [{ name: 'Indie Game Studio' }],
  openGraph: {
    type: 'website',
    locale: 'zh_CN',
    url: 'https://studio.example.com',
    siteName: '独立游戏工作室',
    title: '独立游戏工作室 - Indie Game Studio',
    description: '探索独立游戏的无限可能',
  },
  twitter: {
    card: 'summary_large_image',
    title: '独立游戏工作室',
    description: '探索独立游戏的无限可能',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className={inter.className}>
        <Providers>
          {children}
          <MusicPlayer />
        </Providers>
      </body>
    </html>
  );
}
