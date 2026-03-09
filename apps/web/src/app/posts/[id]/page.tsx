'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Heart, MessageCircle, Share2, Gift, Send } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { PostCard } from '@/components/post/post-card';
import { apiClient, Post, Comment } from '@/lib/api-client';

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

function TipModal({ toUserId, onClose }: { toUserId: string; onClose: () => void }) {
  const [amount, setAmount] = useState(10);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const presets = [1, 5, 10, 20, 50];

  async function handleTip() {
    setLoading(true);
    try {
      const order = await apiClient.createTip(toUserId, amount, message);
      const { pay_url } = await apiClient.payTipAlipay(order.id, window.location.href);
      window.open(pay_url, '_blank');
      onClose();
    } catch (e: any) {
      alert(e.message || '打赏失败');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-end sm:items-center justify-center z-50" onClick={onClose}>
      <div className="bg-background rounded-t-2xl sm:rounded-2xl p-6 w-full max-w-sm" onClick={e => e.stopPropagation()}>
        <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
          <Gift className="h-5 w-5 text-yellow-500" />
          打赏创作者
        </h3>
        <div className="flex gap-2 mb-4 flex-wrap">
          {presets.map(p => (
            <button
              key={p}
              onClick={() => setAmount(p)}
              className={`px-3 py-1.5 rounded-full text-sm border transition-colors ${amount === p ? 'bg-primary text-primary-foreground border-primary' : 'hover:border-primary'}`}
            >
              ¥{p}
            </button>
          ))}
        </div>
        <input
          type="number"
          min={0.01}
          step={0.01}
          value={amount}
          onChange={e => setAmount(Number(e.target.value))}
          className="w-full border rounded-md px-3 py-2 text-sm mb-3"
          placeholder="自定义金额"
        />
        <Textarea
          value={message}
          onChange={e => setMessage(e.target.value)}
          placeholder="留言给创作者（可选）"
          className="mb-4 text-sm"
          rows={2}
        />
        <div className="flex gap-2">
          <Button variant="outline" className="flex-1" onClick={onClose}>取消</Button>
          <Button className="flex-1" onClick={handleTip} disabled={loading || amount <= 0}>
            {loading ? '处理中…' : `打赏 ¥${amount}`}
          </Button>
        </div>
      </div>
    </div>
  );
}

export default function PostDetailPage() {
  const params = useParams();
  const postId = params.id as string;
  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentTotal, setCommentTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [commentText, setCommentText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [showTip, setShowTip] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const commentInputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      apiClient.setToken(token);
      setIsLoggedIn(true);
    }
    if (postId) loadData(postId);
  }, [postId]);

  async function loadData(id: string) {
    try {
      const [postData, commentsData] = await Promise.all([
        apiClient.getPost(id),
        apiClient.getComments(id, 1, 50),
      ]);
      setPost(postData);
      setComments(commentsData.comments ?? []);
      setCommentTotal(commentsData.total);
    } catch {
      setPost(null);
    } finally {
      setLoading(false);
    }
  }

  async function handleLike() {
    if (!isLoggedIn || !post) return;
    try {
      if (post.is_liked_by_me) {
        await apiClient.unlikePost(post.id);
      } else {
        await apiClient.likePost(post.id);
      }
      setPost(p => p ? {
        ...p,
        is_liked_by_me: !p.is_liked_by_me,
        like_count: p.is_liked_by_me ? p.like_count - 1 : p.like_count + 1,
      } : p);
    } catch { /* ignore */ }
  }

  async function handleComment() {
    if (!commentText.trim() || !postId) return;
    setSubmitting(true);
    try {
      const newComment = await apiClient.createComment(postId, commentText.trim());
      setComments(prev => [...prev, newComment]);
      setCommentTotal(t => t + 1);
      setCommentText('');
      setPost(p => p ? { ...p, comment_count: p.comment_count + 1 } : p);
    } catch (e: any) {
      alert(e.message || '评论失败');
    } finally {
      setSubmitting(false);
    }
  }

  function handleShare() {
    navigator.clipboard.writeText(window.location.href).then(() => {
      alert('链接已复制到剪贴板');
    }).catch(() => {
      alert(window.location.href);
    });
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-60 bg-muted animate-pulse rounded-xl mb-4" />
        <div className="h-40 bg-muted animate-pulse rounded-xl" />
      </div>
    );
  }

  if (!post) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        帖子不存在或已被删除
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      {/* Post */}
      <PostCard post={post} showFull onLike={handleLike} />

      {/* Action bar */}
      <div className="flex items-center gap-2 mt-3 mb-6">
        <Button variant="ghost" size="sm" onClick={() => commentInputRef.current?.focus()} className="text-muted-foreground">
          <MessageCircle className="h-4 w-4 mr-1" />
          评论 ({commentTotal})
        </Button>
        <Button variant="ghost" size="sm" onClick={handleShare} className="text-muted-foreground">
          <Share2 className="h-4 w-4 mr-1" />
          分享
        </Button>
        {isLoggedIn && post.author_id && (
          <Button variant="ghost" size="sm" onClick={() => setShowTip(true)} className="text-yellow-600 hover:text-yellow-700 ml-auto">
            <Gift className="h-4 w-4 mr-1" />
            打赏
          </Button>
        )}
      </div>

      {/* Comment input */}
      {isLoggedIn && (
        <div className="flex gap-2 mb-6">
          <Textarea
            ref={commentInputRef}
            value={commentText}
            onChange={e => setCommentText(e.target.value)}
            placeholder="写下你的评论…"
            rows={2}
            className="flex-1 text-sm resize-none"
            onKeyDown={e => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handleComment();
            }}
          />
          <Button onClick={handleComment} disabled={submitting || !commentText.trim()} size="icon" className="self-end">
            <Send className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Comments */}
      <div className="space-y-4">
        <h3 className="font-semibold text-sm text-muted-foreground">
          {commentTotal > 0 ? `${commentTotal} 条评论` : '暂无评论'}
        </h3>
        {comments.map(c => (
          <div key={c.id} className="flex gap-3">
            <Link href={`/users/${c.user_id}`}>
              <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 hover:opacity-80 transition-opacity">
                <span className="text-xs font-bold text-primary">
                  {(c.author_username || c.user_id)[0]?.toUpperCase()}
                </span>
              </div>
            </Link>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-0.5">
                <Link href={`/users/${c.user_id}`} className="text-sm font-semibold hover:text-primary transition-colors">
                  {c.author_username || c.user_id}
                </Link>
                <span className="text-xs text-muted-foreground">{timeAgo(c.created_at)}</span>
              </div>
              <p className="text-sm leading-relaxed whitespace-pre-wrap">{c.content}</p>
              <div className="flex items-center gap-3 mt-1.5">
                <button className="flex items-center gap-1 text-xs text-muted-foreground hover:text-red-500 transition-colors">
                  <Heart className="h-3.5 w-3.5" />
                  {c.like_count > 0 && c.like_count}
                </button>
              </div>
            </div>
          </div>
        ))}

        {!isLoggedIn && commentTotal === 0 && (
          <p className="text-sm text-muted-foreground text-center py-4">
            <Link href="/login" className="text-primary hover:underline">登录</Link>后参与评论
          </p>
        )}
      </div>

      {/* Tip modal */}
      {showTip && (
        <TipModal toUserId={post.author_id} onClose={() => setShowTip(false)} />
      )}
    </div>
  );
}
