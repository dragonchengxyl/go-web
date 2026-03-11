'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import Link from 'next/link'
import {
  LayoutDashboard,
  BarChart2,
  ShieldCheck,
  Flag,
  Users,
  MessageSquare,
} from 'lucide-react'

const navItems = [
  { icon: LayoutDashboard, label: '总览', href: '/admin' },
  { icon: BarChart2, label: '数据分析', href: '/admin/analytics' },
  { icon: ShieldCheck, label: '内容审核', href: '/admin/moderation' },
  { icon: Flag, label: '举报处理', href: '/admin/reports' },
  { icon: Users, label: '用户管理', href: '/admin/users' },
  { icon: MessageSquare, label: '评论管理', href: '/admin/comments' },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()

  useEffect(() => {
    try {
      const token = localStorage.getItem('access_token')
      if (!token) { router.replace('/login'); return }
      const payload = JSON.parse(atob(token.split('.')[1]))
      if (!['admin', 'moderator', 'super_admin'].includes(payload.role)) {
        router.replace('/')
      }
    } catch {
      router.replace('/login')
    }
  }, [router])

  const isActive = (href: string) =>
    href === '/admin' ? pathname === '/admin' : pathname.startsWith(href)

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-56 bg-gray-950 text-gray-300 flex flex-col shrink-0">
        <div className="px-5 py-5 border-b border-gray-800">
          <span className="text-white font-bold text-base">管理控制台</span>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1">
          {navItems.map(({ icon: Icon, label, href }) => (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
                isActive(href)
                  ? 'bg-gray-800 text-white'
                  : 'hover:bg-gray-800/60 hover:text-white'
              }`}
            >
              <Icon size={16} className="shrink-0" />
              {label}
            </Link>
          ))}
        </nav>

        <div className="px-5 py-4 border-t border-gray-800">
          <Link
            href="/"
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            返回前台 ↗
          </Link>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto bg-gray-50 dark:bg-gray-900 p-6">
        {children}
      </main>
    </div>
  )
}
