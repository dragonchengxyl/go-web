'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Flag, ExternalLink, ShieldCheck, Clock3, Ban } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface ReportRecord {
  id: string;
  target_type: 'post' | 'comment' | 'user';
  target_id: string;
  reason: string;
  description?: string;
  status: 'pending' | 'reviewed' | 'dismissed';
  action_taken?: 'block_post' | 'delete_comment' | 'ban_user';
  created_at: string;
  reviewed_at?: string;
}

const STATUS_TABS = [
  { value: '', label: '全部' },
  { value: 'pending', label: '处理中' },
  { value: 'reviewed', label: '已处理' },
  { value: 'dismissed', label: '已忽略' },
];

const STATUS_LABEL: Record<ReportRecord['status'], string> = {
  pending: '处理中',
  reviewed: '已处理',
  dismissed: '已忽略',
};

const ACTION_LABEL: Record<string, string> = {
  block_post: '帖子已封禁',
  delete_comment: '评论已删除',
  ban_user: '用户已封禁',
};

function targetLink(report: ReportRecord): string | null {
  if (report.target_type === 'post') return `/posts/${report.target_id}`;
  if (report.target_type === 'user') return `/users/${report.target_id}`;
  return null;
}

export default function MyReportsPage() {
  const [tab, setTab] = useState('');
  const [reports, setReports] = useState<ReportRecord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) apiClient.setToken(token);
    setLoading(true);
    apiClient.getMyReports(tab || undefined, 1, 50)
      .then((data) => setReports((data.reports ?? []) as ReportRecord[]))
      .catch(() => setReports([]))
      .finally(() => setLoading(false));
  }, [tab]);

  return (
    <div className="max-w-3xl mx-auto pt-20 px-4 pb-10">
      <div className="flex items-center justify-between mb-6 gap-4">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Flag className="h-6 w-6" />
            我的举报
          </h1>
          <p className="text-sm text-muted-foreground mt-1">查看你提交过的举报及处理进度</p>
        </div>
        <Link href="/settings">
          <Button variant="outline">返回设置</Button>
        </Link>
      </div>

      <div className="flex gap-2 flex-wrap mb-6">
        {STATUS_TABS.map((item) => (
          <button
            key={item.label}
            onClick={() => setTab(item.value)}
            className={cn(
              'px-3 py-1.5 rounded-full text-sm border transition-colors',
              tab === item.value
                ? 'bg-primary text-primary-foreground border-primary'
                : 'text-muted-foreground hover:border-primary/40 hover:text-foreground'
            )}
          >
            {item.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-28 rounded-xl bg-muted animate-pulse" />
          ))}
        </div>
      ) : reports.length === 0 ? (
        <div className="text-center py-20 text-muted-foreground">
          <Flag className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p>还没有举报记录</p>
        </div>
      ) : (
        <div className="space-y-3">
          {reports.map((report) => {
            const link = targetLink(report);
            return (
              <div key={report.id} className="rounded-xl border bg-card p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap mb-2">
                      <span className="text-xs px-2 py-0.5 rounded-full border">
                        {report.target_type === 'post' ? '帖子' : report.target_type === 'comment' ? '评论' : '用户'}
                      </span>
                      <span className="font-medium text-sm">{report.reason}</span>
                      <span className={cn(
                        'text-xs px-2 py-0.5 rounded-full',
                        report.status === 'pending'
                          ? 'bg-amber-500/10 text-amber-600'
                          : report.status === 'reviewed'
                            ? 'bg-green-500/10 text-green-600'
                            : 'bg-muted text-muted-foreground'
                      )}>
                        {STATUS_LABEL[report.status]}
                      </span>
                    </div>
                    {report.description && (
                      <p className="text-sm text-muted-foreground mb-2 whitespace-pre-wrap">{report.description}</p>
                    )}
                    <div className="flex items-center gap-3 flex-wrap text-xs text-muted-foreground">
                      <span className="inline-flex items-center gap-1">
                        <Clock3 className="h-3.5 w-3.5" />
                        提交于 {new Date(report.created_at).toLocaleString('zh-CN')}
                      </span>
                      {report.reviewed_at && (
                        <span>处理于 {new Date(report.reviewed_at).toLocaleString('zh-CN')}</span>
                      )}
                      {report.action_taken && (
                        <span className="inline-flex items-center gap-1 text-green-600 dark:text-green-400">
                          <ShieldCheck className="h-3.5 w-3.5" />
                          {ACTION_LABEL[report.action_taken] ?? report.action_taken}
                        </span>
                      )}
                      {report.status === 'dismissed' && (
                        <span className="inline-flex items-center gap-1">
                          <Ban className="h-3.5 w-3.5" />
                          未采取额外动作
                        </span>
                      )}
                    </div>
                  </div>

                  {link && (
                    <Link
                      href={link}
                      className="text-xs text-blue-500 hover:underline inline-flex items-center gap-1 shrink-0"
                    >
                      查看目标 <ExternalLink className="h-3 w-3" />
                    </Link>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
