"use client";

import Link from "next/link";
import type {
  DoudizhuLeaderboardEntry,
  DoudizhuMatchSummary,
} from "@/lib/api-client";
import {
  roomModeLabel,
  seatLabel,
} from "@/lib/games/doudizhu/presenter";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";

interface DoudizhuMatchLinkCardProps {
  match: DoudizhuMatchSummary;
  mine?: boolean;
}

export function DoudizhuMatchLinkCard({
  match,
  mine = false,
}: DoudizhuMatchLinkCardProps) {
  return (
    <Link
      href={`/games/dou-dizhu/matches/${match.match_id}`}
      className={cn(
        "block rounded-[24px] border px-4 py-4 transition-colors",
        mine
          ? "border-sky-300/15 bg-sky-300/10 hover:border-sky-300/30 hover:bg-sky-300/14"
          : "border-white/10 bg-black/20 hover:border-white/20 hover:bg-white/[0.05]",
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-semibold text-white">{match.room_title}</span>
            <Badge className="border-white/15 bg-white/8 text-white">
              {roomModeLabel(match.match_mode)}
            </Badge>
          </div>
          <div className="mt-1 text-sm text-slate-400">
            {new Date(match.finished_at).toLocaleString("zh-CN")} · 地主{" "}
            {seatLabel(match.landlord_seat)}
          </div>
        </div>

        <div className="text-right text-sm text-slate-300">
          <div>{match.winner_side === "landlord" ? "地主胜" : "农民胜"}</div>
          <div className="mt-1">倍率 x{match.multiplier}</div>
        </div>
      </div>
    </Link>
  );
}

export function DoudizhuLeaderboardCard({
  entry,
}: {
  entry: DoudizhuLeaderboardEntry;
}) {
  return (
    <div className="rounded-[22px] border border-white/10 bg-black/20 px-4 py-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="font-semibold text-white">
            #{entry.rank} {entry.display_name}
          </div>
          <div className="mt-1 text-sm text-slate-400">
            {entry.matches} 局 · 胜 {entry.wins} 局
          </div>
        </div>
        <div className="text-right">
          <div className="text-xs uppercase tracking-[0.24em] text-slate-500">
            Total
          </div>
          <div className="mt-1 text-2xl font-black tracking-tight text-white">
            {entry.total_score}
          </div>
        </div>
      </div>
    </div>
  );
}
