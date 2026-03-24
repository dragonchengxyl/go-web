"use client";

import type { DoudizhuCard } from "@/lib/api-client";
import {
  cardRankLabel,
  cardSuitSymbol,
  isRedCard,
} from "@/lib/games/doudizhu/cards";
import { cn } from "@/lib/utils";

interface DoudizhuPlayCardProps {
  card: DoudizhuCard;
  selected: boolean;
  invalid?: boolean;
  pulse?: boolean;
  onClick: () => void;
}

export function DoudizhuPlayCard({
  card,
  selected,
  invalid,
  pulse,
  onClick,
}: DoudizhuPlayCardProps) {
  const red = isRedCard(card);

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "relative h-28 w-20 shrink-0 rounded-[22px] border bg-[linear-gradient(180deg,#fffaf1_0%,#ffffff_55%,#f0ece4_100%)] px-3 py-2 text-left shadow-[0_24px_42px_-20px_rgba(0,0,0,0.65)] transition-all",
        pulse ? "scale-[1.03]" : "",
        selected
          ? "-translate-y-5 rotate-[-2deg] border-amber-400 shadow-[0_26px_60px_-24px_rgba(255,122,69,0.72)] ring-2 ring-amber-300/45"
          : "border-slate-300/90 hover:-translate-y-2 hover:rotate-[-1deg] hover:shadow-[0_22px_55px_-24px_rgba(255,122,69,0.55)]",
        invalid ? "border-red-400 ring-2 ring-red-300/40" : "",
      )}
    >
      {/* 涂油按钮油渍纹理占位 */}
      <div className="absolute inset-x-2 top-1 h-5 rounded-full bg-white/80 blur-sm" />
      <div
        className={cn(
          "text-xs font-semibold",
          red ? "text-red-500" : "text-slate-900",
        )}
      >
        {cardSuitSymbol(card)}
      </div>
      <div
        className={cn(
          "mt-2 text-2xl font-black",
          red ? "text-red-500" : "text-slate-900",
        )}
      >
        {cardRankLabel(card)}
      </div>
      <div
        className={cn(
          "absolute bottom-2 right-2 text-lg font-semibold",
          red ? "text-red-400" : "text-slate-400",
        )}
      >
        {cardSuitSymbol(card)}
      </div>
    </button>
  );
}
