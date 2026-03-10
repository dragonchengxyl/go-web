'use client';

import Link from 'next/link';
import { useState, useEffect } from 'react';
import { Menu, X, Bell, MessageCircle, Compass, PenSquare } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { GlobalSearch } from '@/components/search/global-search';
import { apiClient } from '@/lib/api-client';
import { useWS } from '@/contexts/ws-context';
import { usePathname } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';

const NAV_LINKS = [
  { href: '/feed', label: '动态' },
  { href: '/explore', label: '发现', icon: Compass },
  { href: '/sponsor', label: '赞助', className: 'text-orange-500' },
];

export function Header() {
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const { subscribe } = useWS();
  const pathname = usePathname();

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) return;
    apiClient.setToken(token);
    apiClient.getUnreadCount().then(data => setUnreadCount(data.count)).catch(() => {});
    return subscribe('notification', () => setUnreadCount(c => c + 1));
  }, [subscribe]);

  return (
    <header
      className={cn(
        'fixed top-0 left-0 right-0 z-50 transition-all duration-300',
        isScrolled
          ? 'bg-background/75 backdrop-blur-xl border-b border-border/50 shadow-sm'
          : 'bg-transparent'
      )}
    >
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16 gap-4">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-2 flex-shrink-0">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center text-white font-bold text-sm shadow-sm">
              F
            </div>
            <span className="font-bold text-xl bg-gradient-to-r from-brand-purple to-brand-teal bg-clip-text text-transparent">
              Furry社区
            </span>
          </Link>

          {/* Search Bar - Desktop */}
          <div className="hidden md:block flex-1 max-w-2xl mx-4">
            <GlobalSearch />
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center space-x-5 flex-shrink-0">
            {NAV_LINKS.map(({ href, label, icon: Icon, className }) => (
              <Link
                key={href}
                href={href}
                className={cn(
                  'text-sm font-medium transition-colors flex items-center gap-1',
                  pathname === href
                    ? 'text-primary font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                  className
                )}
              >
                {Icon && <Icon className="h-4 w-4" />}
                {label}
              </Link>
            ))}
            <Link href="/posts/create">
              <Button
                size="sm"
                className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 transition-all"
              >
                <PenSquare className="h-4 w-4 mr-1" />
                发帖
              </Button>
            </Link>
          </nav>

          {/* Actions */}
          <div className="flex items-center space-x-1">
            <Link href="/messages" className="hidden md:block">
              <Button variant="ghost" size="icon" title="消息">
                <MessageCircle className="h-5 w-5" />
              </Button>
            </Link>
            <Link href="/notifications" className="hidden md:block">
              <Button variant="ghost" size="icon" title="通知" className="relative">
                <Bell className="h-5 w-5" />
                {unreadCount > 0 && (
                  <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center font-medium leading-none">
                    {unreadCount > 9 ? '9+' : unreadCount}
                  </span>
                )}
              </Button>
            </Link>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            >
              {isMobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </Button>
          </div>
        </div>

        {/* Mobile Menu */}
        <AnimatePresence>
          {isMobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.2, ease: 'easeInOut' }}
              className="md:hidden overflow-hidden border-t"
            >
              <div className="py-4">
                <div className="mb-4">
                  <GlobalSearch />
                </div>
                <nav className="flex flex-col space-y-1">
                  {NAV_LINKS.map(({ href, label, className }) => (
                    <Link
                      key={href}
                      href={href}
                      className={cn(
                        'px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                        pathname === href
                          ? 'bg-primary/10 text-primary'
                          : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                        className
                      )}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      {label}
                    </Link>
                  ))}
                  <Link
                    href="/posts/create"
                    className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    发帖
                  </Link>
                  <Link
                    href="/messages"
                    className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    消息
                  </Link>
                  <Link
                    href="/notifications"
                    className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    通知{unreadCount > 0 && ` (${unreadCount})`}
                  </Link>
                  <Link
                    href="/profile"
                    className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    个人主页
                  </Link>
                </nav>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </header>
  );
}
