'use client';

import Link from 'next/link';
import { Heart, MessageCircle } from 'lucide-react';
import { Post } from '@/lib/api-client';

const GRADIENTS = [
  'from-purple-500 to-teal-400',
  'from-teal-400 to-blue-500',
  'from-orange-400 to-pink-500',
  'from-blue-500 to-indigo-600',
  'from-green-400 to-teal-500',
  'from-pink-500 to-purple-600',
];

function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash * 31 + str.charCodeAt(i)) | 0;
  }
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

interface PostGalleryCardProps {
  post: Post;
}

export function PostGalleryCard({ post }: PostGalleryCardProps) {
  const displayName = post.author_username || post.author_id;
  const initial = (displayName[0] || '?').toUpperCase();
  const gradient = hashGradient(post.id);
  const authorGradient = hashGradient(post.author_id);
  const firstImage = post.media_urls?.[0];
  const hasImage = !!firstImage;

  return (
    <Link href={`/posts/${post.id}`} className="group block">
      <div className="rounded-xl overflow-hidden bg-card border border-border/50 hover:border-primary/30 transition-all duration-200 hover:shadow-lg hover:shadow-primary/5 hover:-translate-y-0.5">
        {/* Image area */}
        <div className="relative aspect-[4/3] overflow-hidden">
          {hasImage ? (
            <img
              src={firstImage}
              alt={post.title || ''}
              className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
            />
          ) : (
            <div className={`w-full h-full bg-gradient-to-br ${gradient} flex items-center justify-center p-4`}>
              <p className="text-white/90 text-sm text-center line-clamp-5 font-medium leading-relaxed">
                {post.content}
              </p>
            </div>
          )}
          {/* Hover overlay */}
          <div className="absolute inset-0 bg-black/0 group-hover:bg-black/25 transition-colors duration-200" />
          {/* Stats overlay (show on hover) */}
          <div className="absolute bottom-2 left-2 right-2 flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
            <span className="flex items-center gap-1 text-white text-xs bg-black/60 rounded-full px-2 py-0.5 backdrop-blur-sm">
              <Heart className="h-3 w-3" />
              {post.like_count}
            </span>
            <span className="flex items-center gap-1 text-white text-xs bg-black/60 rounded-full px-2 py-0.5 backdrop-blur-sm">
              <MessageCircle className="h-3 w-3" />
              {post.comment_count}
            </span>
          </div>
        </div>

        {/* Info bar */}
        <div className="px-3 py-2.5">
          {post.title ? (
            <p className="text-sm font-semibold line-clamp-1 mb-1.5">{post.title}</p>
          ) : (
            <p className="text-xs text-muted-foreground line-clamp-2 mb-1.5 leading-relaxed">{post.content}</p>
          )}
          <div className="flex items-center gap-2">
            <div
              className={`w-5 h-5 rounded-full bg-gradient-to-br ${authorGradient} flex items-center justify-center flex-shrink-0`}
            >
              <span className="text-white text-[9px] font-bold">{initial}</span>
            </div>
            <span className="text-xs text-muted-foreground truncate flex-1">{displayName}</span>
            {!hasImage && (
              <span className="flex items-center gap-1 text-muted-foreground text-xs flex-shrink-0">
                <Heart className="h-3 w-3" />
                {post.like_count}
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
