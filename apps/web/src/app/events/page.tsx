'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient, Event } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

const STATUS_LABELS: Record<string, string> = {
  published: '报名中',
  cancelled: '已取消',
  completed: '已结束',
  draft: '草稿',
};

function EventCard({ event }: { event: Event }) {
  const start = new Date(event.start_time);
  return (
    <Link href={`/events/${event.id}`}>
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        className="group rounded-2xl border border-white/10 bg-white/5 p-5 hover:bg-white/10 transition-colors cursor-pointer"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-white truncate group-hover:text-purple-300 transition-colors">
              {event.title}
            </h3>
            <p className="mt-1 text-sm text-white/60 line-clamp-2">{event.description}</p>
          </div>
          <span
            className={cn(
              'shrink-0 text-xs px-2 py-0.5 rounded-full',
              event.status === 'published' ? 'bg-green-500/20 text-green-400' : 'bg-white/10 text-white/50'
            )}
          >
            {STATUS_LABELS[event.status] ?? event.status}
          </span>
        </div>

        <div className="mt-3 flex flex-wrap gap-3 text-xs text-white/50">
          <span>
            🗓{' '}
            {start.toLocaleDateString('zh-CN', { month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
          </span>
          {event.is_online ? (
            <span>🌐 线上活动</span>
          ) : (
            event.location && <span>📍 {event.location}</span>
          )}
          <span>👥 {event.attendee_count} 人参与</span>
          {event.max_capacity > 0 && <span>/ 上限 {event.max_capacity}</span>}
        </div>

        {event.tags?.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {event.tags.map((t) => (
              <span key={t} className="text-xs px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300">
                #{t}
              </span>
            ))}
          </div>
        )}
      </motion.div>
    </Link>
  );
}

export default function EventsPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 12;

  useEffect(() => {
    setLoading(true);
    apiClient
      .listEvents(page, pageSize)
      .then((res) => {
        setEvents(res.events ?? []);
        setTotal(res.total ?? 0);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [page]);

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-4xl">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white">活动广场</h1>
            <p className="mt-1 text-white/50">发现 Furry 圈的线下/线上聚会</p>
          </div>
          <Link href="/events/create">
            <Button className="bg-purple-600 hover:bg-purple-500 text-white">+ 发起活动</Button>
          </Link>
        </div>

        {loading ? (
          <div className="grid gap-4 sm:grid-cols-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="h-40 rounded-2xl bg-white/5 animate-pulse" />
            ))}
          </div>
        ) : events.length === 0 ? (
          <div className="text-center py-20 text-white/40">暂无活动，快来发起第一个吧！</div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2">
            {events.map((e) => (
              <EventCard key={e.id} event={e} />
            ))}
          </div>
        )}

        {total > pageSize && (
          <div className="mt-8 flex justify-center gap-3">
            <Button
              variant="outline"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="border-white/20 text-white/70"
            >
              上一页
            </Button>
            <span className="flex items-center text-sm text-white/50">
              {page} / {Math.ceil(total / pageSize)}
            </span>
            <Button
              variant="outline"
              onClick={() => setPage((p) => p + 1)}
              disabled={page * pageSize >= total}
              className="border-white/20 text-white/70"
            >
              下一页
            </Button>
          </div>
        )}
      </div>
    </main>
  );
}
