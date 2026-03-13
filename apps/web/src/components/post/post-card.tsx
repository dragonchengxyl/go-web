"use client";

import { useState, useRef, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Post, apiClient } from "@/lib/api-client";
import {
  Heart,
  MessageCircle,
  MoreHorizontal,
  Pin,
  Flag,
  Share2,
  Bookmark,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes}分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时前`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}天前`;
  return new Date(dateStr).toLocaleDateString("zh-CN");
}

const REPORT_REASONS = [
  "垃圾信息",
  "色情低俗",
  "违法内容",
  "侮辱谩骂",
  "欺诈诈骗",
  "其他",
];

function ReportModal({
  postId,
  onClose,
}: {
  postId: string;
  onClose: () => void;
}) {
  const [reason, setReason] = useState("");
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  async function submit() {
    if (!reason) return;
    setLoading(true);
    try {
      await apiClient.createReport("post", postId, reason);
      setDone(true);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-background rounded-2xl p-6 w-full max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        {done ? (
          <div className="text-center py-4">
            <p className="text-lg font-bold mb-2">举报已提交</p>
            <p className="text-sm text-muted-foreground mb-4">
              感谢你的反馈，我们会尽快处理
            </p>
            <Button onClick={onClose}>关闭</Button>
          </div>
        ) : (
          <>
            <h3 className="font-bold mb-4 flex items-center gap-2">
              <Flag className="h-4 w-4 text-red-500" />
              举报内容
            </h3>
            <div className="space-y-2 mb-4">
              {REPORT_REASONS.map((r) => (
                <button
                  key={r}
                  onClick={() => setReason(r)}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm border transition-colors ${reason === r ? "bg-red-50 border-red-300 text-red-700 dark:bg-red-950/30 dark:border-red-700 dark:text-red-300" : "hover:bg-muted"}`}
                >
                  {r}
                </button>
              ))}
            </div>
            <div className="flex gap-2">
              <Button variant="outline" className="flex-1" onClick={onClose}>
                取消
              </Button>
              <Button
                className="flex-1 bg-red-500 hover:bg-red-600"
                onClick={submit}
                disabled={!reason || loading}
              >
                {loading ? "提交中..." : "提交举报"}
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
  showFull?: boolean;
}

export function PostCard({ post, showFull = false }: PostCardProps) {
  const queryClient = useQueryClient();
  const createdAgo = timeAgo(post.created_at);
  const [menuOpen, setMenuOpen] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const [liked, setLiked] = useState(post.is_liked_by_me ?? false);
  const [likeCount, setLikeCount] = useState(post.like_count);
  const [bookmarked, setBookmarked] = useState(
    post.is_bookmarked_by_me ?? false,
  );
  const menuRef = useRef<HTMLDivElement>(null);

  const isPending = post.moderation_status === "pending";
  const isBlocked = post.moderation_status === "blocked";

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const likeMutation = useMutation({
    mutationFn: () =>
      liked ? apiClient.unlikePost(post.id) : apiClient.likePost(post.id),
    onMutate: () => {
      const prevLiked = liked;
      const prevCount = likeCount;
      if (!liked) {
        setLiked(true);
        setLikeCount((c) => c + 1);
      } else {
        setLiked(false);
        setLikeCount((c) => c - 1);
      }
      return { prevLiked, prevCount };
    },
    onError: (_err, _vars, context) => {
      if (context) {
        setLiked(context.prevLiked);
        setLikeCount(context.prevCount);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["post", post.id] });
    },
  });

  const bookmarkMutation = useMutation({
    mutationFn: () =>
      bookmarked
        ? apiClient.unbookmarkPost(post.id)
        : apiClient.bookmarkPost(post.id),
    onMutate: () => {
      const prev = bookmarked;
      setBookmarked(!prev);
      return { prev };
    },
    onError: (_err, _vars, context) => {
      if (context) setBookmarked(context.prev);
    },
  });

  const mediaUrls = post.media_urls?.slice(0, 4) ?? [];

  return (
    <div className="relative bg-card border rounded-xl p-4 hover:-translate-y-0.5 hover:shadow-md transition-all duration-200">
      {/* Moderation overlays */}
      {isPending && (
        <div className="absolute inset-0 bg-gray-100/60 dark:bg-gray-900/60 rounded-xl flex items-center justify-center z-10 pointer-events-none">
          <span className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300 text-xs px-2 py-1 rounded-full">
            ⏳ 审核中
          </span>
        </div>
      )}
      {isBlocked && (
        <div className="absolute inset-0 bg-gray-200/80 dark:bg-gray-900/80 rounded-xl flex items-center justify-center z-10 pointer-events-none">
          <span className="bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300 text-xs px-3 py-1.5 rounded-full">
            内容不符合社区规范
          </span>
        </div>
      )}

      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <Link href={`/users/${post.author_id}`}>
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-purple to-brand-teal flex items-center justify-center flex-shrink-0 hover:opacity-80 transition-opacity shadow-sm">
              <span className="text-sm font-semibold text-white">
                {post.author_username?.[0]?.toUpperCase() || "?"}
              </span>
            </div>
          </Link>
          <div>
            <Link
              href={`/users/${post.author_id}`}
              className="font-semibold text-sm hover:text-primary transition-colors"
            >
              {post.author_username || "未知用户"}
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
              {post.content_labels?.is_ai_generated && (
                <>
                  <span>·</span>
                  <span className="text-brand-purple">AI 生成</span>
                </>
              )}
              {post.group_id && post.group_name && (
                <>
                  <span>·</span>
                  <Link
                    href={`/groups/${post.group_id}`}
                    className="text-brand-teal hover:underline"
                  >
                    {post.group_name}
                  </Link>
                </>
              )}
            </div>
          </div>
        </div>
        {/* More menu */}
        <div className="relative" ref={menuRef}>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => setMenuOpen((v) => !v)}
          >
            <MoreHorizontal className="h-4 w-4" />
          </Button>
          {menuOpen && (
            <div className="absolute right-0 top-8 bg-background border rounded-lg shadow-lg py-1 z-10 w-32">
              <button
                onClick={() => {
                  setMenuOpen(false);
                  setShowReport(true);
                }}
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
          <Link
            href={`/posts/${post.id}`}
            className="hover:text-primary transition-colors"
          >
            {post.title}
          </Link>
        </h2>
      )}

      {/* Content */}
      <div className="mb-3">
        <Link href={`/posts/${post.id}`}>
          <p
            className={`text-sm leading-relaxed whitespace-pre-wrap ${!showFull && "line-clamp-5"}`}
          >
            {post.content}
          </p>
        </Link>
      </div>

      {/* Media */}
      {mediaUrls.length > 0 && (
        <div
          className={cn(
            "gap-2 mb-3",
            mediaUrls.length === 1 && "grid grid-cols-1",
            mediaUrls.length === 2 && "grid grid-cols-2",
            mediaUrls.length === 3 && "grid grid-cols-[2fr_1fr]",
            mediaUrls.length === 4 && "grid grid-cols-2",
          )}
        >
          {mediaUrls.length === 3 ? (
            <>
              <img
                src={mediaUrls[0]}
                alt=""
                className="w-full aspect-square object-cover rounded-lg row-span-2 hover:scale-[1.02] transition-transform duration-200"
              />
              <img
                src={mediaUrls[1]}
                alt=""
                className="w-full aspect-square object-cover rounded-lg hover:scale-[1.02] transition-transform duration-200"
              />
              <img
                src={mediaUrls[2]}
                alt=""
                className="w-full aspect-square object-cover rounded-lg hover:scale-[1.02] transition-transform duration-200"
              />
            </>
          ) : (
            mediaUrls.map((url, i) => (
              <img
                key={i}
                src={url}
                alt=""
                className="w-full aspect-square object-cover rounded-lg hover:scale-[1.02] transition-transform duration-200"
              />
            ))
          )}
        </div>
      )}

      {/* Tags */}
      {post.tags && post.tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3">
          {post.tags.map((tag) => (
            <Badge
              key={tag}
              variant="outline"
              className="text-xs border-brand-purple/40 text-brand-purple hover:bg-brand-purple/10 transition-colors cursor-pointer"
            >
              #{tag}
            </Badge>
          ))}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-4 pt-2 border-t">
        <motion.button
          onClick={() => !isPending && !isBlocked && likeMutation.mutate()}
          disabled={isPending || isBlocked}
          whileTap={{ scale: 0.82 }}
          transition={{ type: "spring", stiffness: 400, damping: 15 }}
          className={cn(
            "flex items-center gap-1.5 text-sm transition-colors disabled:opacity-40 disabled:cursor-not-allowed",
            liked
              ? "text-red-500 hover:text-red-600"
              : "text-muted-foreground hover:text-red-500",
          )}
        >
          <Heart className={cn("h-4 w-4", liked && "fill-current")} />
          <motion.span
            key={likeCount}
            initial={{ y: liked ? -6 : 6, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.15 }}
          >
            {likeCount}
          </motion.span>
        </motion.button>

        <Link
          href={isPending || isBlocked ? "#" : `/posts/${post.id}`}
          className={cn(
            "flex items-center gap-1.5 text-sm text-muted-foreground transition-colors",
            isPending || isBlocked
              ? "opacity-40 pointer-events-none"
              : "hover:text-primary",
          )}
        >
          <MessageCircle className="h-4 w-4" />
          <span>{post.comment_count}</span>
        </Link>

        <button
          className={cn(
            "flex items-center gap-1.5 text-sm transition-colors",
            bookmarked
              ? "text-brand-purple hover:text-brand-purple/80"
              : "text-muted-foreground hover:text-brand-purple",
          )}
          onClick={() => bookmarkMutation.mutate()}
        >
          <Bookmark className={cn("h-4 w-4", bookmarked && "fill-current")} />
          <span>收藏</span>
        </button>

        <button
          className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors ml-auto"
          onClick={() => {
            if (navigator.share) {
              navigator.share({
                url: `/posts/${post.id}`,
                title: post.title || post.content.slice(0, 40),
              });
            }
          }}
        >
          <Share2 className="h-4 w-4" />
        </button>
      </div>

      {showReport && (
        <ReportModal postId={post.id} onClose={() => setShowReport(false)} />
      )}
    </div>
  );
}
