"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import {
  ClipboardList,
  BarChart3,
  Bot,
  CalendarRange,
  Flag,
  LayoutDashboard,
  MessageSquare,
  Shield,
  Receipt,
  ShieldCheck,
  SlidersHorizontal,
  Shapes,
  Users,
} from "lucide-react";
import { cn } from "@/lib/utils";

type NavItem = {
  href: string;
  label: string;
  description: string;
  section: string;
  icon: React.ComponentType<{ className?: string }>;
};

const navItems: NavItem[] = [
  {
    href: "/admin",
    label: "运营总览",
    description: "查看核心指标与待办队列",
    section: "总览",
    icon: LayoutDashboard,
  },
  {
    href: "/admin/analytics",
    label: "数据分析",
    description: "观察增长与业务趋势",
    section: "监控",
    icon: BarChart3,
  },
  {
    href: "/admin/moderation",
    label: "内容审核",
    description: "处理待审帖子与内容状态",
    section: "治理",
    icon: ShieldCheck,
  },
  {
    href: "/admin/reports",
    label: "举报处理",
    description: "处置用户举报与动作闭环",
    section: "治理",
    icon: Flag,
  },
  {
    href: "/admin/users",
    label: "用户管理",
    description: "角色、状态与账号治理",
    section: "治理",
    icon: Users,
  },
  {
    href: "/admin/orders",
    label: "订单运营",
    description: "打赏订单与支付状态管理",
    section: "运营",
    icon: Receipt,
  },
  {
    href: "/admin/groups",
    label: "圈子运营",
    description: "调整圈子可见性与运营状态",
    section: "运营",
    icon: Shapes,
  },
  {
    href: "/admin/events",
    label: "活动运营",
    description: "活动状态流转与报名巡检",
    section: "运营",
    icon: CalendarRange,
  },
  {
    href: "/admin/comments",
    label: "评论管理",
    description: "巡检评论区和异常内容",
    section: "治理",
    icon: MessageSquare,
  },
  {
    href: "/admin/audit-logs",
    label: "审计日志",
    description: "追踪关键后台操作记录",
    section: "治理",
    icon: ClipboardList,
  },
  {
    href: "/admin/assistant",
    label: "AI 助手",
    description: "管理人设、提示词和来源",
    section: "配置",
    icon: Bot,
  },
  {
    href: "/admin/permissions",
    label: "权限矩阵",
    description: "查看角色与权限映射关系",
    section: "配置",
    icon: Shield,
  },
  {
    href: "/admin/system",
    label: "系统配置",
    description: "查看运行态配置并维护赞助展示",
    section: "配置",
    icon: SlidersHorizontal,
  },
];

const allowedRoles = new Set(["admin", "moderator", "super_admin"]);

function isActive(pathname: string, href: string) {
  return href === "/admin" ? pathname === href : pathname.startsWith(href);
}

function getCurrentItem(pathname: string) {
  return (
    navItems.find((item) => isActive(pathname, item.href)) ??
    navItems[0]
  );
}

export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    try {
      const token = localStorage.getItem("access_token");
      if (!token) {
        router.replace("/login?from=/admin");
        return;
      }
      const payload = JSON.parse(atob(token.split(".")[1]));
      if (!allowedRoles.has(payload.role)) {
        router.replace("/");
        return;
      }
      setReady(true);
    } catch {
      router.replace("/login?from=/admin");
    }
  }, [router]);

  const current = useMemo(() => getCurrentItem(pathname), [pathname]);

  if (!ready) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#f4f7fb]">
        <div className="rounded-3xl border border-slate-200 bg-white px-6 py-5 text-sm text-slate-500 shadow-sm">
          正在进入运营后台...
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#f4f7fb] text-slate-950">
      <div className="pointer-events-none fixed inset-0 opacity-60">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(59,130,246,0.07),_transparent_26%),radial-gradient(circle_at_top_right,_rgba(14,165,233,0.08),_transparent_28%),linear-gradient(to_bottom,_rgba(255,255,255,0.96),_rgba(244,247,251,0.92))]" />
        <div className="absolute inset-0 bg-[linear-gradient(to_right,rgba(15,23,42,0.04)_1px,transparent_1px),linear-gradient(to_bottom,rgba(15,23,42,0.04)_1px,transparent_1px)] bg-[size:24px_24px]" />
      </div>

      <div className="relative flex min-h-screen">
        <aside className="hidden w-72 shrink-0 flex-col border-r border-slate-200/80 bg-slate-950 text-slate-100 lg:flex">
          <div className="border-b border-white/10 px-6 py-6">
            <p className="text-xs uppercase tracking-[0.28em] text-sky-300/75">
              Operations
            </p>
            <h1 className="mt-2 text-xl font-semibold tracking-tight">
              运营后台
            </h1>
            <p className="mt-2 text-sm leading-6 text-slate-400">
              面向内容治理、用户运营和配置管理的统一工作台。
            </p>
          </div>

          <nav className="flex-1 space-y-6 overflow-y-auto px-4 py-6">
            {["总览", "监控", "治理", "运营", "配置"].map((section) => {
              const items = navItems.filter((item) => item.section === section);
              if (items.length === 0) return null;

              return (
                <div key={section}>
                  <p className="px-3 text-[11px] uppercase tracking-[0.24em] text-slate-500">
                    {section}
                  </p>
                  <div className="mt-3 space-y-1.5">
                    {items.map((item) => {
                      const Icon = item.icon;
                      const active = isActive(pathname, item.href);
                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          className={cn(
                            "group flex items-start gap-3 rounded-2xl px-3 py-3 transition-colors",
                            active
                              ? "bg-white text-slate-950 shadow-[0_12px_40px_rgba(15,23,42,0.22)]"
                              : "text-slate-300 hover:bg-white/8 hover:text-white",
                          )}
                        >
                          <span
                            className={cn(
                              "mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border",
                              active
                                ? "border-slate-200 bg-slate-100 text-slate-950"
                                : "border-white/10 bg-white/5 text-slate-300",
                            )}
                          >
                            <Icon className="h-4 w-4" />
                          </span>
                          <span className="min-w-0">
                            <span className="block text-sm font-medium">
                              {item.label}
                            </span>
                            <span
                              className={cn(
                                "mt-1 block text-xs leading-5",
                                active ? "text-slate-500" : "text-slate-400",
                              )}
                            >
                              {item.description}
                            </span>
                          </span>
                        </Link>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </nav>

          <div className="border-t border-white/10 px-6 py-5">
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
              <p className="text-sm font-medium text-white">当前模式</p>
              <p className="mt-1 text-xs leading-5 text-slate-400">
                仅管理员与版主角色可进入此工作台。
              </p>
              <Link
                href="/"
                className="mt-4 inline-flex text-xs font-medium text-sky-300 hover:text-sky-200"
              >
                返回前台
              </Link>
            </div>
          </div>
        </aside>

        <div className="flex min-h-screen min-w-0 flex-1 flex-col">
          <header className="sticky top-0 z-20 border-b border-slate-200/80 bg-white/85 backdrop-blur">
            <div className="px-4 py-4 md:px-6 xl:px-10">
              <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <p className="text-[11px] uppercase tracking-[0.26em] text-slate-400">
                      {current.section}
                    </p>
                    <div className="mt-1 flex items-center gap-3">
                      <h2 className="text-2xl font-semibold tracking-tight text-slate-950">
                        {current.label}
                      </h2>
                      <span className="hidden rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-500 md:inline-flex">
                        Standard Ops Console
                      </span>
                    </div>
                    <p className="mt-1 text-sm text-slate-500">
                      {current.description}
                    </p>
                  </div>

                  <div className="flex items-center gap-2">
                    <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700">
                      管理权限已生效
                    </span>
                    <Link
                      href="/"
                      className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 hover:text-slate-950"
                    >
                      返回前台
                    </Link>
                  </div>
                </div>

                <div className="flex gap-2 overflow-x-auto pb-1 lg:hidden">
                  {navItems.map((item) => {
                    const Icon = item.icon;
                    const active = isActive(pathname, item.href);
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        className={cn(
                          "inline-flex shrink-0 items-center gap-2 rounded-full border px-3 py-2 text-sm font-medium transition-colors",
                          active
                            ? "border-sky-200 bg-sky-50 text-sky-700"
                            : "border-slate-200 bg-white text-slate-500",
                        )}
                      >
                        <Icon className="h-4 w-4" />
                        {item.label}
                      </Link>
                    );
                  })}
                </div>
              </div>
            </div>
          </header>

          <main className="flex-1 px-4 py-6 md:px-6 xl:px-10 xl:py-8">
            <div className="mx-auto w-full max-w-7xl">{children}</div>
          </main>
        </div>
      </div>
    </div>
  );
}
