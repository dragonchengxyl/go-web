'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { apiClient, Post } from '@/lib/api-client';
import { PostCard } from '@/components/post/post-card';

export default function PostDetailPage() {
  const params = useParams();
  const [post, setPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const id = params.id as string;
    if (id) {
      apiClient.getPost(id)
        .then(setPost)
        .catch(() => setPost(null))
        .finally(() => setLoading(false));
    }
  }, [params.id]);

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-60 bg-muted animate-pulse rounded-lg" />
      </div>
    );
  }

  if (!post) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center">
        <p className="text-muted-foreground">帖子不存在或已被删除</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-8">
      <PostCard
        post={post}
        showFull
        onLike={async () => {
          try {
            if (post.is_liked_by_me) {
              await apiClient.unlikePost(post.id);
            } else {
              await apiClient.likePost(post.id);
            }
            setPost((p) =>
              p
                ? {
                    ...p,
                    is_liked_by_me: !p.is_liked_by_me,
                    like_count: p.is_liked_by_me ? p.like_count - 1 : p.like_count + 1,
                  }
                : p
            );
          } catch {
            // ignore
          }
        }}
      />
    </div>
  );
}
