import type {
  DoudizhuMatchMode,
  DoudizhuPlayerRole,
  DoudizhuRoom,
  DoudizhuRoomStatus,
} from "@/lib/api-client";

export const DOUDIZHU_PRODUCT_NAME = "涂油斗地主";
export const DOUDIZHU_LOBBY_NAME = "涂油大厅";
export const DOUDIZHU_REPLAY_NAME = "涂油斗地主战报";

export function roomStatusLabel(status?: DoudizhuRoomStatus): string {
  switch (status) {
    case "bidding":
      return "叫分中";
    case "playing":
      return "出牌中";
    case "settlement":
      return "结算中";
    case "redeal":
      return "重新发牌";
    default:
      return "待开局";
  }
}

export function roomModeLabel(mode?: DoudizhuMatchMode): string {
  return mode === "demo_ai" ? "人机热身" : "三人联机";
}

export function seatLabel(seat?: number): string {
  switch (seat) {
    case 0:
      return "一号位";
    case 1:
      return "二号位";
    case 2:
      return "三号位";
    default:
      return "--";
  }
}

export function roleLabel(role?: DoudizhuPlayerRole): string {
  switch (role) {
    case "landlord":
      return "地主";
    case "farmer":
      return "农民";
    default:
      return "--";
  }
}

export function actionTypeLabel(actionType?: string): string {
  switch (actionType) {
    case "bid":
      return "叫分";
    case "auto_bid":
      return "托管叫分";
    case "play_cards":
      return "出牌";
    case "auto_play_cards":
      return "托管出牌";
    case "pass_turn":
      return "过牌";
    case "auto_pass_turn":
      return "托管过牌";
    case "timeout_auto_play":
      return "超时托管";
    case "landlord_assigned":
      return "地主确定";
    case "settlement":
      return "本局结算";
    default:
      return actionType ?? "操作";
  }
}

export function formatRemaining(
  target?: string,
  nowMs: number = Date.now(),
): string {
  if (!target) {
    return "--";
  }
  const targetMs = new Date(target).getTime();
  const remaining = Math.max(0, targetMs - nowMs);
  return `${Math.ceil(remaining / 1000)}s`;
}

export function connectionStatusLabel(status: string): string {
  switch (status) {
    case "connected":
      return "稳定";
    case "connecting":
      return "连线中";
    case "closed":
      return "重连中";
    default:
      return "待连接";
  }
}

export function buildRoomNotice(room: DoudizhuRoom): string {
  switch (room.status) {
    case "bidding":
      return "叫分开始了，稳住牌势，争下这一手的涂油地主。";
    case "playing":
      return room.match_mode === "demo_ai"
        ? "人机热身进行中，两名陪练已经入桌，按节奏把这一局打完。"
        : "三人联机进行中，轮到谁、谁能压，都以牌桌裁决为准。";
    case "settlement":
      return "这一局已经结算，可以继续留桌再来一局，或者直接查看战报。";
    case "redeal":
      return "这一轮没人抢下地主，马上重新发牌。";
    default:
      return room.match_mode === "demo_ai"
        ? "这是人机热身房，准备之后就能直接开局。"
        : "这是三人联机牌桌，凑齐三位牌手并全部准备后即可开局。";
  }
}
