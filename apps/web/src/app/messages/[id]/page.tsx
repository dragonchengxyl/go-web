'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams } from 'next/navigation';
import { apiClient, Message } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Send } from 'lucide-react';

export default function ConversationPage() {
  const params = useParams();
  const id = params.id as string;

  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(true);
  const bottomRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!id) return;

    apiClient.getMessages(id, 1, 50)
      .then((res) => setMessages((res.messages || []).reverse()))
      .catch(() => setMessages([]))
      .finally(() => setLoading(false));

    apiClient.markRead(id).catch(() => {});

    // WebSocket connection
    const ws = apiClient.connectWebSocket((msg) => {
      if (msg.type === 'chat' && msg.conversation_id === id) {
        setMessages((prev) => [...prev, msg.payload as Message]);
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
      }
    });
    if (ws) wsRef.current = ws;

    return () => {
      wsRef.current?.close();
    };
  }, [id]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || !id) return;
    const content = input;
    setInput('');
    try {
      const msg = await apiClient.sendMessage(id, content);
      setMessages((prev) => [...prev, msg]);
    } catch {
      setInput(content);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col h-screen pt-16">
        <div className="flex-1 flex items-center justify-center text-muted-foreground">
          加载中...
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen pt-16">
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map((msg) => (
          <div key={msg.id} className="flex gap-2 items-end">
            <div className="w-8 h-8 rounded-full bg-muted flex-shrink-0" />
            <div className="max-w-xs lg:max-w-md">
              <p className="text-xs text-muted-foreground mb-1">{msg.sender_username}</p>
              <div className="bg-card border rounded-2xl rounded-tl-none px-4 py-2 text-sm">
                {msg.content}
              </div>
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <div className="border-t p-4">
        <form onSubmit={handleSend} className="flex gap-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="输入消息..."
            className="flex-1"
          />
          <Button type="submit" size="icon">
            <Send className="h-4 w-4" />
          </Button>
        </form>
      </div>
    </div>
  );
}
