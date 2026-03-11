'use client';

import Link from 'next/link';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { Home, Compass } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 relative overflow-hidden">
      {/* Background blobs */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-purple/10 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-64 h-64 bg-brand-teal/10 rounded-full blur-3xl pointer-events-none" />

      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="relative z-10 text-center"
      >
        {/* Big 404 */}
        <motion.div
          initial={{ scale: 0.8 }}
          animate={{ scale: 1 }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
          className="text-[9rem] font-black leading-none bg-gradient-to-br from-brand-purple to-brand-teal bg-clip-text text-transparent select-none mb-4"
        >
          404
        </motion.div>

        {/* Furry icon */}
        <motion.div
          animate={{ y: [0, -8, 0] }}
          transition={{ duration: 2.5, repeat: Infinity, ease: 'easeInOut' }}
          className="text-5xl mb-6"
        >
          🐾
        </motion.div>

        <h1 className="text-2xl font-bold mb-2">迷路的毛毛</h1>
        <p className="text-muted-foreground text-sm mb-8 max-w-xs">
          这个页面好像跑去玩了，找不到它了。让我们回到社区吧！
        </p>

        <div className="flex gap-3 justify-center">
          <Link href="/">
            <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
              <Home className="h-4 w-4 mr-1.5" />
              回到首页
            </Button>
          </Link>
          <Link href="/explore">
            <Button variant="outline">
              <Compass className="h-4 w-4 mr-1.5" />
              去探索
            </Button>
          </Link>
        </div>
      </motion.div>
    </div>
  );
}
