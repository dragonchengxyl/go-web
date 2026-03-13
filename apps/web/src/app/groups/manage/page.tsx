"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  BarChart3,
  MessageSquareText,
  PenSquare,
  Sparkles,
  Users,
} from "lucide-react";
import {
  apiClient,
  Group,
  GroupDashboard,
  GroupDashboardItem,
  Post,
} from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

export default function GroupManagePage() {
  const [dashboard, setDashboard] = useState<GroupDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [announcementGroup, setAnnouncementGroup] = useState<Group | null>(
    null,
  );
  const [announcementDraft, setAnnouncementDraft] = useState("");
  const [savingAnnouncement, setSavingAnnouncement] = useState(false);
  const [featuredGroup, setFeaturedGroup] = useState<Group | null>(null);
  const [candidatePosts, setCandidatePosts] = useState<Post[]>([]);
  const [loadingCandidates, setLoadingCandidates] = useState(false);
  const [savingFeaturedPostId, setSavingFeaturedPostId] = useState<
    string | null
  >(null);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (token) {
      apiClient.setToken(token);
    }
    apiClient
      .getMyGroupsDashboard()
      .then(setDashboard)
      .finally(() => setLoading(false));
  }, []);

  async function reloadDashboard() {
    const next = await apiClient.getMyGroupsDashboard();
    setDashboard(next);
  }

  async function handleSaveAnnouncement() {
    if (!announcementGroup) return;
    setSavingAnnouncement(true);
    try {
      await apiClient.updateGroup(announcementGroup.id, {
        name: announcementGroup.name,
        description: announcementGroup.description,
        announcement: announcementDraft,
        rules: announcementGroup.rules,
        tags: announcementGroup.tags,
        privacy: announcementGroup.privacy,
      });
      setAnnouncementGroup(null);
      await reloadDashboard();
    } finally {
      setSavingAnnouncement(false);
    }
  }

  async function openFeaturedDialog(group: Group) {
    setFeaturedGroup(group);
    setLoadingCandidates(true);
    try {
      const data = await apiClient.getGroupPosts(group.id, 1, 8, {
        sort: "latest",
      });
      setCandidatePosts(data.posts ?? []);
    } finally {
      setLoadingCandidates(false);
    }
  }

  async function setFeatured(postId?: string) {
    if (!featuredGroup) return;
    setSavingFeaturedPostId(postId ?? "clear");
    try {
      await apiClient.setGroupFeaturedPost(featuredGroup.id, postId);
      await reloadDashboard();
      setFeaturedGroup(null);
    } finally {
      setSavingFeaturedPostId(null);
    }
  }

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto pt-20 px-4 pb-10">
        <div className="h-10 w-52 rounded-xl bg-muted animate-pulse mb-6" />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className="h-28 rounded-2xl bg-muted animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  const createdGroups = dashboard?.created_groups ?? [];
  const managedGroups = dashboard?.managed_groups ?? [];
  const stats = dashboard?.stats;

  return (
    <div className="max-w-5xl mx-auto pt-20 px-4 pb-10">
      <div className="flex items-start justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold">我的圈子主页</h1>
          <p className="text-sm text-muted-foreground mt-2">
            查看你创建或管理的圈子，快速发布公告、设置精选和发圈子动态。
          </p>
        </div>
        <Link href="/groups/create">
          <Button className="bg-gradient-to-r from-brand-purple to-brand-teal text-white border-0 hover:brightness-110">
            <PenSquare className="h-4 w-4 mr-2" />
            创建新圈子
          </Button>
        </Link>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <StatCard label="我创建的圈子" value={stats?.created_count ?? 0} />
        <StatCard label="我管理的圈子" value={stats?.managed_count ?? 0} />
        <StatCard label="覆盖成员总数" value={stats?.total_members ?? 0} />
        <StatCard
          label="已设精选圈子"
          value={stats?.featured_group_count ?? 0}
        />
      </div>

      <section className="mb-10">
        <SectionTitle
          title="我创建的圈子"
          subtitle="你是圈主，可以进行完整运营操作。"
        />
        {createdGroups.length === 0 ? (
          <EmptyBlock text="你还没有创建任何圈子。" />
        ) : (
          <div className="grid gap-4">
            {createdGroups.map((item) => (
              <GroupManageCard
                key={`owner-${item.group.id}`}
                item={item}
                onQuickAnnouncement={() => {
                  setAnnouncementGroup(item.group);
                  setAnnouncementDraft(item.group.announcement || "");
                }}
                onQuickFeatured={() => openFeaturedDialog(item.group)}
              />
            ))}
          </div>
        )}
      </section>

      <section>
        <SectionTitle
          title="我管理的圈子"
          subtitle="你是管理员，可以协助维护内容和公告。"
        />
        {managedGroups.length === 0 ? (
          <EmptyBlock text="你目前还没有担任任何圈子的管理员。" />
        ) : (
          <div className="grid gap-4">
            {managedGroups.map((item) => (
              <GroupManageCard
                key={`mod-${item.group.id}`}
                item={item}
                onQuickAnnouncement={() => {
                  setAnnouncementGroup(item.group);
                  setAnnouncementDraft(item.group.announcement || "");
                }}
                onQuickFeatured={() => openFeaturedDialog(item.group)}
              />
            ))}
          </div>
        )}
      </section>

      <Dialog
        open={!!announcementGroup}
        onOpenChange={(open) => !open && setAnnouncementGroup(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>快速发布公告</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              {announcementGroup?.name}
            </p>
            <Textarea
              rows={6}
              value={announcementDraft}
              onChange={(e) => setAnnouncementDraft(e.target.value)}
              placeholder="写一条新的圈子公告..."
            />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setAnnouncementGroup(null)}>
              取消
            </Button>
            <Button
              onClick={handleSaveAnnouncement}
              disabled={savingAnnouncement}
            >
              {savingAnnouncement ? "保存中..." : "保存公告"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!featuredGroup}
        onOpenChange={(open) => !open && setFeaturedGroup(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>设置精选内容</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              {featuredGroup?.name}
            </p>
            {loadingCandidates ? (
              <div className="h-32 rounded-xl bg-muted animate-pulse" />
            ) : candidatePosts.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                这个圈子还没有可选帖子。
              </p>
            ) : (
              <div className="space-y-2">
                {candidatePosts.map((post) => (
                  <button
                    key={post.id}
                    type="button"
                    onClick={() => setFeatured(post.id)}
                    disabled={savingFeaturedPostId === post.id}
                    className="w-full rounded-xl border p-3 text-left hover:bg-muted/50 transition-colors"
                  >
                    <p className="font-medium">
                      {post.title || post.content.slice(0, 32)}
                    </p>
                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                      {post.content}
                    </p>
                  </button>
                ))}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setFeaturedGroup(null)}>
              取消
            </Button>
            <Button
              variant="outline"
              onClick={() => setFeatured()}
              disabled={savingFeaturedPostId === "clear"}
            >
              {savingFeaturedPostId === "clear" ? "处理中..." : "清除精选"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-2xl border bg-card p-5">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="text-3xl font-bold mt-2">{value}</p>
    </div>
  );
}

function SectionTitle({
  title,
  subtitle,
}: {
  title: string;
  subtitle: string;
}) {
  return (
    <div className="mb-4">
      <h2 className="text-xl font-semibold">{title}</h2>
      <p className="text-sm text-muted-foreground mt-1">{subtitle}</p>
    </div>
  );
}

function EmptyBlock({ text }: { text: string }) {
  return (
    <div className="rounded-2xl border bg-card p-8 text-sm text-muted-foreground">
      {text}
    </div>
  );
}

function GroupManageCard({
  item,
  onQuickAnnouncement,
  onQuickFeatured,
}: {
  item: GroupDashboardItem;
  onQuickAnnouncement: () => void;
  onQuickFeatured: () => void;
}) {
  const group = item.group;

  return (
    <div className="rounded-2xl border bg-card p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-lg truncate">{group.name}</h3>
            <span className="text-xs rounded-full border px-2 py-0.5 text-muted-foreground">
              {item.role === "owner" ? "圈主" : "管理员"}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
            {group.description}
          </p>
          <div className="flex flex-wrap gap-3 text-xs text-muted-foreground mt-3">
            <span>👥 {group.member_count} 成员</span>
            <span>📝 {group.post_count} 帖子</span>
            <span>{group.featured_post_id ? "⭐ 已设精选" : "☆ 未设精选"}</span>
          </div>
        </div>
        <div className="flex flex-wrap justify-end gap-2 shrink-0">
          <Link href={`/posts/create?group_id=${group.id}`}>
            <Button size="sm" variant="outline">
              <PenSquare className="h-4 w-4 mr-1" />
              发帖
            </Button>
          </Link>
          <Button size="sm" variant="outline" onClick={onQuickAnnouncement}>
            <MessageSquareText className="h-4 w-4 mr-1" />
            发公告
          </Button>
          <Button size="sm" variant="outline" onClick={onQuickFeatured}>
            <Sparkles className="h-4 w-4 mr-1" />
            设精选
          </Button>
          <Link href={`/groups/${group.id}`}>
            <Button size="sm">
              <BarChart3 className="h-4 w-4 mr-1" />
              进入圈子
            </Button>
          </Link>
        </div>
      </div>

      {group.announcement && (
        <div className="mt-4 rounded-xl border bg-muted/40 p-3">
          <p className="text-xs uppercase tracking-[0.16em] text-muted-foreground mb-1">
            最新公告
          </p>
          <p className="text-sm text-muted-foreground line-clamp-3">
            {group.announcement}
          </p>
        </div>
      )}

      <div className="mt-4">
        <p className="text-xs uppercase tracking-[0.16em] text-muted-foreground mb-2">
          最新活跃成员
        </p>
        {item.active_members.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            暂时还没有活跃成员数据。
          </p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {item.active_members.map((member) => (
              <Link
                key={member.user_id}
                href={`/users/${member.user_id}`}
                className="rounded-full border px-3 py-1.5 text-sm hover:bg-muted/60 transition-colors"
              >
                {member.furry_name ||
                  member.username ||
                  member.user_id.slice(0, 8)}
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
