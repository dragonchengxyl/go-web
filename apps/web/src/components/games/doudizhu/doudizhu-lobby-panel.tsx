"use client";

import { Bot, Users2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface DoudizhuLobbyPanelProps {
  playerName: string;
  roomTitle: string;
  errorMessage: string;
  onPlayerNameChange: (value: string) => void;
  onRoomTitleChange: (value: string) => void;
  onCreateDemoRoom: () => void;
  onCreateRoom: () => void;
}

const lobbyFeatureCards = [
  {
    title: "真牌桌视角",
    text: "中央就是牌桌本身，叫分、出牌、结算都围着桌面展开，不再像调试页面。",
  },
  {
    title: "单人也能开局",
    text: "直接开一把人机热身，不用等人齐，就能完整走一局牌。",
  },
  {
    title: "三人联机保留",
    text: "想和真人一起打时，照样可以从这里建桌并邀请其他牌手加入。",
  },
  {
    title: "战报随手可查",
    text: "一局结束后最近战报会立即更新，不用担心打完就没痕迹。",
  },
];

export function DoudizhuLobbyPanel({
  playerName,
  roomTitle,
  errorMessage,
  onPlayerNameChange,
  onRoomTitleChange,
  onCreateDemoRoom,
  onCreateRoom,
}: DoudizhuLobbyPanelProps) {
  return (
    <div className="mt-8 grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
      <div className="rounded-[30px] border border-white/10 bg-black/20 p-6">
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="ddz-player-name" className="text-white/85">
              牌桌昵称
            </Label>
            <Input
              id="ddz-player-name"
              value={playerName}
              onChange={(event) => onPlayerNameChange(event.target.value)}
              className="border-white/10 bg-black/30 text-white placeholder:text-emerald-50/35"
              placeholder="给自己起个上桌名称"
            />
          </div>

          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="ddz-room-title" className="text-white/85">
              牌局标题
            </Label>
            <Input
              id="ddz-room-title"
              value={roomTitle}
              onChange={(event) => onRoomTitleChange(event.target.value)}
              className="border-white/10 bg-black/30 text-white placeholder:text-emerald-50/35"
              placeholder="例如：夜场涂油局"
            />
          </div>
        </div>

        <div className="mt-5 flex flex-wrap gap-3">
          <Button
            onClick={onCreateDemoRoom}
            className="border-0 bg-[linear-gradient(135deg,#f4b63f_0%,#db5a3f_100%)] text-slate-950 hover:brightness-110"
          >
            <Bot className="mr-2 h-4 w-4" />
            开始人机热身
          </Button>
          <Button
            onClick={onCreateRoom}
            variant="outline"
            className="border-white/15 bg-transparent text-white hover:bg-white/8 hover:text-white"
          >
            <Users2 className="mr-2 h-4 w-4" />
            创建三人牌局
          </Button>
        </div>

        <div className="mt-5 rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-4 text-sm leading-7 text-emerald-50/70">
          一个人也能从这里直接开牌。人机热身会补齐两名陪练，三人联机则继续走完整房间链路。
        </div>

        {errorMessage && (
          <div className="mt-4 rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
            {errorMessage}
          </div>
        )}
      </div>

      <div className="rounded-[30px] border border-white/10 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.08),transparent_35%),linear-gradient(180deg,rgba(255,255,255,0.06),rgba(255,255,255,0.03))] p-6">
        <div className="grid gap-4 md:grid-cols-2">
          {lobbyFeatureCards.map((item) => (
            <div
              key={item.title}
              className="rounded-2xl border border-white/10 bg-black/20 px-4 py-4"
            >
              <div className="font-semibold text-white">{item.title}</div>
              <div className="mt-2 text-sm leading-6 text-emerald-50/70">
                {item.text}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
