'use client';

import { useEffect, useRef, useState } from 'react';
import { useWS } from '@/contexts/ws-context';

interface ToastItem {
  id: number;
  message: string;
}

export function ModerationToast() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const { subscribe } = useWS();
  const nextIdRef = useRef(0);

  useEffect(() => {
    return subscribe('notification', (payload: unknown) => {
      const p = payload as any;
      if (p?.type === 'system' && p?.post_id && p?.status === 'approved') {
        const id = ++nextIdRef.current;
        setToasts(prev => [...prev, { id, message: '✓ 您的帖子已通过审核' }]);
        setTimeout(() => {
          setToasts(prev => prev.filter(t => t.id !== id));
        }, 4000);
      }
    });
  }, [subscribe]);

  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-2 pointer-events-none">
      {toasts.map(t => (
        <div
          key={t.id}
          className="bg-green-600 text-white text-sm px-4 py-2.5 rounded-lg shadow-lg animate-in slide-in-from-bottom-2 duration-300"
        >
          {t.message}
        </div>
      ))}
    </div>
  );
}
