"use client";

import Link from "next/link";
import { useState, useEffect, useRef } from "react";
import {
  Menu,
  X,
  Bell,
  MessageCircle,
  Compass,
  PenSquare,
  LogOut,
  Settings,
  Users,
  Calendar,
  Trophy,
  ChevronDown,
  Flag,
  Bookmark,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { GlobalSearch } from "@/components/search/global-search";
import { apiClient } from "@/lib/api-client";
import { useWS } from "@/contexts/ws-context";
import { useAuth } from "@/contexts/auth-context";
import { usePathname, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";

const NAV_LINKS = [
  { href: "/feed", label: "动态" },
  { href: "/explore", label: "发现", icon: Compass },
  { href: "/groups", label: "圈子", icon: Users },
  { href: "/events", label: "活动", icon: Calendar },
  { href: "/leaderboard", label: "排行", icon: Trophy },
  { href: "/sponsor", label: "赞助", className: "text-orange-500" },
];

function UserAvatar() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  if (!user) return null;

  const initial = (user.username || "U")[0].toUpperCase();

  function handleLogout() {
    logout();
    setOpen(false);
    router.push("/");
  }

  return (
    <div className="flex items-center gap-1" ref={ref}>
      {/* 头像 — 直接跳转个人中心 */}
      <Link
        href="/profile"
        className="w-8 h-8 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center text-white font-bold text-sm shadow-sm hover:brightness-110 transition-all focus:outline-none focus:ring-2 focus:ring-brand-purple/50"
        title={`${user.username} 的个人中心`}
      >
        {initial}
      </Link>

      {/* 展开更多选项 */}
      <div className="relative">
        <button
          onClick={() => setOpen((v) => !v)}
          className="h-5 w-5 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors rounded focus:outline-none"
          title="更多选项"
        >
          <ChevronDown
            className={cn(
              "h-3.5 w-3.5 transition-transform duration-150",
              open && "rotate-180",
            )}
          />
        </button>

        <AnimatePresence>
          {open && (
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: -4 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: -4 }}
              transition={{ duration: 0.12 }}
              className="absolute right-0 mt-2 w-52 rounded-xl bg-background border border-border shadow-lg overflow-hidden z-50"
            >
              {/* 用户信息头 */}
              <Link
                href="/profile"
                onClick={() => setOpen(false)}
                className="flex items-center gap-3 px-3 py-3 border-b border-border hover:bg-muted/60 transition-colors"
              >
                <div className="w-9 h-9 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                  {initial}
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-semibold truncate">
                    {user.username}
                  </p>
                  <p className="text-xs text-muted-foreground truncate">
                    {user.email}
                  </p>
                </div>
              </Link>
              <div className="py-1">
                <Link
                  href="/creator"
                  onClick={() => setOpen(false)}
                  className="flex items-center gap-2.5 px-3 py-2 text-sm hover:bg-muted transition-colors"
                >
                  <PenSquare className="h-4 w-4 text-muted-foreground" />
                  创作中心
                </Link>
                <Link
                  href="/bookmarks"
                  onClick={() => setOpen(false)}
                  className="flex items-center gap-2.5 px-3 py-2 text-sm hover:bg-muted transition-colors"
                >
                  <Bookmark className="h-4 w-4 text-muted-foreground" />
                  我的收藏
                </Link>
                <Link
                  href="/reports"
                  onClick={() => setOpen(false)}
                  className="flex items-center gap-2.5 px-3 py-2 text-sm hover:bg-muted transition-colors"
                >
                  <Flag className="h-4 w-4 text-muted-foreground" />
                  我的举报
                </Link>
                <Link
                  href="/settings"
                  onClick={() => setOpen(false)}
                  className="flex items-center gap-2.5 px-3 py-2 text-sm hover:bg-muted transition-colors"
                >
                  <Settings className="h-4 w-4 text-muted-foreground" />
                  设置
                </Link>
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center gap-2.5 px-3 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors"
                >
                  <LogOut className="h-4 w-4" />
                  退出登录
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}

export function Header() {
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const { subscribe } = useWS();
  const { isLoggedIn, user, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 50);
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    if (!isLoggedIn) return;
    apiClient
      .getUnreadCount()
      .then((data) => setUnreadCount(data.count))
      .catch(() => {});
    return subscribe("notification", () => setUnreadCount((c) => c + 1));
  }, [subscribe, isLoggedIn]);

  function handleMobileLogout() {
    logout();
    setIsMobileMenuOpen(false);
    router.push("/");
  }

  return (
    <header
      className={cn(
        "fixed top-0 left-0 right-0 z-50 transition-all duration-300",
        isScrolled
          ? "bg-background/75 backdrop-blur-xl border-b border-border/50 shadow-sm"
          : "bg-transparent",
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
          <nav className="hidden md:flex items-center space-x-3 flex-shrink-0">
            {NAV_LINKS.map(({ href, label, icon: Icon, className }) => (
              <Link
                key={href}
                href={href}
                className={cn(
                  "text-sm font-medium transition-colors flex items-center gap-1",
                  pathname === href
                    ? "text-primary font-semibold"
                    : "text-muted-foreground hover:text-foreground",
                  className,
                )}
              >
                {Icon && <Icon className="h-4 w-4" />}
                {label}
              </Link>
            ))}

            {isLoggedIn ? (
              <>
                <Link href="/posts/create">
                  <Button
                    size="sm"
                    className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 transition-all"
                  >
                    <PenSquare className="h-4 w-4 mr-1" />
                    发帖
                  </Button>
                </Link>
                <Link href="/messages">
                  <Button variant="ghost" size="icon" title="消息">
                    <MessageCircle className="h-5 w-5" />
                  </Button>
                </Link>
                <Link href="/notifications">
                  <Button
                    variant="ghost"
                    size="icon"
                    title="通知"
                    className="relative"
                  >
                    <Bell className="h-5 w-5" />
                    {unreadCount > 0 && (
                      <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center font-medium leading-none">
                        {unreadCount > 9 ? "9+" : unreadCount}
                      </span>
                    )}
                  </Button>
                </Link>
                <UserAvatar />
              </>
            ) : (
              <div className="flex items-center gap-2">
                <Link href="/login">
                  <Button variant="ghost" size="sm">
                    登录
                  </Button>
                </Link>
                <Link href="/register">
                  <Button
                    size="sm"
                    className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 transition-all"
                  >
                    注册
                  </Button>
                </Link>
              </div>
            )}
          </nav>

          {/* Mobile menu toggle */}
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

        {/* Mobile Menu */}
        <AnimatePresence>
          {isMobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.2, ease: "easeInOut" }}
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
                        "px-3 py-2.5 rounded-lg text-sm font-medium transition-colors",
                        pathname === href
                          ? "bg-primary/10 text-primary"
                          : "text-muted-foreground hover:bg-muted hover:text-foreground",
                        className,
                      )}
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      {label}
                    </Link>
                  ))}

                  {isLoggedIn ? (
                    <>
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
                        href="/bookmarks"
                        className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        我的收藏
                      </Link>
                      <Link
                        href="/profile"
                        className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        {user?.username
                          ? `我的主页 (@${user.username})`
                          : "我的主页"}
                      </Link>
                      <button
                        onClick={handleMobileLogout}
                        className="px-3 py-2.5 rounded-lg text-sm font-medium text-destructive hover:bg-destructive/10 transition-colors text-left"
                      >
                        退出登录
                      </button>
                    </>
                  ) : (
                    <>
                      <Link
                        href="/login"
                        className="px-3 py-2.5 rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        登录
                      </Link>
                      <Link
                        href="/register"
                        className="px-3 py-2.5 rounded-lg text-sm font-medium text-primary hover:bg-primary/10 transition-colors"
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        注册
                      </Link>
                    </>
                  )}
                </nav>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </header>
  );
}
