import Link from "next/link";
import { ArrowLeft, Maximize2 } from "lucide-react";
import {
  DouDizhuPlayStage,
  roomPagePath,
} from "@/components/games/doudizhu-play-stage";

export default function DoudizhuRoomPage({
  params,
}: {
  params: { roomId: string };
}) {
  return (
    <main className="min-h-screen bg-[#07131b] px-3 pb-4 pt-4 text-white md:px-4">
      <section className="mx-auto max-w-[1800px]">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <Link
            href="/games/dou-dizhu/play"
            className="inline-flex items-center gap-2 text-sm text-slate-400 transition-colors hover:text-white"
          >
            <ArrowLeft className="h-4 w-4" />
            返回大厅
          </Link>

          <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 text-sm text-slate-300">
            <Maximize2 className="h-4 w-4 text-amber-300" />
            房间页
            <span className="text-white/50">{roomPagePath(params.roomId)}</span>
          </div>
        </div>

        <DouDizhuPlayStage immersive fixedRoomId={params.roomId} />
      </section>
    </main>
  );
}
