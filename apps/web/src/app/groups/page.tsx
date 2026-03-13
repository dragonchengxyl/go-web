"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { apiClient, Group } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

function GroupCard({ group }: { group: Group }) {
  return (
    <Link href={`/groups/${group.id}`}>
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        className="group rounded-2xl border border-white/10 bg-white/5 p-5 hover:bg-white/10 transition-colors cursor-pointer"
      >
        <div className="flex items-start gap-3">
          {group.avatar_key ? (
            <img
              src={group.avatar_key}
              alt={group.name}
              className="w-12 h-12 rounded-xl object-cover shrink-0"
            />
          ) : (
            <div className="w-12 h-12 rounded-xl bg-purple-600/40 flex items-center justify-center text-lg shrink-0">
              🐾
            </div>
          )}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-white truncate group-hover:text-purple-300 transition-colors">
                {group.name}
              </h3>
              {group.privacy === "private" && (
                <span className="text-xs px-1.5 py-0.5 rounded bg-white/10 text-white/40">
                  🔒 私密
                </span>
              )}
            </div>
            <p className="mt-0.5 text-sm text-white/50 line-clamp-2">
              {group.description}
            </p>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap gap-3 text-xs text-white/40">
          <span>👥 {group.member_count} 成员</span>
          <span>📝 {group.post_count} 帖子</span>
        </div>

        {group.tags?.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {group.tags.map((t) => (
              <span
                key={t}
                className="text-xs px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300"
              >
                #{t}
              </span>
            ))}
          </div>
        )}
      </motion.div>
    </Link>
  );
}

type Tab = "discover" | "mine";

export default function GroupsPage() {
  const [tab, setTab] = useState<Tab>("discover");
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");

  useEffect(() => {
    setLoading(true);
    const fetch =
      tab === "discover"
        ? apiClient.listGroups({ search, page: 1, page_size: 24 })
        : apiClient.myGroups(1, 24);

    fetch
      .then((res) => setGroups(res.groups ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [tab, search]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-950 via-purple-950/30 to-slate-950 px-4 py-10">
      <div className="mx-auto max-w-4xl">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-3xl font-bold text-white">兴趣圈子</h1>
            <p className="mt-1 text-white/50">找到志同道合的 Furry 伙伴</p>
          </div>
          <div className="flex items-center gap-2">
            <Link href="/groups/manage">
              <Button
                variant="outline"
                className="border-white/20 text-white/80 hover:bg-white/10"
              >
                我的圈子
              </Button>
            </Link>
            <Link href="/groups/create">
              <Button className="bg-purple-600 hover:bg-purple-500 text-white">
                + 创建圈子
              </Button>
            </Link>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 p-1 rounded-xl bg-white/5 w-fit mb-6">
          {(["discover", "mine"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={cn(
                "px-4 py-1.5 rounded-lg text-sm transition-colors",
                tab === t
                  ? "bg-purple-600 text-white"
                  : "text-white/50 hover:text-white",
              )}
            >
              {t === "discover" ? "发现圈子" : "我的圈子"}
            </button>
          ))}
        </div>

        {/* Search */}
        {tab === "discover" && (
          <form onSubmit={handleSearch} className="flex gap-2 mb-6">
            <input
              className="flex-1 rounded-xl bg-white/5 border border-white/10 px-4 py-2 text-white placeholder-white/30 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="搜索圈子名称…"
            />
            <Button
              type="submit"
              variant="outline"
              className="border-white/20 text-white/70"
            >
              搜索
            </Button>
          </form>
        )}

        {loading ? (
          <div className="grid gap-4 sm:grid-cols-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div
                key={i}
                className="h-32 rounded-2xl bg-white/5 animate-pulse"
              />
            ))}
          </div>
        ) : groups.length === 0 ? (
          <div className="text-center py-20 text-white/40">
            {tab === "mine" ? "还没有加入任何圈子" : "没有找到相关圈子"}
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2">
            {groups.map((g) => (
              <GroupCard key={g.id} group={g} />
            ))}
          </div>
        )}
      </div>
    </main>
  );
}
