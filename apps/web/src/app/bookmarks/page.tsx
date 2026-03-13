"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { ElementType } from "react";
import { Bookmark, Calendar, Users, FileText } from "lucide-react";
import { apiClient, Event, Group, Post } from "@/lib/api-client";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PostCard } from "@/components/post/post-card";
import { Button } from "@/components/ui/button";

export default function BookmarksPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      apiClient.setToken(token);
    }
    Promise.all([
      apiClient.getBookmarkedPosts().catch(() => ({ posts: [] })),
      apiClient.getBookmarkedGroups().catch(() => ({ groups: [] })),
      apiClient.getBookmarkedEvents().catch(() => ({ events: [] })),
    ])
      .then(([postsRes, groupsRes, eventsRes]) => {
        setPosts(postsRes.posts ?? []);
        setGroups(groupsRes.groups ?? []);
        setEvents(eventsRes.events ?? []);
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="max-w-3xl mx-auto pt-20 px-4 pb-10">
      <div className="mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Bookmark className="h-6 w-6" />
          我的收藏
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          收藏你想稍后再看的帖子、圈子和活动。
        </p>
      </div>

      <Tabs defaultValue="posts">
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
                <PostCard key={post.id} post={post} />
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
                <Link
                  key={group.id}
                  href={`/groups/${group.id}`}
                  className="rounded-2xl border bg-card p-5 hover:shadow-md transition-shadow"
                >
                  <p className="font-semibold">{group.name}</p>
                  <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                    {group.description}
                  </p>
                  <p className="text-xs text-muted-foreground mt-3">
                    {group.member_count} 成员 · {group.post_count} 帖子
                  </p>
                </Link>
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
                <Link
                  key={event.id}
                  href={`/events/${event.id}`}
                  className="rounded-2xl border bg-card p-5 hover:shadow-md transition-shadow"
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
