"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function AdminPagination({
  page,
  totalPages,
  onPageChange,
}: {
  page: number;
  totalPages: number;
  onPageChange: (next: number) => void;
}) {
  if (totalPages <= 1) return null;

  return (
    <div className="flex flex-col gap-3 rounded-3xl border border-slate-200 bg-white px-4 py-4 shadow-[0_16px_40px_rgba(15,23,42,0.05)] sm:flex-row sm:items-center sm:justify-between">
      <p className="text-sm text-slate-500">
        第 <span className="font-medium text-slate-900">{page}</span> /{" "}
        <span className="font-medium text-slate-900">{totalPages}</span> 页
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.max(1, page - 1))}
          disabled={page <= 1}
        >
          <ChevronLeft className="mr-1 h-4 w-4" />
          上一页
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
          disabled={page >= totalPages}
        >
          下一页
          <ChevronRight className="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
