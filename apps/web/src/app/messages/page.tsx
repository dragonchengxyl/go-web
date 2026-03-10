'use client';

import { useEffect, useState } from 'react';
import { apiClient, Conversation, Message } from '@/lib/api-client';
import Link from 'next/link';
import { MessageCircle } from 'lucide-react';
import { useWS } from '@/contexts/ws-context';

export default function MessagesPage() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const { subscribe } = useWS();

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) apiClient.setToken(token);

    apiClient.getConversations(1, 50)
      .then((res) => setConversations(res.conversations || []))
      .catch(() => setConversations([]))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    return subscribe('chat', (payload: unknown) => {
      const msg = payload as Message & { conversation_id: string };
      setConversations(convs => convs.map(c =>
        c.id === msg.conversation_id
          ? { ...c, last_message: msg, unread_count: (c.unread_count ?? 0) + 1 }
          : c
      ));
    });
  }, [subscribe]);

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <h1 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <MessageCircle className="h-6 w-6" />
        消息
      </h1>
      {conversations.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">
          <p>暂无消息</p>
        </div>
      ) : (
        <div className="space-y-1">
          {conversations.map((conv) => (
            <Link
              key={conv.id}
              href={`/messages/${conv.id}`}
              className="flex items-center gap-3 p-4 rounded-lg hover:bg-muted transition-colors border"
            >
              <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                <MessageCircle className="h-5 w-5 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="font-medium truncate">
                  {conv.name || `会话 ${conv.id.slice(0, 8)}`}
                </p>
                {conv.last_message && (
                  <p className="text-sm text-muted-foreground truncate">
                    {conv.last_message.content}
                  </p>
                )}
              </div>
              {conv.unread_count ? (
                <span className="bg-primary text-primary-foreground text-xs rounded-full px-2 py-0.5">
                  {conv.unread_count}
                </span>
              ) : null}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
