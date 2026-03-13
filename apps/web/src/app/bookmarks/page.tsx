"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { ElementType } from "react";
import { Bookmark, Calendar, Users, FileText, Trash2 } from "lucide-react";
import { apiClient, Event, Group, Post } from "@/lib/api-client";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PostCard } from "@/components/post/post-card";
import { Button } from "@/components/ui/button";

type TabKey = "posts" | "groups" | "events";
type SortKey = "latest" | "oldest";

export default function BookmarksPage() {
  const [tab, setTab] = useState<TabKey>("posts");
  const [sort, setSort] = useState<SortKey>("latest");
  const [posts, setPosts] = useState<Post[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      apiClient.setToken(token);
    }
    setLoading(true);
    Promise.all([
      apiClient
        .getBookmarkedPostsWithSort(1, 20, sort)
        .catch(() => ({ posts: [] })),
      apiClient
        .getBookmarkedGroupsWithSort(1, 20, sort)
        .catch(() => ({ groups: [] })),
      apiClient
        .getBookmarkedEventsWithSort(1, 20, sort)
        .catch(() => ({ events: [] })),
    ])
      .then(([postsRes, groupsRes, eventsRes]) => {
        setPosts(postsRes.posts ?? []);
        setGroups(groupsRes.groups ?? []);
        setEvents(eventsRes.events ?? []);
      })
      .finally(() => setLoading(false));
  }, [sort]);

  useEffect(() => {
    setSelectedIds([]);
  }, [tab, sort]);

  function toggleSelected(id: string) {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id],
    );
  }

  async function handleBatchRemove() {
    if (selectedIds.length === 0) return;
    const targetType =
      tab === "posts" ? "post" : tab === "groups" ? "group" : "event";

    await apiClient.batchDeleteBookmarks(targetType, selectedIds);

    if (tab === "posts") {
      setPosts((prev) => prev.filter((item) => !selectedIds.includes(item.id)));
    } else if (tab === "groups") {
      setGroups((prev) =>
        prev.filter((item) => !selectedIds.includes(item.id)),
      );
    } else {
      setEvents((prev) =>
        prev.filter((item) => !selectedIds.includes(item.id)),
      );
    }

    setSelectedIds([]);
  }

  return (
    <div className="max-w-3xl mx-auto pt-20 px-4 pb-10">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Bookmark className="h-6 w-6" />
            我的收藏
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            收藏你想稍后再看的帖子、圈子和活动。
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={sort}
            onChange={(e) => setSort(e.target.value as SortKey)}
            className="rounded-lg border bg-background px-3 py-2 text-sm"
          >
            <option value="latest">最近收藏</option>
            <option value="oldest">最早收藏</option>
          </select>
          {selectedIds.length > 0 && (
            <Button variant="outline" onClick={handleBatchRemove}>
              <Trash2 className="h-4 w-4 mr-2" />
              批量移除 ({selectedIds.length})
            </Button>
          )}
        </div>
      </div>

      <Tabs
        defaultValue="posts"
        onValueChange={(value) => setTab(value as TabKey)}
      >
        <TabsList className="w-full mb-6">
          <TabsTrigger value="posts" className="flex-1">
            帖子
          </TabsTrigger>
          <TabsTrigger value="groups" className="flex-1">
            圈子
          </TabsTrigger>
          <TabsTrigger value="events" className="flex-1">
            活动
          </TabsTrigger>
        </TabsList>

        <TabsContent value="posts">
          {loading ? (
            <div className="space-y-3">
              {[0, 1].map((i) => (
                <div
                  key={i}
                  className="h-32 rounded-xl bg-muted animate-pulse"
                />
              ))}
            </div>
          ) : posts.length === 0 ? (
            <EmptyState
              icon={FileText}
              title="还没有收藏任何帖子"
              description="看到喜欢的内容时，点一下收藏按钮就会出现在这里。"
              href="/explore"
              action="去发现页逛逛"
            />
          ) : (
            <div className="space-y-4">
              {posts.map((post) => (
                <div key={post.id} className="space-y-2">
                  <label className="flex items-center gap-2 text-sm text-muted-foreground">
                    <input
                      type="checkbox"
                      checked={selectedIds.includes(post.id)}
                      onChange={() => toggleSelected(post.id)}
                    />
                    选择这条收藏
                  </label>
                  <PostCard post={post} />
                </div>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="groups">
          {loading ? (
            <div className="grid gap-3 sm:grid-cols-2">
              {[0, 1].map((i) => (
                <div
                  key={i}
                  className="h-36 rounded-xl bg-muted animate-pulse"
                />
              ))}
            </div>
          ) : groups.length === 0 ? (
            <EmptyState
              icon={Users}
              title="还没有收藏任何圈子"
              description="把感兴趣的圈子先收藏起来，之后回来看更方便。"
              href="/groups"
              action="去圈子广场"
            />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {groups.map((group) => (
                <div key={group.id} className="rounded-2xl border bg-card p-5">
                  <label className="flex items-center gap-2 text-sm text-muted-foreground mb-3">
                    <input
                      type="checkbox"
                      checked={selectedIds.includes(group.id)}
                      onChange={() => toggleSelected(group.id)}
                    />
                    选择这个圈子
                  </label>
                  <Link
                    href={`/groups/${group.id}`}
                    className="block hover:opacity-90 transition-opacity"
                  >
                    <p className="font-semibold">{group.name}</p>
                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                      {group.description}
                    </p>
                    <p className="text-xs text-muted-foreground mt-3">
                      {group.member_count} 成员 · {group.post_count} 帖子
                    </p>
                  </Link>
                </div>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="events">
          {loading ? (
            <div className="grid gap-3 sm:grid-cols-2">
              {[0, 1].map((i) => (
                <div
                  key={i}
                  className="h-36 rounded-xl bg-muted animate-pulse"
                />
              ))}
            </div>
          ) : events.length === 0 ? (
            <EmptyState
              icon={Calendar}
              title="还没有收藏任何活动"
              description="先收藏感兴趣的活动，后面再统一安排时间。"
              href="/events"
              action="去活动广场"
            />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {events.map((event) => (
                <div key={event.id} className="rounded-2xl border bg-card p-5">
                  <label className="flex items-center gap-2 text-sm text-muted-foreground mb-3">
                    <input
                      type="checkbox"
                      checked={selectedIds.includes(event.id)}
                      onChange={() => toggleSelected(event.id)}
                    />
                    选择这个活动
                  </label>
                  <Link
                    href={`/events/${event.id}`}
                    className="block hover:opacity-90 transition-opacity"
                  >
                    <p className="font-semibold">{event.title}</p>
                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                      {event.description}
                    </p>
                    <p className="text-xs text-muted-foreground mt-3">
                      {new Date(event.start_time).toLocaleDateString("zh-CN")} ·{" "}
                      {event.is_online ? "线上" : event.location || "线下"}
                    </p>
                  </Link>
                </div>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

function EmptyState({
  icon: Icon,
  title,
  description,
  href,
  action,
}: {
  icon: ElementType;
  title: string;
  description: string;
  href: string;
  action: string;
}) {
  return (
    <div className="text-center py-16 text-muted-foreground">
      <Icon className="h-12 w-12 mx-auto mb-4 opacity-30" />
      <p className="font-medium mb-2">{title}</p>
      <p className="text-sm mb-6">{description}</p>
      <Link href={href}>
        <Button variant="outline">{action}</Button>
      </Link>
    </div>
  );
}
