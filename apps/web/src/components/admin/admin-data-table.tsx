"use client";

import { Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AdminColumn<T> {
  key: string;
  header: React.ReactNode;
  className?: string;
  headerClassName?: string;
  render: (row: T) => React.ReactNode;
}

export function AdminDataTable<T>({
  data,
  columns,
  keyExtractor,
  loading,
  empty,
}: {
  data: T[];
  columns: AdminColumn<T>[];
  keyExtractor: (row: T) => string;
  loading?: boolean;
  empty?: React.ReactNode;
}) {
  if (loading) {
    return (
      <div className="rounded-3xl border border-slate-200 bg-white p-10 shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
        <div className="flex items-center justify-center gap-3 text-sm text-slate-500">
          <Loader2 className="h-4 w-4 animate-spin" />
          正在加载数据...
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return empty ?? null;
  }

  return (
    <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-[0_16px_40px_rgba(15,23,42,0.05)]">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200">
          <thead className="bg-slate-50/90">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.key}
                  className={cn(
                    "px-4 py-3 text-left text-xs font-semibold uppercase tracking-[0.18em] text-slate-400",
                    column.headerClassName,
                  )}
                >
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200 bg-white">
            {data.map((row) => (
              <tr
                key={keyExtractor(row)}
                className="align-top transition-colors hover:bg-slate-50/70"
              >
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={cn("px-4 py-4 text-sm text-slate-700", column.className)}
                  >
                    {column.render(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
