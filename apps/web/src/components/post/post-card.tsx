'use client';

import { Post } from '@/lib/api-client';
import { Heart, MessageCircle, MoreHorizontal, Pin } from 'lucide-react';
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

interface PostCardProps {
  post: Post;
  onLike?: () => void;
  showFull?: boolean;
}

export function PostCard({ post, onLike, showFull = false }: PostCardProps) {
  const createdAgo = timeAgo(post.created_at);

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
            <Link
              href={`/users/${post.author_id}`}
              className="font-semibold text-sm hover:text-primary transition-colors"
            >
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
        <Button variant="ghost" size="icon" className="h-8 w-8">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
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
            <img
              key={i}
              src={url}
              alt=""
              className="w-full h-48 object-cover rounded-lg"
            />
          ))}
        </div>
      )}

      {/* Tags */}
      {post.tags && post.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3">
          {post.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              #{tag}
            </Badge>
          ))}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-4 pt-2 border-t">
        <button
          onClick={onLike}
          className={`flex items-center gap-1.5 text-sm transition-colors ${
            post.is_liked_by_me
              ? 'text-red-500 hover:text-red-600'
              : 'text-muted-foreground hover:text-red-500'
          }`}
        >
          <Heart className={`h-4 w-4 ${post.is_liked_by_me ? 'fill-current' : ''}`} />
          <span>{post.like_count}</span>
        </button>
        <Link
          href={`/posts/${post.id}`}
          className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-primary transition-colors"
        >
          <MessageCircle className="h-4 w-4" />
          <span>{post.comment_count}</span>
        </Link>
      </div>
    </div>
  );
}
