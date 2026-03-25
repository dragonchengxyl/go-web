"use client";

import { Crown } from "lucide-react";
import type { DoudizhuRoomPlayer } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { seatLabel } from "@/lib/games/doudizhu/presenter";
import { Badge } from "@/components/ui/badge";

interface DoudizhuSeatProps {
  player: DoudizhuRoomPlayer | null;
  position: "left" | "right" | "bottom";
  isCurrentTurn: boolean;
  isCurrentBidder: boolean;
  isLandlord: boolean;
  isMe: boolean;
}

export function DoudizhuSeat({
  player,
  position,
  isCurrentTurn,
  isCurrentBidder,
  isLandlord,
  isMe,
}: DoudizhuSeatProps) {
  if (!player) {
    return null;
  }

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-[30px] border px-4 py-4 backdrop-blur-md",
        isMe
          ? "border-amber-300/35 bg-[linear-gradient(180deg,rgba(255,203,120,0.16),rgba(255,203,120,0.06))]"
          : "border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.09),rgba(255,255,255,0.03))]",
        position === "bottom"
          ? "shadow-[0_28px_90px_-34px_rgba(0,0,0,0.65)]"
          : "",
      )}
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.08),transparent_35%)]" />
      <div className="relative flex items-start gap-3">
        <div
          className={cn(
            "flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border text-lg font-black",
            isLandlord
              ? "border-red-300/30 bg-red-400/15 text-red-50"
              : "border-white/10 bg-black/25 text-white",
          )}
        >
          {isLandlord ? <Crown className="h-5 w-5" /> : player.name.slice(0, 1)}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="truncate text-lg font-semibold text-white">
              {player.name}
            </span>
            <Badge className="border-white/15 bg-black/25 text-white">
              {seatLabel(player.seat)}
            </Badge>
            {isMe && (
              <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                我
              </Badge>
            )}
            {player.is_host && (
              <Badge className="border-sky-300/20 bg-sky-300/10 text-sky-100">
                房主
              </Badge>
            )}
            {player.is_bot && (
              <Badge className="border-amber-300/20 bg-amber-300/10 text-amber-100">
                陪练
              </Badge>
            )}
            {isLandlord && (
              <Badge className="border-red-400/20 bg-red-400/10 text-red-100">
                涂油地主
              </Badge>
            )}
            {(isCurrentTurn || isCurrentBidder) && (
              <Badge className="border-emerald-300/20 bg-emerald-300/10 text-emerald-100">
                <span className="mr-1 inline-block h-2 w-2 rounded-full bg-emerald-300" />
                {isCurrentTurn ? "行动中" : "叫分中"}
              </Badge>
            )}
          </div>

          <div className="mt-2 flex flex-wrap items-center justify-between gap-3 text-xs text-slate-300">
            <div>剩余 {player.card_count} 张</div>
            <div>{player.connected ? "在线" : "离线"}</div>
            <div>{player.auto_play ? "托管中" : "手动操作"}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
