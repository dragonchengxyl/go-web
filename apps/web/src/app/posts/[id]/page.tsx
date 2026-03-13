'use client';

import { useEffect, useState, useRef, Suspense } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Heart, MessageCircle, Share2, Gift, Send, Flag, MapPin, Globe, UserPlus, UserMinus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { PostCard } from '@/components/post/post-card';
import { PostGalleryCard } from '@/components/post/post-gallery-card';
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

const GRADIENTS = [
  'from-purple-500 to-teal-400',
  'from-teal-400 to-blue-500',
  'from-orange-400 to-pink-500',
  'from-blue-500 to-indigo-600',
  'from-green-400 to-teal-500',
];
function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = (hash * 31 + str.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

const REPORT_REASONS = ['垃圾信息', '色情低俗', '违法内容', '侮辱谩骂', '欺诈诈骗', '其他'];

function ReportModal({ postId, onClose }: { postId: string; onClose: () => void }) {
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  async function handleSubmit() {
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
    <div className="fixed inset-0 bg-black/50 flex items-end sm:items-center justify-center z-50" onClick={onClose}>
      <div className="bg-background rounded-t-2xl sm:rounded-2xl p-6 w-full max-w-sm" onClick={e => e.stopPropagation()}>
        {done ? (
          <div className="text-center py-4">
            <p className="text-green-500 font-semibold mb-2">举报已提交</p>
            <p className="text-sm text-muted-foreground mb-4">感谢你的反馈，我们会尽快处理</p>
            <Button onClick={onClose} className="w-full">关闭</Button>
          </div>
        ) : (
          <>
            <h3 className="text-lg font-bold mb-4 flex items-center gap-2">
              <Flag className="h-5 w-5 text-destructive" />
              举报帖子
            </h3>
            <div className="flex flex-wrap gap-2 mb-4">
              {REPORT_REASONS.map(r => (
                <button
                  key={r}
                  onClick={() => setReason(r)}
                  className={`px-3 py-1.5 rounded-full text-sm border transition-colors ${reason === r ? 'bg-destructive text-destructive-foreground border-destructive' : 'hover:border-destructive hover:text-destructive'}`}
                >
                  {r}
                </button>
              ))}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" className="flex-1" onClick={onClose}>取消</Button>
              <Button variant="destructive" className="flex-1" onClick={handleSubmit} disabled={!reason || loading}>
                {loading ? '提交中…' : '提交举报'}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
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

interface AuthorProfile {
  id: string
  username: string
  furry_name?: string
  species?: string
  bio?: string
  location?: string
  website?: string
  avatar_key?: string
}

function AuthorCard({ authorId, currentUserId }: { authorId: string; currentUserId: string | null }) {
  const [author, setAuthor] = useState<AuthorProfile | null>(null);
  const [isFollowing, setIsFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);

  useEffect(() => {
    if (!authorId) return;
    apiClient.getUser(authorId).then(u => setAuthor(u)).catch(() => {});
    if (apiClient.getToken() && currentUserId && currentUserId !== authorId) {
      apiClient.getFollowers(authorId, 1, 1000).then(res => {
        setIsFollowing(res.followers?.some(f => f.follower_id === currentUserId) ?? false);
      }).catch(() => {});
    }
  }, [authorId, currentUserId]);

  async function handleFollow() {
    if (!apiClient.getToken()) return;
    setFollowLoading(true);
    try {
      if (isFollowing) {
        await apiClient.unfollowUser(authorId);
        setIsFollowing(false);
      } else {
        await apiClient.followUser(authorId);
        setIsFollowing(true);
      }
    } catch {
      // ignore
    } finally {
      setFollowLoading(false);
    }
  }

  if (!author) return null;

  const displayName = author.furry_name || author.username;
  const gradient = hashGradient(author.id);
  const isSelf = currentUserId === authorId;

  return (
    <div className="bg-card border rounded-xl p-5 mt-6">
      <div className="flex items-start gap-3">
        <Link href={`/users/${author.id}`} className="flex-shrink-0">
          <div className={`w-12 h-12 rounded-full bg-gradient-to-br ${gradient} flex items-center justify-center text-white font-bold hover:brightness-110 transition-all`}>
            {displayName[0]?.toUpperCase()}
          </div>
        </Link>
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <Link href={`/users/${author.id}`} className="font-semibold hover:text-primary transition-colors">
                {displayName}
              </Link>
              {author.furry_name && (
                <p className="text-xs text-muted-foreground">@{author.username}</p>
              )}
              {author.species && (
                <p className="text-xs text-primary">🐾 {author.species}</p>
              )}
            </div>
            {!isSelf && apiClient.getToken() && (
              <Button
                variant={isFollowing ? 'outline' : 'default'}
                size="sm"
                onClick={handleFollow}
                disabled={followLoading}
                className={isFollowing ? '' : 'bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110 flex-shrink-0'}
              >
                {isFollowing
                  ? <><UserMinus className="h-3.5 w-3.5 mr-1" />取消关注</>
                  : <><UserPlus className="h-3.5 w-3.5 mr-1" />关注 TA</>
                }
              </Button>
            )}
          </div>
          {author.bio && (
            <p className="text-sm text-muted-foreground mt-1.5 line-clamp-2">{author.bio}</p>
          )}
          <div className="flex gap-3 mt-1.5 text-xs text-muted-foreground">
            {author.location && (
              <span className="flex items-center gap-1"><MapPin className="h-3 w-3" />{author.location}</span>
            )}
            {author.website && (
              <a href={author.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-primary">
                <Globe className="h-3 w-3" />{author.website}
              </a>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function SidebarMorePosts({ authorId, currentPostId }: { authorId: string; currentPostId: string }) {
  const [posts, setPosts] = useState<Post[]>([]);

  useEffect(() => {
    if (!authorId) return;
    apiClient.getUserPosts(authorId, 1, 7).then(res => {
      setPosts((res.posts ?? []).filter(p => p.id !== currentPostId).slice(0, 6));
    }).catch(() => {});
  }, [authorId, currentPostId]);

  if (posts.length === 0) return null;

  return (
    <div className="mt-8">
      <h3 className="font-semibold text-sm text-muted-foreground mb-4">TA 的其他创作</h3>
      <div className="grid grid-cols-2 gap-3">
        {posts.map(p => <PostGalleryCard key={p.id} post={p} />)}
      </div>
    </div>
  );
}

function PostDetailContent() {
  const params = useParams();
  const searchParams = useSearchParams();
  const justSubmitted = searchParams.get('submitted') === '1';
  const postId = params.id as string;
  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentTotal, setCommentTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [commentText, setCommentText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [showTip, setShowTip] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const commentInputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      apiClient.setToken(token);
      setIsLoggedIn(true);
      apiClient.getMe().then(u => setCurrentUserId(u.id)).catch(() => {});
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
        <p className="text-xl font-medium mb-2">帖子不存在</p>
        <p className="text-sm">该帖子可能已被删除或不存在</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      {/* Post submitted success banner */}
      {justSubmitted && (
        <div className="mb-4 flex items-start gap-2.5 px-4 py-3 rounded-xl bg-green-500/10 border border-green-500/30 text-green-700 dark:text-green-400 text-sm">
          <span className="mt-0.5 text-base leading-none">✅</span>
          <p>帖子已成功提交！正在等待审核，审核通过后将对所有用户可见。</p>
        </div>
      )}

      {/* Moderation pending banner */}
      {post.moderation_status === 'pending' && currentUserId === post.author_id && !justSubmitted && (
        <div className="mb-4 flex items-start gap-2.5 px-4 py-3 rounded-xl bg-yellow-500/10 border border-yellow-500/30 text-yellow-700 dark:text-yellow-400 text-sm">
          <span className="mt-0.5 text-base leading-none">⏳</span>
          <p>您的帖子正在审核中，审核通过后将对其他用户可见。通常在 24 小时内完成。</p>
        </div>
      )}

      {post.moderation_status === 'blocked' && currentUserId === post.author_id && (
        <div className="mb-4 flex items-start gap-2.5 px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-700 dark:text-red-400 text-sm">
          <span className="mt-0.5 text-base leading-none">🚫</span>
          <p>这条帖子未通过社区审核，目前仅自己可见。你可以修改后重新发布。</p>
        </div>
      )}

      {/* Post */}
      <PostCard post={post} showFull />

      {/* Action bar */}
      <div className="flex items-center gap-1 mt-3 mb-6">
        <Button variant="ghost" size="sm" onClick={() => commentInputRef.current?.focus()} className="text-muted-foreground">
          <MessageCircle className="h-4 w-4 mr-1" />
          评论 ({commentTotal})
        </Button>
        <Button variant="ghost" size="sm" onClick={handleShare} className="text-muted-foreground">
          <Share2 className="h-4 w-4 mr-1" />
          分享
        </Button>
        {isLoggedIn && post.author_id && currentUserId !== post.author_id && (
          <Button variant="ghost" size="sm" onClick={() => setShowReport(true)} className="text-muted-foreground hover:text-destructive">
            <Flag className="h-4 w-4 mr-1" />
            举报
          </Button>
        )}
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
        {comments.length === 0 && !isLoggedIn && (
          <div className="text-center py-8 text-muted-foreground">
            <MessageCircle className="h-10 w-10 mx-auto mb-3 opacity-30" />
            <p className="text-sm">
              <Link href="/login" className="text-primary hover:underline">登录</Link>后参与评论
            </p>
          </div>
        )}
        {comments.length === 0 && isLoggedIn && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">还没有评论，来抢沙发吧</p>
          </div>
        )}
        {comments.map(c => (
          <div key={c.id} className="flex gap-3">
            <Link href={`/users/${c.user_id}`}>
              <div className={`w-8 h-8 rounded-full bg-gradient-to-br ${hashGradient(c.user_id)} flex items-center justify-center flex-shrink-0 hover:brightness-110 transition-all`}>
                <span className="text-xs font-bold text-white">
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
      </div>

      {/* Author card */}
      {post.author_id && (
        <AuthorCard authorId={post.author_id} currentUserId={currentUserId} />
      )}

      {/* More from author */}
      {post.author_id && (
        <SidebarMorePosts authorId={post.author_id} currentPostId={post.id} />
      )}

      {/* Tip modal */}
      {showTip && (
        <TipModal toUserId={post.author_id} onClose={() => setShowTip(false)} />
      )}

      {/* Report modal */}
      {showReport && (
        <ReportModal postId={post.id} onClose={() => setShowReport(false)} />
      )}
    </div>
  );
}

export default function PostDetailPage() {
  return (
    <Suspense fallback={<div className="max-w-2xl mx-auto pt-20 px-4"><div className="h-64 bg-muted animate-pulse rounded-xl" /></div>}>
      <PostDetailContent />
    </Suspense>
  );
}
