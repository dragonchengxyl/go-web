"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { ChevronDown, PanelTop, type LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AdminWorkspaceSwitcherItem {
  href: string;
  label: string;
  description: string;
  section: string;
  icon: React.ComponentType<{ className?: string }>;
}

function isActive(pathname: string, href: string) {
  return href === "/admin" ? pathname === href : pathname.startsWith(href);
}

export function getCurrentWorkspaceItem(
  pathname: string,
  items: AdminWorkspaceSwitcherItem[],
) {
  return items.find((item) => isActive(pathname, item.href)) ?? items[0];
}

function WorkspaceRow({
  item,
  active,
  onSelect,
}: {
  item: AdminWorkspaceSwitcherItem;
  active: boolean;
  onSelect: () => void;
}) {
  const Icon = item.icon;

  return (
    <Link
      href={item.href}
      onClick={onSelect}
      className={cn(
        "flex items-start gap-3 rounded-2xl px-3 py-3 transition-colors",
        active
          ? "bg-[#effaf5] text-[#17342d]"
          : "text-[#52776d] hover:bg-[#f3fbf7] hover:text-[#17342d]",
      )}
    >
      <span
        className={cn(
          "mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border",
          active
            ? "border-[#c9e5d7] bg-[#dcf3ea] text-[#1f7a73]"
            : "border-[#d7e9df] bg-white text-[#64897e]",
        )}
      >
        <Icon className="h-4 w-4" />
      </span>
      <span className="min-w-0">
        <span className="block text-sm font-medium">{item.label}</span>
        <span className="mt-1 block text-xs leading-5 opacity-90">
          {item.description}
        </span>
      </span>
    </Link>
  );
}

export function AdminWorkspaceSwitcher({
  pathname,
  items,
  variant = "header",
  className,
}: {
  pathname: string;
  items: AdminWorkspaceSwitcherItem[];
  variant?: "header" | "sidebar";
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const current = useMemo(
    () => getCurrentWorkspaceItem(pathname, items),
    [items, pathname],
  );

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const grouped = useMemo(() => {
    const result = new Map<string, AdminWorkspaceSwitcherItem[]>();
    for (const item of items) {
      if (!result.has(item.section)) {
        result.set(item.section, []);
      }
      result.get(item.section)?.push(item);
    }
    return Array.from(result.entries());
  }, [items]);

  if (variant === "sidebar") {
    return (
      <div ref={ref} className={cn("relative", className)}>
        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          className="flex w-full items-center gap-3 rounded-[22px] border border-white/10 bg-white/5 px-3 py-2.5 text-left text-[#f0faf5] transition-colors hover:bg-white/8"
        >
          <span className="h-10 w-1.5 shrink-0 rounded-full bg-gradient-to-b from-[#73d4be] via-[#f4b45f] to-[#21584e]" />
          <span className="min-w-0 flex-1">
            <span className="block text-[10px] uppercase tracking-[0.28em] text-[#a4d7c7]">
              Workspace
            </span>
            <span className="mt-1 block truncate text-sm font-semibold">
              {current.label}
            </span>
          </span>
          <ChevronDown
            className={cn(
              "h-4 w-4 shrink-0 text-[#d1e8df] transition-transform",
              open && "rotate-180",
            )}
          />
        </button>

        {open ? (
          <div className="absolute left-0 right-0 top-[calc(100%+0.75rem)] z-30 overflow-hidden rounded-[24px] border border-[#2c6a5d] bg-[#f9fffb] shadow-[0_24px_70px_rgba(14,46,38,0.26)]">
            <div className="max-h-[70vh] overflow-y-auto p-3">
              {grouped.map(([section, sectionItems]) => (
                <div key={section} className="pb-3 last:pb-0">
                  <p className="px-3 py-2 text-[10px] uppercase tracking-[0.26em] text-[#6f9d90]">
                    {section}
                  </p>
                  <div className="space-y-1">
                    {sectionItems.map((item) => (
                      <WorkspaceRow
                        key={item.href}
                        item={item}
                        active={isActive(pathname, item.href)}
                        onSelect={() => setOpen(false)}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : null}
      </div>
    );
  }

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="inline-flex items-center gap-3 rounded-full border border-slate-200 bg-white px-4 py-2 text-left shadow-sm transition-colors hover:bg-slate-50"
      >
        <span className="flex h-9 w-9 items-center justify-center rounded-full bg-[#dcf3ea] text-[#1f7a73]">
          <PanelTop className="h-4 w-4" />
        </span>
        <span className="min-w-0">
          <span className="block text-[10px] uppercase tracking-[0.24em] text-slate-400">
            当前工作区
          </span>
          <span className="block max-w-[15rem] truncate text-sm font-semibold text-slate-950">
            {current.label}
          </span>
        </span>
        <ChevronDown
          className={cn(
            "h-4 w-4 text-slate-400 transition-transform",
            open && "rotate-180",
          )}
        />
      </button>

      {open ? (
        <div className="absolute right-0 top-[calc(100%+0.75rem)] z-30 w-[26rem] max-w-[calc(100vw-2rem)] overflow-hidden rounded-[26px] border border-[#d3e7dc] bg-[#fbfffc] shadow-[0_28px_80px_rgba(27,67,55,0.16)]">
          <div className="max-h-[75vh] overflow-y-auto p-3">
            {grouped.map(([section, sectionItems]) => (
              <div key={section} className="pb-3 last:pb-0">
                <p className="px-3 py-2 text-[10px] uppercase tracking-[0.24em] text-[#83a399]">
                  {section}
                </p>
                <div className="space-y-1">
                  {sectionItems.map((item) => (
                    <WorkspaceRow
                      key={item.href}
                      item={item}
                      active={isActive(pathname, item.href)}
                      onSelect={() => setOpen(false)}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
