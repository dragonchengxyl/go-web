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
    <div className="mt-8">
      <div className="mx-auto max-w-3xl rounded-[30px] border border-white/10 bg-black/20 p-6">
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

        {errorMessage && (
          <div className="mt-4 rounded-2xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-100">
            {errorMessage}
          </div>
        )}
      </div>
    </div>
  );
}
