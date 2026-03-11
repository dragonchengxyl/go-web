'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { apiClient, Event, EventAttendee } from '@/lib/api-client';
import { Button } from '@/components/ui/button';

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [attendees, setAttendees] = useState<EventAttendee[]>([]);
  const [loading, setLoading] = useState(true);
  const [attending, setAttending] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    Promise.all([apiClient.getEvent(id), apiClient.listEventAttendees(id)])
      .then(([ev, atRes]) => {
        setEvent(ev);
        setAttendees(atRes.attendees ?? []);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [id]);

  const handleAttend = async () => {
    if (!id) return;
    setAttending(true);
    setError('');
    try {
      await apiClient.attendEvent(id);
      const updated = await apiClient.getEvent(id);
      setEvent(updated);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '报名失败');
    } finally {
      setAttending(false);
    }
  };

  if (loading) {
    return (
      <main className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="w-10 h-10 rounded-full border-2 border-purple-500 border-t-transparent animate-spin" />
      </main>
    );
  }

  if (!event) {
    return (
      <main className="min-h-screen bg-slate-950 flex items-center justify-center">
        <p className="text-white/50">活动不存在或已被删除</p>
      </main>
    );
  }

  const start = new Date(event.start_time);
  const end = new Date(event.end_time);

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-2xl">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-start justify-between gap-4">
            <h1 className="text-3xl font-bold text-white">{event.title}</h1>
            <span
              className={`shrink-0 text-xs px-2.5 py-1 rounded-full ${
                event.status === 'published'
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-white/10 text-white/50'
              }`}
            >
              {event.status === 'published' ? '报名中' : event.status}
            </span>
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
        </div>

        {/* Info card */}
        <div className="rounded-2xl border border-white/10 bg-white/5 p-6 space-y-3 mb-6">
          <div className="flex items-center gap-2 text-white/70">
            <span>🗓</span>
            <span>
              {start.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })}{' '}
              {start.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })} —{' '}
              {end.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
          {event.is_online ? (
            <div className="flex items-center gap-2 text-white/70">
              <span>🌐</span>
              <span>线上活动</span>
            </div>
          ) : (
            event.location && (
              <div className="flex items-center gap-2 text-white/70">
                <span>📍</span>
                <span>{event.location}</span>
              </div>
            )
          )}
          <div className="flex items-center gap-2 text-white/70">
            <span>👥</span>
            <span>
              {event.attendee_count} 人参与
              {event.max_capacity > 0 && ` / 上限 ${event.max_capacity}`}
            </span>
          </div>
        </div>

        {/* Description */}
        {event.description && (
          <div className="rounded-2xl border border-white/10 bg-white/5 p-6 mb-6">
            <h2 className="text-sm font-medium text-white/50 mb-2">活动详情</h2>
            <p className="text-white/80 whitespace-pre-wrap leading-relaxed">{event.description}</p>
          </div>
        )}

        {/* Attend button */}
        {event.status === 'published' && (
          <div className="mb-8">
            {error && <p className="text-red-400 text-sm mb-2">{error}</p>}
            <Button
              onClick={handleAttend}
              disabled={attending}
              className="w-full bg-purple-600 hover:bg-purple-500 text-white py-3 text-base"
            >
              {attending ? '提交中…' : '我要参加'}
            </Button>
          </div>
        )}

        {/* Attendees */}
        {attendees.length > 0 && (
          <div>
            <h2 className="text-lg font-semibold text-white mb-3">参与者（{attendees.length}）</h2>
            <div className="flex flex-wrap gap-2">
              {attendees.map((a) => (
                <div
                  key={a.user_id}
                  className="w-9 h-9 rounded-full bg-purple-600/40 flex items-center justify-center text-xs text-purple-200"
                  title={a.user_id}
                >
                  {a.user_id.slice(0, 2).toUpperCase()}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
