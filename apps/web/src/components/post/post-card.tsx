'use client';

import { useState, useRef, useEffect } from 'react';
import { Post, apiClient } from '@/lib/api-client';
import { Heart, MessageCircle, MoreHorizontal, Pin, Flag } from 'lucide-react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时前`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}天前`;
  return new Date(dateStr).toLocaleDateString('zh-CN');
}

const REPORT_REASONS = ['垃圾信息', '色情低俗', '违法内容', '侮辱谩骂', '欺诈诈骗', '其他'];

function ReportModal({ postId, onClose }: { postId: string; onClose: () => void }) {
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  async function submit() {
    if (!reason) return;
    setLoading(true);
    try {
      await apiClient.createReport('post', postId, reason);
      setDone(true);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-background rounded-2xl p-6 w-full max-w-sm mx-4" onClick={e => e.stopPropagation()}>
        {done ? (
          <div className="text-center py-4">
            <p className="text-lg font-bold mb-2">举报已提交</p>
            <p className="text-sm text-muted-foreground mb-4">感谢你的反馈，我们会尽快处理</p>
            <Button onClick={onClose}>关闭</Button>
          </div>
        ) : (
          <>
            <h3 className="font-bold mb-4 flex items-center gap-2">
              <Flag className="h-4 w-4 text-red-500" />
              举报内容
            </h3>
            <div className="space-y-2 mb-4">
              {REPORT_REASONS.map(r => (
                <button
                  key={r}
                  onClick={() => setReason(r)}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm border transition-colors ${reason === r ? 'bg-red-50 border-red-300 text-red-700 dark:bg-red-950/30 dark:border-red-700 dark:text-red-300' : 'hover:bg-muted'}`}
                >
                  {r}
                </button>
              ))}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" className="flex-1" onClick={onClose}>取消</Button>
              <Button className="flex-1 bg-red-500 hover:bg-red-600" onClick={submit} disabled={!reason || loading}>
                {loading ? '提交中...' : '提交举报'}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

interface PostCardProps {
  post: Post;
  onLike?: () => void;
  showFull?: boolean;
}

export function PostCard({ post, onLike, showFull = false }: PostCardProps) {
  const createdAgo = timeAgo(post.created_at);
  const [menuOpen, setMenuOpen] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  return (
    <div className="bg-card border rounded-xl p-4 hover:border-primary/30 transition-colors">
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <Link href={`/users/${post.author_id}`}>
            <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 hover:opacity-80 transition-opacity">
              <span className="text-sm font-semibold text-primary">
                {post.author_username?.[0]?.toUpperCase() || '?'}
              </span>
            </div>
          </Link>
          <div>
            <Link href={`/users/${post.author_id}`} className="font-semibold text-sm hover:text-primary transition-colors">
              {post.author_username || '未知用户'}
            </Link>
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <span>{createdAgo}</span>
              {post.is_pinned && (
                <>
                  <span>·</span>
                  <span className="flex items-center gap-0.5">
                    <Pin className="h-3 w-3" />
                    置顶
                  </span>
                </>
              )}
            </div>
          </div>
        </div>
        {/* More menu */}
        <div className="relative" ref={menuRef}>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setMenuOpen(v => !v)}>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
          {menuOpen && (
            <div className="absolute right-0 top-8 bg-background border rounded-lg shadow-lg py-1 z-10 w-32">
              <button
                onClick={() => { setMenuOpen(false); setShowReport(true); }}
                className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-muted transition-colors"
              >
                <Flag className="h-4 w-4" />
                举报
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Title */}
      {post.title && (
        <h2 className="font-bold text-lg mb-2">
          <Link href={`/posts/${post.id}`} className="hover:text-primary transition-colors">
            {post.title}
          </Link>
        </h2>
      )}

      {/* Content */}
      <div className="mb-3">
        <Link href={`/posts/${post.id}`}>
          <p className={`text-sm leading-relaxed whitespace-pre-wrap ${!showFull && 'line-clamp-5'}`}>
            {post.content}
          </p>
        </Link>
      </div>

      {/* Media */}
      {post.media_urls && post.media_urls.length > 0 && (
        <div className={`grid gap-2 mb-3 ${post.media_urls.length > 1 ? 'grid-cols-2' : 'grid-cols-1'}`}>
          {post.media_urls.slice(0, 4).map((url, i) => (
            <img key={i} src={url} alt="" className="w-full h-48 object-cover rounded-lg" />
          ))}
        </div>
      )}

      {/* Tags */}
      {post.tags && post.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3">
          {post.tags.map(tag => (
            <Badge key={tag} variant="secondary" className="text-xs">#{tag}</Badge>
          ))}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-4 pt-2 border-t">
        <button
          onClick={onLike}
          className={`flex items-center gap-1.5 text-sm transition-colors ${
            post.is_liked_by_me ? 'text-red-500 hover:text-red-600' : 'text-muted-foreground hover:text-red-500'
          }`}
        >
          <Heart className={`h-4 w-4 ${post.is_liked_by_me ? 'fill-current' : ''}`} />
          <span>{post.like_count}</span>
        </button>
        <Link href={`/posts/${post.id}`} className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-primary transition-colors">
          <MessageCircle className="h-4 w-4" />
          <span>{post.comment_count}</span>
        </Link>
      </div>

      {showReport && <ReportModal postId={post.id} onClose={() => setShowReport(false)} />}
    </div>
  );
}
