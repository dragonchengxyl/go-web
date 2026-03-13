"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Users,
  FileText,
  Lock,
  Globe,
  Crown,
  Shield,
  UserPlus,
  UserMinus,
  Sparkles,
  PenSquare,
  Bookmark,
  Pin,
} from "lucide-react";
import { apiClient, Group, GroupMember, Post } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { PostCard } from "@/components/post/post-card";
import { Textarea } from "@/components/ui/textarea";

const ROLE_CONFIG: Record<
  string,
  { label: string; icon: typeof Crown; color: string }
> = {
  owner: { label: "圈主", icon: Crown, color: "text-yellow-500" },
  moderator: { label: "管理员", icon: Shield, color: "text-blue-500" },
  member: { label: "成员", icon: Users, color: "text-muted-foreground" },
};

const GRADIENTS = [
  "from-purple-500 to-teal-400",
  "from-teal-400 to-blue-500",
  "from-orange-400 to-pink-500",
  "from-blue-500 to-indigo-600",
  "from-green-400 to-teal-500",
  "from-pink-500 to-purple-600",
];
function hashGradient(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++)
    hash = (hash * 31 + str.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(hash) % GRADIENTS.length];
}

export default function GroupDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [sortMode, setSortMode] = useState<"latest" | "hot">("latest");
  const [activeTag, setActiveTag] = useState("");
  const [group, setGroup] = useState<Group | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [postsLoading, setPostsLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [isMember, setIsMember] = useState(false);
  const [myId, setMyId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [posts, setPosts] = useState<Post[]>([]);
  const [highlights, setHighlights] = useState<Post[]>([]);
  const [featuredPost, setFeaturedPost] = useState<Post | null>(null);
  const [bookmarked, setBookmarked] = useState(false);
  const [bookmarkLoading, setBookmarkLoading] = useState(false);
  const [pinningPostId, setPinningPostId] = useState<string | null>(null);
  const [featuringPostId, setFeaturingPostId] = useState<string | null>(null);
  const [savingGroupMeta, setSavingGroupMeta] = useState(false);
  const [announcementDraft, setAnnouncementDraft] = useState("");
  const [rulesDraft, setRulesDraft] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      apiClient.setToken(token);
      setIsLoggedIn(true);
      apiClient
        .getMe()
        .then((me) => {
          setMyId(me?.id ?? null);
        })
        .catch(() => {});
      if (id) {
        apiClient
          .checkBookmark("group", id)
          .then((res) => setBookmarked(res.bookmarked))
          .catch(() => {});
      }
    }
    if (!id) return;
    Promise.all([
      apiClient.getGroup(id),
      apiClient.listGroupMembers(id),
      token ? apiClient.getMe().catch(() => null) : Promise.resolve(null),
    ])
      .then(([g, mRes, me]) => {
        const list = mRes.members ?? [];
        setGroup(g);
        setAnnouncementDraft(g.announcement || "");
        setRulesDraft(g.rules || "");
        setMembers(list);
        setIsMember(!!me && list.some((m) => m.user_id === me?.id));
      })
      .catch(console.error)
      .finally(() => {
        setLoading(false);
      });
  }, [id]);

  useEffect(() => {
    if (!id) return;
    setPostsLoading(true);
    Promise.all([
      apiClient
        .getGroupPosts(id, 1, 20, {
          sort: sortMode,
          tag: activeTag || undefined,
        })
        .catch(() => ({ posts: [] })),
      apiClient.getGroupHighlights(id).catch(() => ({ posts: [] })),
      apiClient.getGroupPostTags(id).catch(() => []),
    ])
      .then(([postsRes, highlightsRes, tagsRes]) => {
        setPosts(postsRes.posts ?? []);
        setHighlights(highlightsRes.posts ?? []);
        setAvailableTags(tagsRes ?? []);
      })
      .finally(() => setPostsLoading(false));
  }, [id, sortMode, activeTag]);

  useEffect(() => {
    if (!group?.featured_post_id) {
      setFeaturedPost(null);
      return;
    }
    apiClient
      .getPost(group.featured_post_id)
      .then(setFeaturedPost)
      .catch(() => setFeaturedPost(null));
  }, [group?.featured_post_id]);

  const handleJoin = async () => {
    if (!id) return;
    setJoining(true);
    setError("");
    try {
      await apiClient.joinGroup(id);
      const [updated, mRes] = await Promise.all([
        apiClient.getGroup(id),
        apiClient.listGroupMembers(id),
      ]);
      setGroup(updated);
      setMembers(mRes.members ?? []);
      setIsMember(true);
      setActiveTag("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "加入失败");
    } finally {
      setJoining(false);
    }
  };

  const handleLeave = async () => {
    if (!id) return;
    setJoining(true);
    setError("");
    try {
      await apiClient.leaveGroup(id);
      const [updated, mRes] = await Promise.all([
        apiClient.getGroup(id),
        apiClient.listGroupMembers(id),
      ]);
      setGroup(updated);
      setMembers(mRes.members ?? []);
      setIsMember(false);
      if (updated.privacy === "private") {
        setPosts([]);
        setHighlights([]);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "退出失败");
    } finally {
      setJoining(false);
    }
  };

  const handleBookmark = async () => {
    if (!id) return;
    setBookmarkLoading(true);
    try {
      if (bookmarked) {
        await apiClient.unbookmarkGroup(id);
        setBookmarked(false);
      } else {
        await apiClient.bookmarkGroup(id);
        setBookmarked(true);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBookmarkLoading(false);
    }
  };

  const handlePin = async (postId: string, pin: boolean) => {
    if (!id) return;
    setPinningPostId(postId);
    setError("");
    try {
      if (pin) {
        await apiClient.pinGroupPost(id, postId);
      } else {
        await apiClient.unpinGroupPost(id, postId);
      }
      const [postsRes, highlightsRes] = await Promise.all([
        apiClient
          .getGroupPosts(id, 1, 20, {
            sort: sortMode,
            tag: activeTag || undefined,
          })
          .catch(() => ({ posts: [] })),
        apiClient.getGroupHighlights(id).catch(() => ({ posts: [] })),
      ]);
      setPosts(postsRes.posts ?? []);
      setHighlights(highlightsRes.posts ?? []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "置顶操作失败");
    } finally {
      setPinningPostId(null);
    }
  };

  const handleSetFeatured = async (postId?: string) => {
    if (!id) return;
    setFeaturingPostId(postId ?? "clear");
    setError("");
    try {
      const updated = await apiClient.setGroupFeaturedPost(id, postId);
      setGroup(updated);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "设置精选失败");
    } finally {
      setFeaturingPostId(null);
    }
  };

  const handleSaveGroupMeta = async () => {
    if (!id || !group) return;
    setSavingGroupMeta(true);
    setError("");
    try {
      const updated = await apiClient.updateGroup(id, {
        name: group.name,
        description: group.description,
        announcement: announcementDraft,
        rules: rulesDraft,
        tags: group.tags,
        privacy: group.privacy,
      });
      setGroup(updated);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "保存圈子信息失败");
    } finally {
      setSavingGroupMeta(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4">
        <div className="h-8 w-32 bg-muted animate-pulse rounded mb-6" />
        <div className="h-48 bg-muted animate-pulse rounded-2xl mb-4" />
        <div className="h-32 bg-muted animate-pulse rounded-2xl" />
      </div>
    );
  }

  if (!group) {
    return (
      <div className="max-w-2xl mx-auto pt-20 px-4 text-center py-16 text-muted-foreground">
        <p className="text-xl font-medium mb-2">圈子不存在</p>
        <p className="text-sm mb-6">该圈子可能已被删除</p>
        <Link href="/groups">
          <Button variant="outline">返回圈子列表</Button>
        </Link>
      </div>
    );
  }

  const isOwner =
    myId && members.find((m) => m.user_id === myId)?.role === "owner";
  const myRole = members.find((m) => m.user_id === myId)?.role;
  const canManagePosts = myRole === "owner" || myRole === "moderator";
  const gradient = hashGradient(group.id);

  return (
    <div className="max-w-2xl mx-auto pt-20 px-4 pb-12">
      {/* Back */}
      <Link href="/groups">
        <Button variant="ghost" size="sm" className="mb-6 -ml-2">
          <ArrowLeft className="h-4 w-4 mr-1" />
          返回圈子
        </Button>
      </Link>

      {/* Header card */}
      <div className="bg-card border rounded-2xl overflow-hidden mb-5">
        {/* Banner */}
        <div className={`h-24 bg-gradient-to-br ${gradient} opacity-60`} />

        <div className="px-6 pb-6">
          {/* Avatar row */}
          <div className="flex items-end justify-between -mt-8 mb-4">
            <div className="w-16 h-16 rounded-2xl bg-card border-4 border-background shadow-md flex items-center justify-center text-2xl bg-gradient-to-br from-brand-purple/20 to-brand-teal/20">
              🐾
            </div>
            <div className="flex items-center gap-1.5">
              {isLoggedIn && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleBookmark}
                  disabled={bookmarkLoading}
                  className={
                    bookmarked ? "border-brand-purple text-brand-purple" : ""
                  }
                >
                  <Bookmark
                    className={`h-4 w-4 mr-1 ${bookmarked ? "fill-current" : ""}`}
                  />
                  {bookmarked ? "已收藏" : "收藏"}
                </Button>
              )}
              {group.privacy === "private" && (
                <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground border border-border">
                  <Lock className="h-3 w-3" />
                  私密
                </span>
              )}
              {group.privacy === "public" && (
                <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20">
                  <Globe className="h-3 w-3" />
                  公开
                </span>
              )}
            </div>
          </div>

          <h1 className="text-xl font-bold mb-1">{group.name}</h1>

          <div className="flex gap-4 text-sm text-muted-foreground mb-3">
            <span className="flex items-center gap-1.5">
              <Users className="h-3.5 w-3.5" />
              {group.member_count} 成员
            </span>
            <span className="flex items-center gap-1.5">
              <FileText className="h-3.5 w-3.5" />
              {group.post_count} 帖子
            </span>
          </div>

          {group.tags?.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-3">
              {group.tags.map((t) => (
                <Link key={t} href={`/tags/${encodeURIComponent(t)}`}>
                  <span className="text-xs px-2 py-0.5 rounded-full bg-primary/10 text-primary hover:bg-primary/20 transition-colors cursor-pointer">
                    #{t}
                  </span>
                </Link>
              ))}
            </div>
          )}

          {group.description && (
            <p className="text-sm text-muted-foreground leading-relaxed">
              {group.description}
            </p>
          )}
        </div>
      </div>

      {/* Join/Leave */}
      {!isOwner && isLoggedIn && (
        <div className="mb-5">
          {error && <p className="text-destructive text-sm mb-2">{error}</p>}
          {isMember ? (
            <Button
              onClick={handleLeave}
              disabled={joining}
              variant="outline"
              className="w-full"
              size="lg"
            >
              <UserMinus className="h-4 w-4 mr-2" />
              {joining ? "处理中…" : "退出圈子"}
            </Button>
          ) : (
            <Button
              onClick={handleJoin}
              disabled={joining}
              size="lg"
              className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
            >
              <UserPlus className="h-4 w-4 mr-2" />
              {joining ? "处理中…" : "加入圈子"}
            </Button>
          )}
        </div>
      )}

      {!isLoggedIn && (
        <Link href="/login" className="block mb-5">
          <Button
            size="lg"
            className="w-full bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110"
          >
            <UserPlus className="h-4 w-4 mr-2" />
            登录后加入
          </Button>
        </Link>
      )}

      {/* Create post CTA */}
      {isMember && (
        <div className="mb-5 rounded-2xl border bg-card p-5 flex items-center justify-between gap-4">
          <div>
            <p className="font-semibold">在圈子里发点新内容</p>
            <p className="text-sm text-muted-foreground mt-1">
              发布后会进入这个圈子的最新内容流。
            </p>
          </div>
          <Link href={`/posts/create?group_id=${id}`}>
            <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
              <PenSquare className="h-4 w-4 mr-2" />
              圈子发帖
            </Button>
          </Link>
        </div>
      )}

      {/* Announcement / Rules */}
      {(group.announcement || canManagePosts) && (
        <div className="bg-card border rounded-2xl p-6 mb-5">
          <h2 className="font-semibold mb-3">圈子公告</h2>
          {canManagePosts ? (
            <div className="space-y-3">
              <Textarea
                value={announcementDraft}
                onChange={(e) => setAnnouncementDraft(e.target.value)}
                rows={4}
                placeholder="给圈友们写一条公告..."
              />
              <h3 className="font-medium">圈子规则</h3>
              <Textarea
                value={rulesDraft}
                onChange={(e) => setRulesDraft(e.target.value)}
                rows={5}
                placeholder="写下发帖规范、活动要求或交流边界..."
              />
              <div className="flex justify-end">
                <Button
                  onClick={handleSaveGroupMeta}
                  disabled={savingGroupMeta}
                >
                  {savingGroupMeta ? "保存中..." : "保存公告与规则"}
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {group.announcement && (
                <p className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
                  {group.announcement}
                </p>
              )}
              {group.rules && (
                <div>
                  <h3 className="font-medium mb-2">圈子规则</h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
                    {group.rules}
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Featured post */}
      {featuredPost && (
        <div className="bg-card border rounded-2xl p-6 mb-5">
          <div className="flex items-center justify-between gap-3 mb-4">
            <div className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-yellow-500" />
              <h2 className="font-semibold">精选内容</h2>
            </div>
            {canManagePosts && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleSetFeatured()}
                disabled={featuringPostId === "clear"}
              >
                {featuringPostId === "clear" ? "处理中..." : "取消精选"}
              </Button>
            )}
          </div>
          <PostCard post={featuredPost} />
        </div>
      )}

      {/* Highlights */}
      {highlights.length > 0 && sortMode === "latest" && !activeTag && (
        <div className="bg-card border rounded-2xl p-6 mb-5">
          <div className="flex items-center gap-2 mb-4">
            <Sparkles className="h-4 w-4 text-yellow-500" />
            <h2 className="font-semibold">圈子精选</h2>
          </div>
          <div className="grid gap-3">
            {highlights.map((post) => (
              <Link
                key={post.id}
                href={`/posts/${post.id}`}
                className="rounded-xl border p-4 hover:bg-muted/40 transition-colors"
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-medium truncate">
                      {post.title || post.content.slice(0, 32)}
                    </p>
                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                      {post.content}
                    </p>
                  </div>
                  <div className="text-xs text-muted-foreground shrink-0">
                    {post.like_count} 赞 · {post.comment_count} 评
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Latest posts */}
      <div className="bg-card border rounded-2xl p-6 mb-5">
        <div className="flex items-center justify-between gap-3 mb-4">
          <h2 className="font-semibold">
            {sortMode === "hot" ? "热门内容" : "最新内容"}
          </h2>
          <span className="text-xs text-muted-foreground">
            {posts.length} 条
          </span>
        </div>
        <div className="flex flex-wrap gap-2 mb-4">
          {(["latest", "hot"] as const).map((mode) => (
            <button
              key={mode}
              type="button"
              onClick={() => setSortMode(mode)}
              className={`px-3 py-1.5 rounded-full text-xs border transition-colors ${
                sortMode === mode
                  ? "bg-primary text-primary-foreground border-primary"
                  : "text-muted-foreground hover:border-primary/40 hover:text-foreground"
              }`}
            >
              {mode === "latest" ? "最新" : "热门"}
            </button>
          ))}
        </div>
        {availableTags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-4">
            <button
              type="button"
              onClick={() => setActiveTag("")}
              className={`px-3 py-1.5 rounded-full text-xs border transition-colors ${
                activeTag === ""
                  ? "bg-brand-teal text-white border-brand-teal"
                  : "text-muted-foreground hover:border-brand-teal/40 hover:text-foreground"
              }`}
            >
              全部
            </button>
            {availableTags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => setActiveTag(tag)}
                className={`px-3 py-1.5 rounded-full text-xs border transition-colors ${
                  activeTag === tag
                    ? "bg-brand-teal text-white border-brand-teal"
                    : "text-muted-foreground hover:border-brand-teal/40 hover:text-foreground"
                }`}
              >
                #{tag}
              </button>
            ))}
          </div>
        )}
        {postsLoading ? (
          <div className="space-y-3">
            {[0, 1].map((i) => (
              <div key={i} className="h-32 rounded-xl bg-muted animate-pulse" />
            ))}
          </div>
        ) : posts.length === 0 ? (
          <div className="text-center py-10 text-muted-foreground">
            <FileText className="h-10 w-10 mx-auto mb-3 opacity-30" />
            <p className="font-medium mb-1">这个圈子还没有内容</p>
            <p className="text-sm">加入后发出第一条圈子动态吧。</p>
          </div>
        ) : (
          <div className="space-y-4">
            {posts.map((post) => (
              <div key={post.id} className="space-y-2">
                {canManagePosts && (
                  <div className="flex justify-end gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handlePin(post.id, !post.is_pinned)}
                      disabled={pinningPostId === post.id}
                      className={
                        post.is_pinned
                          ? "border-yellow-500 text-yellow-600"
                          : ""
                      }
                    >
                      <Pin
                        className={`h-4 w-4 mr-1 ${post.is_pinned ? "fill-current" : ""}`}
                      />
                      {pinningPostId === post.id
                        ? "处理中..."
                        : post.is_pinned
                          ? "取消置顶"
                          : "置顶帖子"}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleSetFeatured(post.id)}
                      disabled={featuringPostId === post.id}
                      className={
                        group.featured_post_id === post.id
                          ? "border-yellow-500 text-yellow-600"
                          : ""
                      }
                    >
                      <Sparkles className="h-4 w-4 mr-1" />
                      {featuringPostId === post.id
                        ? "处理中..."
                        : group.featured_post_id === post.id
                          ? "当前精选"
                          : "设为精选"}
                    </Button>
                  </div>
                )}
                <PostCard post={post} />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Members */}
      {members.length > 0 && (
        <div className="bg-card border rounded-2xl p-6">
          <h2 className="font-semibold mb-4">成员（{members.length}）</h2>
          <div className="space-y-2">
            {members.map((m) => {
              const roleConf = ROLE_CONFIG[m.role] ?? ROLE_CONFIG.member;
              const RoleIcon = roleConf.icon;
              const mGradient = hashGradient(m.user_id);
              return (
                <Link
                  key={m.user_id}
                  href={`/users/${m.user_id}`}
                  className="flex items-center justify-between p-2.5 rounded-xl hover:bg-muted/60 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-9 h-9 rounded-full bg-gradient-to-br ${mGradient} flex items-center justify-center text-xs text-white font-bold flex-shrink-0`}
                    >
                      {m.user_id.slice(0, 2).toUpperCase()}
                    </div>
                    <span className="text-sm font-medium">
                      {m.user_id.slice(0, 8)}…
                    </span>
                  </div>
                  <span
                    className={`flex items-center gap-1 text-xs ${roleConf.color}`}
                  >
                    <RoleIcon className="h-3 w-3" />
                    {roleConf.label}
                  </span>
                </Link>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
