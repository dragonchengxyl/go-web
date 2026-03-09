'use client';

import Link from 'next/link';
import { useState, useEffect } from 'react';
import { Menu, X, User, Bell, MessageCircle, Compass } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { GlobalSearch } from '@/components/search/global-search';
import { apiClient } from '@/lib/api-client';

export function Header() {
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 50);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) return;
    apiClient.setToken(token);

    const fetchUnread = async () => {
      try {
        const data = await apiClient.getUnreadCount();
        setUnreadCount(data.count);
      } catch {
        // not logged in or error — ignore
      }
    };

    fetchUnread();

    // Re-check every 60 seconds
    const interval = setInterval(fetchUnread, 60000);
    return () => clearInterval(interval);
  }, []);

  return (
    <header
      className={cn(
        'fixed top-0 left-0 right-0 z-50 transition-all duration-300',
        isScrolled
          ? 'bg-background/80 backdrop-blur-md border-b'
          : 'bg-transparent'
      )}
    >
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16 gap-4">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-2 flex-shrink-0">
            <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-bold text-sm">
              F
            </div>
            <span className="font-bold text-xl">Furry社区</span>
          </Link>

          {/* Search Bar - Desktop */}
          <div className="hidden md:block flex-1 max-w-2xl mx-4">
            <GlobalSearch />
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center space-x-6 flex-shrink-0">
            <Link
              href="/feed"
              className="text-sm font-medium hover:text-primary transition-colors"
            >
              动态
            </Link>
            <Link
              href="/explore"
              className="text-sm font-medium hover:text-primary transition-colors flex items-center gap-1"
            >
              <Compass className="h-4 w-4" />
              发现
            </Link>
            <Link
              href="/music"
              className="text-sm font-medium hover:text-primary transition-colors"
            >
              音乐
            </Link>
            <Link
              href="/leaderboard"
              className="text-sm font-medium hover:text-primary transition-colors"
            >
              排行
            </Link>
          </nav>

          {/* Actions */}
          <div className="flex items-center space-x-2">
            <Link href="/messages">
              <Button variant="ghost" size="icon" title="消息">
                <MessageCircle className="h-5 w-5" />
              </Button>
            </Link>
            <Link href="/notifications">
              <Button variant="ghost" size="icon" title="通知" className="relative">
                <Bell className="h-5 w-5" />
                {unreadCount > 0 && (
                  <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center font-medium leading-none">
                    {unreadCount > 9 ? '9+' : unreadCount}
                  </span>
                )}
              </Button>
            </Link>
            <Link href="/profile">
              <Button variant="ghost" size="icon" title="个人主页">
                <User className="h-5 w-5" />
              </Button>
            </Link>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            >
              {isMobileMenuOpen ? (
                <X className="h-5 w-5" />
              ) : (
                <Menu className="h-5 w-5" />
              )}
            </Button>
          </div>
        </div>

        {/* Mobile Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden py-4 border-t">
            <nav className="flex flex-col space-y-4">
              <Link
                href="/feed"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                动态
              </Link>
              <Link
                href="/explore"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                发现
              </Link>
              <Link
                href="/music"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                音乐
              </Link>
              <Link
                href="/messages"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                消息
              </Link>
              <Link
                href="/notifications"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                通知{unreadCount > 0 && ` (${unreadCount})`}
              </Link>
              <Link
                href="/profile"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                个人主页
              </Link>
              <Link
                href="/leaderboard"
                className="text-sm font-medium hover:text-primary transition-colors"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                排行榜
              </Link>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}
