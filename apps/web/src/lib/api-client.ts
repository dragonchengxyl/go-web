const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "/api/v1";
const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  request_id?: string;
  timestamp?: number;
}

export type ModerationStatus = "pending" | "approved" | "blocked";

export interface Post {
  id: string;
  author_id: string;
  group_id?: string;
  title?: string;
  content: string;
  media_urls?: string[];
  tags?: string[];
  content_labels?: Record<string, boolean>;
  visibility: "public" | "followers_only" | "private";
  moderation_status: ModerationStatus;
  like_count: number;
  comment_count: number;
  is_pinned: boolean;
  created_at: string;
  updated_at: string;
  author_username?: string;
  author_avatar_key?: string;
  group_name?: string;
  is_bookmarked_by_me?: boolean;
  is_liked_by_me?: boolean;
}

export interface OSSUploadPolicy {
  host: string;
  OSSAccessKeyId: string;
  policy: string;
  signature: string;
  expire: number;
  dir: string;
}

export interface UserFollow {
  follower_id: string;
  followee_id: string;
  created_at: string;
}

export interface FollowStats {
  user_id: string;
  follower_count: number;
  following_count: number;
}

export interface Conversation {
  id: string;
  type: "direct" | "group";
  name?: string;
  members: string[];
  created_at: string;
  updated_at: string;
  last_message?: Message;
  unread_count?: number;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  media_url?: string;
  is_read: boolean;
  created_at: string;
  sender_username?: string;
  sender_avatar_key?: string;
}

export interface TipOrder {
  id: string;
  order_no: string;
  user_id: string;
  status: string;
  total_cents: number;
  currency: string;
  metadata?: {
    type: string;
    to_user_id: string;
    message: string;
  };
  created_at: string;
}

export interface Comment {
  id: string;
  user_id: string;
  commentable_type: string;
  commentable_id: string;
  parent_id?: string;
  content: string;
  is_edited: boolean;
  like_count: number;
  reply_count: number;
  created_at: string;
  updated_at: string;
  author_username?: string;
  author_avatar_key?: string;
}

export interface Notification {
  id: string;
  user_id: string;
  actor_id?: string;
  type: "like" | "comment" | "follow" | "tip" | "system";
  target_id?: string;
  target_type?: string;
  is_read: boolean;
  created_at: string;
  actor_username?: string;
  actor_avatar_key?: string;
}

export interface AssistantCard {
  ref?: string;
  kind: "page" | "post" | "group" | "event" | "user" | "tag";
  title: string;
  summary: string;
  href: string;
  meta?: string;
  reason?: string;
  source?: string;
}

export interface AssistantInsight {
  kind:
    | "draft_polish"
    | "title_options"
    | "tag_suggestions"
    | "visibility_suggestion"
    | "group_atmosphere"
    | "rules_summary"
    | "join_suggestion"
    | "posting_ideas"
    | "event_summary"
    | "fit_assessment"
    | "signup_notes"
    | "preparation_checklist";
  title: string;
  summary?: string;
  bullets?: string[];
}

export interface AssistantChatMessage {
  id?: string;
  role: "user" | "assistant";
  content: string;
  cards?: AssistantCard[];
  insights?: AssistantInsight[];
  created_at?: string;
  provider?: string;
  fallback?: boolean;
  intent?: string;
  source_counts?: Record<string, number>;
  feedback_value?: "helpful" | "unhelpful";
}

export interface AssistantPageContextPayload {
  path?: string;
  kind: string;
  title: string;
  summary?: string;
  prompt_hints?: string[];
  fields?: Record<string, string>;
}

export interface AssistantMeta {
  query: string;
  provider: string;
  fallback: boolean;
  intent?: string;
  intent_label?: string;
  source_counts?: Record<string, number>;
  conversation_id?: string;
  response_id?: string;
  cards: AssistantCard[];
  insights?: AssistantInsight[];
}

export interface AssistantConversation {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
  updated_at: string;
  last_message_preview?: string;
  last_role?: string;
}

export interface AssistantSettings {
  enabled: boolean;
  persona_name: string;
  system_prompt: string;
  max_context_items: number;
  include_pages: boolean;
  include_posts: boolean;
  include_users: boolean;
  include_tags: boolean;
  include_groups: boolean;
  include_events: boolean;
  updated_at?: string;
  updated_by?: string;
}

export interface AssistantFeedbackInput {
  response_id: string;
  conversation_id?: string;
  value: "helpful" | "unhelpful";
  query: string;
  reply_excerpt: string;
  provider?: string;
  intent?: string;
  fallback?: boolean;
  page_path?: string;
  source_counts?: Record<string, number>;
  cards?: AssistantCard[];
}

export interface AssistantOverviewData {
  overview: {
    indexed_documents: number;
    documents_by_source: Record<string, number>;
    last_indexed_at?: string;
    media_cache_entries: number;
    feedback_helpful: number;
    feedback_unhelpful: number;
    embedding_configured: boolean;
    embedding_model?: string;
    vision_configured: boolean;
    vision_model?: string;
    retrieval_limit: number;
    vector_scan_limit: number;
    sync_interval_sec: number;
  };
  metrics: {
    retrievals_total: number;
    last_retrieval_duration_ms: number;
    last_retrieved_documents: number;
    first_token_observed_total: number;
    last_first_token_latency_ms: number;
    fallback_total: number;
    fallback_by_reason: Record<string, number>;
    feedback_total: number;
    feedback_by_value: Record<string, number>;
    last_index_sync_duration_ms: number;
    last_indexed_documents: number;
    last_index_synced_at?: string;
    last_index_error?: string;
    multimodal_requests_total: number;
    multimodal_cache_hits: number;
    multimodal_retry_total: number;
    multimodal_fallback_total: number;
    last_multimodal_latency_ms: number;
    last_multimodal_error?: string;
    chat_circuit_state?: string;
    vision_circuit_state?: string;
  };
}

export interface MediaAnalysisItem {
  id: string;
  media_url: string;
  alt_text: string;
  tags?: string[];
  image_summary?: string;
  moderation_summary?: string;
  risk_level?: "low" | "medium" | "high" | string;
  safety_notes?: string[];
  provider?: string;
  model?: string;
  fallback: boolean;
  cached_at: string;
  expires_at: string;
}

export interface MediaAnalysisResult {
  items: MediaAnalysisItem[];
  fallback: boolean;
  provider: string;
  circuit_state?: string;
}

export type AudioJobTaskType =
  | "ai_music"
  | "voice_convert"
  | "voice_enhance"
  | "audio_master";

export type AudioJobStatus =
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "dead_lettered";

export interface AudioJob {
  id: string;
  user_id: string;
  title: string;
  task_type: AudioJobTaskType;
  status: AudioJobStatus;
  source_audio_url?: string;
  reference_audio_url?: string;
  prompt?: string;
  params?: Record<string, unknown>;
  result?: Record<string, unknown>;
  error_message?: string;
  attempt_count: number;
  max_attempts: number;
  created_at: string;
  updated_at: string;
  started_at?: string;
  finished_at?: string;
  next_retry_at?: string;
  last_error_at?: string;
  dead_lettered_at?: string;
}

export type AudioWorkVisibility = "public" | "private";

export interface AudioWork {
  id: string;
  author_id: string;
  source_job_id: string;
  title: string;
  description?: string;
  cover_image_url?: string;
  audio_url: string;
  duration_sec: number;
  visibility: AudioWorkVisibility;
  moderation_status: ModerationStatus;
  moderation_note?: string;
  like_count: number;
  comment_count: number;
  tags?: string[];
  waveform_preview?: number[];
  metadata?: Record<string, unknown>;
  published_at: string;
  created_at: string;
  updated_at: string;
  author_username?: string;
}

export type HexBlitzRoomStatus =
  | "waiting"
  | "countdown"
  | "running"
  | "finished";

export interface HexBlitzRoomPlayer {
  session_id: string;
  user_id?: string;
  name: string;
  ready: boolean;
  connected: boolean;
  is_host: boolean;
  score: number;
  joined_at: string;
  updated_at: string;
}

export interface HexBlitzRoom {
  id: string;
  code: string;
  game_slug: string;
  title: string;
  status: HexBlitzRoomStatus;
  host_session_id: string;
  countdown_sec: number;
  round_duration_sec: number;
  player_count: number;
  ready_count: number;
  countdown_started_at?: string;
  started_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
  players: HexBlitzRoomPlayer[];
}

export type HexBlitzTileColor =
  | "ember"
  | "lagoon"
  | "mint"
  | "sun"
  | "violet";

export type HexBlitzTileSpecial = "none" | "spark" | "burst";

export interface HexBlitzTile {
  id: string;
  q: number;
  r: number;
  color: HexBlitzTileColor;
  special: HexBlitzTileSpecial;
}

export interface HexBlitzBoardState {
  session_id: string;
  match_id: string;
  phase: HexBlitzRoomStatus;
  seed: number;
  score: number;
  combo: number;
  best_combo: number;
  moves: number;
  last_gain: number;
  last_cleared: number;
  message: string;
  updated_at: string;
  tiles: HexBlitzTile[];
}

export interface HexBlitzMoveResult {
  session_id: string;
  match_id: string;
  tile_id: string;
  cleared_count: number;
  gained_score: number;
  score: number;
  combo: number;
  best_combo: number;
  moves: number;
  message: string;
  updated_at: string;
}

export interface HexBlitzReplay {
  match: {
    id: string;
    room_id: string;
    room_code: string;
    room_title: string;
    game_slug: string;
    seed: number;
    started_at: string;
    finished_at: string;
    duration_sec: number;
    created_at: string;
  };
  results: HexBlitzMatchResult[];
  events: Array<{
    id: string;
    match_id: string;
    session_id: string;
    user_id?: string;
    player_name: string;
    display_name: string;
    move_index: number;
    tile_id: string;
    cleared_count: number;
    gained_score: number;
    score_after: number;
    combo_after: number;
    occurred_at: string;
  }>;
  players: Array<{
    session_id: string;
    user_id?: string;
    player_name: string;
    display_name: string;
    frames: Array<{
      step: number;
      move_index: number;
      event?: {
        id: string;
        match_id: string;
        session_id: string;
        user_id?: string;
        player_name: string;
        display_name: string;
        move_index: number;
        tile_id: string;
        cleared_count: number;
        gained_score: number;
        score_after: number;
        combo_after: number;
        occurred_at: string;
      };
      board: HexBlitzBoardState;
    }>;
  }>;
}

export interface HexBlitzLeaderboardEntry {
  rank: number;
  user_id?: string;
  player_name: string;
  display_name: string;
  best_score: number;
  matches: number;
  last_played: string;
}

export interface HexBlitzMatchResult {
  id: string;
  match_id: string;
  session_id: string;
  room_id: string;
  room_code: string;
  room_title: string;
  user_id?: string;
  player_name: string;
  display_name: string;
  score: number;
  rank: number;
  started_at: string;
  finished_at: string;
  created_at: string;
}

export interface HexBlitzMatchSummary {
  match_id: string;
  room_id: string;
  room_code: string;
  room_title: string;
  game_slug: string;
  finished_at: string;
  duration_sec: number;
  winner_name: string;
  winner_score: number;
  player_count: number;
  top_results: HexBlitzMatchResult[];
}

export interface AdminGameMetrics {
  room_events_total: number;
  score_reports_total: number;
  rejected_score_reports: number;
  active_rooms: number;
  active_players: number;
  active_connections: number;
  matches_finished_total: number;
  rooms_by_status: Record<string, number>;
  room_events_by_type: Record<string, number>;
  score_report_reasons: Record<string, number>;
}

export interface AdminGameOverview {
  metrics: AdminGameMetrics;
  rooms: HexBlitzRoom[];
  leaderboard: HexBlitzLeaderboardEntry[];
  recent_matches: HexBlitzMatchSummary[];
}

export type DoudizhuRoomStatus =
  | "waiting"
  | "bidding"
  | "playing"
  | "settlement"
  | "redeal";

export type DoudizhuMatchMode = "pvp" | "demo_ai";

export type DoudizhuPlayerRole = "farmer" | "landlord";

export type DoudizhuComboType =
  | "single"
  | "pair"
  | "triple"
  | "triple_with_single"
  | "triple_with_pair"
  | "straight"
  | "straight_pairs"
  | "airplane"
  | "airplane_with_single"
  | "airplane_with_pair"
  | "four_with_two_single"
  | "four_with_two_pair"
  | "bomb"
  | "rocket";

export interface DoudizhuCard {
  suit: "spade" | "heart" | "club" | "diamond" | "joker";
  rank: number;
}

export interface DoudizhuCombo {
  type: DoudizhuComboType;
  main_rank: number;
  sequence_length: number;
  total_cards: number;
}

export interface DoudizhuBidRecord {
  seat: number;
  score: number;
  at: string;
}

export interface DoudizhuActionRecord {
  seat: number;
  action_type: string;
  cards?: DoudizhuCard[];
  combo?: DoudizhuCombo;
  at: string;
  message?: string;
  actor_name?: string;
  winning_side?: DoudizhuPlayerRole;
}

export interface DoudizhuRoomPlayer {
  session_id: string;
  user_id?: string;
  seat: number;
  name: string;
  ready: boolean;
  connected: boolean;
  is_host: boolean;
  is_bot: boolean;
  bot_level?: string;
  auto_play: boolean;
  card_count: number;
  role: DoudizhuPlayerRole;
  joined_at: string;
  updated_at: string;
}

export interface DoudizhuRoom {
  id: string;
  code: string;
  title: string;
  match_mode: DoudizhuMatchMode;
  status: DoudizhuRoomStatus;
  host_session_id: string;
  player_count: number;
  ready_count: number;
  current_bidder?: number;
  highest_bid: number;
  highest_bidder?: number;
  landlord?: number;
  current_turn?: number;
  last_play?: DoudizhuCombo;
  last_play_seat?: number;
  winning_side?: DoudizhuPlayerRole;
  turn_expires_at?: string;
  bottom_cards?: DoudizhuCard[];
  players: DoudizhuRoomPlayer[];
  bid_history?: DoudizhuBidRecord[];
  recent_actions?: DoudizhuActionRecord[];
  created_at: string;
  updated_at: string;
}

export interface DoudizhuPrivateState {
  session_id: string;
  room_id: string;
  status: DoudizhuRoomStatus;
  hand: DoudizhuCard[];
  can_pass: boolean;
  role: DoudizhuPlayerRole;
  bottom_cards?: DoudizhuCard[];
  last_play?: DoudizhuCombo;
  last_play_seat?: number;
  turn_expires_at?: string;
}

export interface DoudizhuActionResult {
  action_type: string;
  seat: number;
  actor_name: string;
  combo?: DoudizhuCombo;
  cards?: DoudizhuCard[];
  hand_count: number;
  next_turn?: number;
  status: DoudizhuRoomStatus;
  winning_side?: DoudizhuPlayerRole;
  message?: string;
}

export interface DoudizhuMatchPlayerResult {
  id: string;
  match_id: string;
  session_id: string;
  user_id?: string;
  is_bot: boolean;
  bot_level?: string;
  seat: number;
  player_name: string;
  display_name: string;
  role: DoudizhuPlayerRole;
  bid_score: number;
  cards_left: number;
  is_winner: boolean;
  score_delta: number;
  created_at: string;
}

export interface DoudizhuMatchSummary {
  match_id: string;
  room_id: string;
  room_code: string;
  room_title: string;
  match_mode: DoudizhuMatchMode;
  finished_at: string;
  winner_side: DoudizhuPlayerRole;
  landlord_seat: number;
  multiplier: number;
  player_count: number;
  top_results: DoudizhuMatchPlayerResult[];
}

export interface DoudizhuLeaderboardEntry {
  rank: number;
  user_id?: string;
  player_name: string;
  display_name: string;
  matches: number;
  wins: number;
  total_score: number;
  last_played: string;
}

export interface DoudizhuReplay {
  match: {
    id: string;
    room_id: string;
    room_code: string;
    room_title: string;
    match_mode: DoudizhuMatchMode;
    started_at: string;
    finished_at: string;
    landlord_seat: number;
    winner_side: DoudizhuPlayerRole;
    multiplier: number;
    bomb_count: number;
    spring: boolean;
    anti_spring: boolean;
    created_at: string;
  };
  results: DoudizhuMatchPlayerResult[];
  events: Array<{
    id: string;
    match_id: string;
    turn_no: number;
    action_index: number;
    session_id: string;
    user_id?: string;
    player_name: string;
    display_name: string;
    seat: number;
    action_type: string;
    cards?: DoudizhuCard[];
    combo?: DoudizhuCombo;
    multiplier_after: number;
    occurred_at: string;
  }>;
}

export interface AdminAIToolSection {
  title: string;
  bullets?: string[];
}

export interface AdminAIToolDraft {
  label: string;
  content: string;
}

export interface AdminAIToolResult {
  run_id: string;
  tool: string;
  title: string;
  summary: string;
  sections?: AdminAIToolSection[];
  drafts?: AdminAIToolDraft[];
  fallback: boolean;
  provider: string;
  generated_at: string;
}

interface AssistantStreamHandlers {
  signal?: AbortSignal;
  onMeta?: (meta: AssistantMeta) => void;
  onToken?: (token: string) => void;
  onDone?: () => void;
  onError?: (message: string) => void;
}

function parseSSEBlock(block: string): { event: string; data: string } | null {
  const lines = block.split("\n");
  let event = "message";
  const dataLines: string[] = [];

  for (const line of lines) {
    if (line.startsWith("event:")) {
      event = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      dataLines.push(line.slice(5).trim());
    }
  }

  if (dataLines.length === 0) return null;
  return { event, data: dataLines.join("\n") };
}

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;
  private refreshing: Promise<void> | null = null;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("access_token");
    }
  }

  setToken(token: string | null) {
    this.token = token;
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem("access_token", token);
        document.cookie = `_auth=1; path=/; max-age=${7 * 24 * 3600}; SameSite=Lax`;
      } else {
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        document.cookie = "_auth=; path=/; max-age=0";
      }
    }
  }

  setRefreshToken(token: string) {
    if (typeof window !== "undefined") {
      localStorage.setItem("refresh_token", token);
    }
  }

  getToken(): string | null {
    return this.token;
  }

  private async tryRefresh(): Promise<boolean> {
    const refreshToken =
      typeof window !== "undefined"
        ? localStorage.getItem("refresh_token")
        : null;
    if (!refreshToken) return false;

    try {
      const res = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      const data: ApiResponse<{ access_token: string; refresh_token: string }> =
        await res.json();
      if (data.code !== 0) return false;
      this.setToken(data.data.access_token);
      this.setRefreshToken(data.data.refresh_token);
      return true;
    } catch {
      return false;
    }
  }

  private async request<T = any>(
    endpoint: string,
    options: RequestInit = {},
    isRetry = false,
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    if (response.status === 401 && !isRetry) {
      if (!this.refreshing) {
        this.refreshing = this.tryRefresh().then((ok) => {
          this.refreshing = null;
          if (!ok) {
            this.setToken(null);
            if (typeof window !== "undefined") {
              window.location.href = "/login";
            }
          }
        });
      }
      await this.refreshing;
      if (this.token) {
        return this.request<T>(endpoint, options, true);
      }
      throw new Error("登录已过期，请重新登录");
    }

    const data: ApiResponse<T> = await response.json();

    if (data.code !== 0) {
      throw new Error(data.message || "Request failed");
    }

    return data.data;
  }

  async get<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "GET" });
  }

  async post<T = any>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async put<T = any>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "DELETE" });
  }

  // ── Auth ──────────────────────────────────────────────────────────────

  async login(email: string, password: string) {
    return this.post<{
      access_token: string;
      refresh_token: string;
      user: any;
    }>("/auth/login", {
      email,
      password,
    });
  }

  async register(username: string, email: string, password: string) {
    return this.post<{
      access_token: string;
      refresh_token: string;
      user: any;
    }>("/auth/register", {
      username,
      email,
      password,
    });
  }

  async logout() {
    return this.post<void>("/auth/logout");
  }

  async verifyEmail(token: string) {
    return this.post<{ message: string }>("/auth/verify-email", { token });
  }

  async resendVerification() {
    return this.post<{ message: string }>("/auth/resend-verification");
  }

  // ── Posts ─────────────────────────────────────────────────────────────

  async getFeed(page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{
      posts: Post[];
      total: number;
      page: number;
      size: number;
    }>(`/feed?${q}`);
  }

  async getExplore(page?: number, pageSize?: number, tag?: string) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    if (tag) q.set("tag", tag);
    return this.get<{
      posts: Post[];
      total: number;
      page: number;
      size: number;
    }>(`/explore?${q}`);
  }

  async getHotTags(): Promise<string[]> {
    return this.get<string[]>("/explore/tags");
  }

  async getPost(id: string) {
    return this.get<Post>(`/posts/${id}`);
  }

  async createPost(data: {
    title?: string;
    content: string;
    media_urls?: string[];
    tags?: string[];
    visibility?: string;
    group_id?: string;
    is_ai_generated?: boolean;
  }) {
    return this.post<Post>("/posts", data);
  }

  async getOSSPolicy(purpose?: string): Promise<OSSUploadPolicy> {
    return this.post<OSSUploadPolicy>("/upload/oss-policy", {
      purpose: purpose ?? "post",
    });
  }

  async updatePost(
    id: string,
    data: {
      title?: string;
      content: string;
      media_urls?: string[];
      tags?: string[];
      visibility?: string;
    },
  ) {
    return this.put<Post>(`/posts/${id}`, data);
  }

  async deletePost(id: string) {
    return this.delete<void>(`/posts/${id}`);
  }

  async likePost(id: string) {
    return this.post<void>(`/posts/${id}/like`);
  }

  async unlikePost(id: string) {
    return this.delete<void>(`/posts/${id}/like`);
  }

  async getComments(postId: string, page?: number, pageSize?: number) {
    return this.getCommentsByTarget("post", postId, page, pageSize);
  }

  async getCommentsByTarget(
    commentableType: "post" | "audio_work",
    commentableId: string,
    page?: number,
    pageSize?: number,
  ) {
    const q = new URLSearchParams({
      commentable_type: commentableType,
      commentable_id: commentableId,
    });
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{
      comments: Comment[];
      total: number;
      page: number;
      size: number;
    }>(`/comments?${q}`);
  }

  async createComment(postId: string, content: string) {
    return this.createCommentForTarget("post", postId, content);
  }

  async createCommentForTarget(
    commentableType: "post" | "audio_work",
    commentableId: string,
    content: string,
  ) {
    return this.post<Comment>("/comments", {
      commentable_type: commentableType,
      commentable_id: commentableId,
      content,
    });
  }

  async getUserPosts(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{
      posts: Post[];
      total: number;
      page: number;
      size: number;
    }>(`/users/${userId}/posts?${q}`);
  }

  async getGroupPosts(
    groupId: string,
    page = 1,
    pageSize = 20,
    params?: { sort?: "latest" | "hot"; tag?: string },
  ) {
    const q = new URLSearchParams({
      page: String(page),
      page_size: String(pageSize),
    });
    if (params?.sort) q.set("sort", params.sort);
    if (params?.tag) q.set("tag", params.tag);
    return this.get<{
      posts: Post[];
      total: number;
      page: number;
      size: number;
    }>(`/groups/${groupId}/posts?${q.toString()}`);
  }

  async getGroupHighlights(groupId: string) {
    return this.get<{ posts: Post[] }>(`/groups/${groupId}/highlights`);
  }

  async getGroupPostTags(groupId: string) {
    return this.get<string[]>(`/groups/${groupId}/tags`);
  }

  async getGroupAnnouncements(groupId: string, page = 1, pageSize = 20) {
    return this.get<{
      announcements: GroupAnnouncement[];
      total: number;
      page: number;
      size: number;
    }>(`/groups/${groupId}/announcements?page=${page}&page_size=${pageSize}`);
  }

  async pinGroupPost(groupId: string, postId: string) {
    return this.post<{ message: string }>(
      `/groups/${groupId}/posts/${postId}/pin`,
    );
  }

  async unpinGroupPost(groupId: string, postId: string) {
    return this.delete<{ message: string }>(
      `/groups/${groupId}/posts/${postId}/pin`,
    );
  }

  async setGroupFeaturedPost(groupId: string, postId?: string) {
    return this.put<Group>(`/groups/${groupId}/featured-post`, {
      post_id: postId ?? "",
    });
  }

  // ── Games / Hex Blitz ─────────────────────────────────────────────────

  async getHexBlitzRooms() {
    return this.get<{ rooms: HexBlitzRoom[] }>("/games/hex-blitz/rooms");
  }

  async getHexBlitzRoom(roomId: string) {
    return this.get<HexBlitzRoom>(`/games/hex-blitz/rooms/${roomId}`);
  }

  async createHexBlitzRoom(data: {
    title?: string;
    player_name: string;
    session_id: string;
  }) {
    return this.post<HexBlitzRoom>("/games/hex-blitz/rooms", data);
  }

  async getHexBlitzLeaderboard(limit = 10) {
    return this.get<{ entries: HexBlitzLeaderboardEntry[] }>(
      `/games/hex-blitz/leaderboard?limit=${limit}`
    );
  }

  async getHexBlitzRecentMatches(limit = 8) {
    return this.get<{ matches: HexBlitzMatchSummary[] }>(
      `/games/hex-blitz/matches?limit=${limit}`
    );
  }

  async getHexBlitzReplay(matchId: string) {
    return this.get<HexBlitzReplay>(`/games/hex-blitz/matches/${matchId}/replay`);
  }

  async getMyHexBlitzRecentMatches(limit = 8) {
    return this.get<{ matches: HexBlitzMatchSummary[] }>(
      `/games/hex-blitz/matches/me?limit=${limit}`
    );
  }

  // ── Games / Dou Dizhu ────────────────────────────────────────────────

  async getDoudizhuRooms() {
    return this.get<{ rooms: DoudizhuRoom[] }>("/games/dou-dizhu/rooms");
  }

  async getDoudizhuRoom(roomId: string) {
    return this.get<DoudizhuRoom>(`/games/dou-dizhu/rooms/${roomId}`);
  }

  async createDoudizhuRoom(data: {
    title?: string;
    player_name: string;
    session_id: string;
  }) {
    return this.post<DoudizhuRoom>("/games/dou-dizhu/rooms", data);
  }

  async createDoudizhuDemoRoom(data: {
    title?: string;
    player_name: string;
    session_id: string;
  }) {
    return this.post<DoudizhuRoom>("/games/dou-dizhu/rooms/demo", data);
  }

  async getDoudizhuLeaderboard(limit = 10) {
    return this.get<{ entries: DoudizhuLeaderboardEntry[] }>(
      `/games/dou-dizhu/leaderboard?limit=${limit}`
    );
  }

  async getDoudizhuRecentMatches(limit = 8) {
    return this.get<{ matches: DoudizhuMatchSummary[] }>(
      `/games/dou-dizhu/matches?limit=${limit}`
    );
  }

  async getMyDoudizhuRecentMatches(limit = 8) {
    return this.get<{ matches: DoudizhuMatchSummary[] }>(
      `/games/dou-dizhu/matches/me?limit=${limit}`
    );
  }

  async getDoudizhuReplay(matchId: string) {
    return this.get<DoudizhuReplay>(`/games/dou-dizhu/matches/${matchId}/replay`);
  }

  // ── Follow ────────────────────────────────────────────────────────────

  async followUser(userId: string) {
    return this.post<void>(`/users/${userId}/follow`);
  }

  async unfollowUser(userId: string) {
    return this.delete<void>(`/users/${userId}/follow`);
  }

  async getFollowers(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{ followers: UserFollow[]; total: number }>(
      `/users/${userId}/followers?${q}`,
    );
  }

  async getFollowing(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{ following: UserFollow[]; total: number }>(
      `/users/${userId}/following?${q}`,
    );
  }

  async getFollowStats(userId: string) {
    return this.get<FollowStats>(`/users/${userId}/follow-stats`);
  }

  // ── Chat ──────────────────────────────────────────────────────────────

  async getConversations(page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{ conversations: Conversation[]; total: number }>(
      `/conversations?${q}`,
    );
  }

  async createDirectConversation(otherUserId: string) {
    return this.post<Conversation>("/conversations", {
      other_user_id: otherUserId,
    });
  }

  async getMessages(conversationId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{ messages: Message[]; total: number }>(
      `/conversations/${conversationId}/messages?${q}`,
    );
  }

  async sendMessage(
    conversationId: string,
    content: string,
    mediaUrl?: string,
  ) {
    return this.post<Message>(`/conversations/${conversationId}/messages`, {
      content,
      media_url: mediaUrl,
    });
  }

  async markRead(conversationId: string) {
    return this.put<void>(`/conversations/${conversationId}/read`);
  }

  // WebSocket connection for chat
  connectWebSocket(onMessage: (msg: any) => void): WebSocket | null {
    if (typeof window === "undefined") return null;
    const token = this.token;
    if (!token) return null;
    const ws = new WebSocket(`${WS_BASE_URL}/ws/chat?token=${token}`);
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        onMessage(msg);
      } catch {
        // ignore
      }
    };
    return ws;
  }

  // ── Tips ──────────────────────────────────────────────────────────────

  async createTip(toUserId: string, amount: number, message?: string) {
    return this.post<TipOrder>("/tips", {
      to_user_id: toUserId,
      amount,
      message,
    });
  }

  async getReceivedTips(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{ tips: TipOrder[]; total: number }>(
      `/users/${userId}/tips/received?${q}`,
    );
  }

  async payTipAlipay(orderId: string, returnUrl?: string) {
    return this.post<{ pay_url: string }>(`/orders/${orderId}/pay/alipay`, {
      return_url: returnUrl,
    });
  }

  async payTipWechat(orderId: string) {
    return this.post<{ qr_code: string }>(`/orders/${orderId}/pay/wechat`, {});
  }

  // ── Notifications ─────────────────────────────────────────────────────

  async getNotifications(page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{
      notifications: Notification[];
      total: number;
      page: number;
      size: number;
    }>(`/notifications?${q}`);
  }

  async markNotificationsRead(ids?: string[]) {
    return this.post<void>("/notifications/read", { ids: ids ?? [] });
  }

  async getUnreadCount() {
    return this.get<{ count: number }>("/notifications/unread-count");
  }

  // ── Search ────────────────────────────────────────────────────────────

  async searchAll(query: string) {
    return this.get<{
      albums: any[];
      users?: any[];
      posts?: any[];
      audio_works?: AudioWork[];
      query: string;
    }>(`/search?q=${encodeURIComponent(query)}`);
  }

  async getPopularSearches(): Promise<string[]> {
    return this.get<string[]>("/search/popular");
  }

  // ── Block / Report ────────────────────────────────────────────────────

  async getBlockedUsers() {
    return this.get<{ users: any[]; total: number }>("/users/me/blocked");
  }

  async blockUser(userId: string) {
    return this.post<{ message: string }>(`/users/${userId}/block`);
  }

  async unblockUser(userId: string) {
    return this.delete<{ message: string }>(`/users/${userId}/block`);
  }

  async createReport(
    targetType: string,
    targetId: string,
    reason: string,
    description?: string,
  ) {
    return this.post<{ message: string }>("/reports", {
      target_type: targetType,
      target_id: targetId,
      reason,
      description,
    });
  }

  async getMyReports(status?: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams();
    if (status) q.set("status", status);
    if (page) q.set("page", String(page));
    if (pageSize) q.set("page_size", String(pageSize));
    return this.get<{
      reports: any[];
      total: number;
      page: number;
      size: number;
    }>(`/reports/mine?${q.toString()}`);
  }

  async getCreatorStats() {
    return this.get<{
      post_count: number;
      total_likes: number;
      total_comments: number;
      follower_count: number;
      following_count: number;
      tip_total_cents: number;
      tip_count: number;
    }>("/creator/stats");
  }

  // ── User Profile ──────────────────────────────────────────────────────

  async getMe() {
    return this.get<any>("/users/me");
  }

  async getUser(userId: string) {
    return this.get<any>(`/users/${userId}`);
  }

  async getSponsorInfo(): Promise<{
    monthly_goal: number;
    current_raised: number;
    alipay_qr_url: string;
    wechat_qr_url: string;
    message: string;
  }> {
    return this.get("/sponsor");
  }

  async updateProfile(data: {
    bio?: string;
    website?: string;
    location?: string;
    furry_name?: string;
    species?: string;
    avatar_key?: string;
  }) {
    return this.put<any>("/users/me", data);
  }

  // ── File Upload ───────────────────────────────────────────────────────

  async uploadFile(endpoint: string, file: File): Promise<{ url: string }> {
    const formData = new FormData();
    formData.append("file", file);

    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: "POST",
      headers,
      body: formData,
    });

    const data: ApiResponse<{ url: string }> = await response.json();
    if (data.code !== 0) {
      throw new Error(data.message || "Upload failed");
    }
    return data.data;
  }

  async uploadAudio(file: File): Promise<{ url: string }> {
    return this.uploadFile("/upload/audio", file);
  }

  // ── Audio Jobs ────────────────────────────────────────────────────────

  async createAudioJob(data: {
    title: string;
    task_type: AudioJobTaskType;
    source_audio_url?: string;
    reference_audio_url?: string;
    prompt?: string;
    params?: Record<string, unknown>;
  }) {
    return this.post<AudioJob>("/audio/jobs", data);
  }

  async listAudioJobs(options?: {
    page?: number;
    page_size?: number;
    status?: AudioJobStatus;
    task_type?: AudioJobTaskType;
  }) {
    const q = new URLSearchParams();
    if (options?.page) q.set("page", String(options.page));
    if (options?.page_size) q.set("page_size", String(options.page_size));
    if (options?.status) q.set("status", options.status);
    if (options?.task_type) q.set("task_type", options.task_type);
    return this.get<{
      items: AudioJob[];
      total: number;
      page: number;
      size: number;
    }>(`/audio/jobs?${q.toString()}`);
  }

  async getAudioJob(id: string) {
    return this.get<AudioJob>(`/audio/jobs/${id}`);
  }

  async retryAudioJob(id: string) {
    return this.post<AudioJob>(`/audio/jobs/${id}/retry`);
  }

  async publishAudioJob(
    id: string,
    data?: {
      title?: string;
      description?: string;
      cover_image_url?: string;
      visibility?: AudioWorkVisibility;
      tags?: string[];
    },
  ) {
    return this.post<AudioWork>(`/audio/jobs/${id}/publish`, data ?? {});
  }

  async listMyAudioWorks(options?: {
    page?: number;
    page_size?: number;
  }) {
    const q = new URLSearchParams();
    if (options?.page) q.set("page", String(options.page));
    if (options?.page_size) q.set("page_size", String(options.page_size));
    return this.get<{
      items: AudioWork[];
      total: number;
      page: number;
      size: number;
    }>(`/users/me/audio/works?${q.toString()}`);
  }

  async listAudioWorks(options?: {
    page?: number;
    page_size?: number;
    search?: string;
    tag?: string;
    sort?: "latest" | "oldest" | "popular" | "recommended";
  }) {
    const q = new URLSearchParams();
    if (options?.page) q.set("page", String(options.page));
    if (options?.page_size) q.set("page_size", String(options.page_size));
    if (options?.search) q.set("q", options.search);
    if (options?.tag) q.set("tag", options.tag);
    if (options?.sort) q.set("sort", options.sort);
    return this.get<{
      items: AudioWork[];
      total: number;
      page: number;
      size: number;
    }>(`/audio/works?${q.toString()}`);
  }

  async listUserAudioWorks(
    userId: string,
    options?: {
      page?: number;
      page_size?: number;
    },
  ) {
    const q = new URLSearchParams();
    if (options?.page) q.set("page", String(options.page));
    if (options?.page_size) q.set("page_size", String(options.page_size));
    return this.get<{
      items: AudioWork[];
      total: number;
      page: number;
      size: number;
    }>(`/users/${userId}/audio/works?${q.toString()}`);
  }

  async getAudioWork(id: string) {
    return this.get<AudioWork>(`/audio/works/${id}`);
  }

  async getRelatedAudioWorks(id: string, limit = 6) {
    const q = new URLSearchParams();
    if (limit > 0) q.set("limit", String(limit));
    return this.get<{ items: AudioWork[] }>(`/audio/works/${id}/related?${q.toString()}`);
  }

  async recordAudioPlaybackEvent(
    id: string,
    payload: {
      event: "open" | "play" | "pause" | "seek" | "complete" | "skip_next" | "skip_previous";
      position_sec?: number;
      source_kind?: string;
    },
  ) {
    return this.post<{ recorded: boolean }>(`/audio/works/${id}/playback-events`, payload);
  }

  async updateAudioWork(
    id: string,
    data: {
      title: string;
      description?: string;
      cover_image_url?: string;
      visibility?: AudioWorkVisibility;
      tags?: string[];
    },
  ) {
    return this.put<AudioWork>(`/audio/works/${id}`, data);
  }

  async deleteAudioWork(id: string) {
    return this.delete<{ message: string }>(`/audio/works/${id}`);
  }

  async getAudioWorkMeState(id: string) {
    return this.get<{ liked: boolean; bookmarked: boolean }>(`/audio/works/${id}/me-state`);
  }

  async likeAudioWork(id: string) {
    return this.post<{ message: string }>(`/audio/works/${id}/like`);
  }

  async unlikeAudioWork(id: string) {
    return this.delete<{ message: string }>(`/audio/works/${id}/like`);
  }

  // ── Events ───────────────────────────────────────────────────────────────

  async listEvents(page = 1, pageSize = 20) {
    return this.request<{
      events: Event[];
      total: number;
      page: number;
      page_size: number;
    }>(`/events?page=${page}&page_size=${pageSize}`);
  }

  async getEvent(id: string) {
    return this.request<Event>(`/events/${id}`);
  }

  async createEvent(data: {
    title: string;
    description?: string;
    location?: string;
    is_online?: boolean;
    start_time: string;
    end_time: string;
    max_capacity?: number;
    tags?: string[];
  }) {
    return this.request<Event>("/events", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async attendEvent(eventId: string, status = "attending") {
    return this.request<void>(`/events/${eventId}/attend`, {
      method: "POST",
      body: JSON.stringify({ status }),
    });
  }

  async listEventAttendees(eventId: string, page = 1, pageSize = 20) {
    return this.request<{ attendees: EventAttendee[]; total: number }>(
      `/events/${eventId}/attendees?page=${page}&page_size=${pageSize}`,
    );
  }

  async myEvents(page = 1, pageSize = 20) {
    return this.request<{ events: Event[]; total: number }>(
      `/users/me/events?page=${page}&page_size=${pageSize}`,
    );
  }

  async myAttending(page = 1, pageSize = 20) {
    return this.request<{ events: Event[]; total: number }>(
      `/users/me/attending?page=${page}&page_size=${pageSize}`,
    );
  }

  // ── Groups ───────────────────────────────────────────────────────────────

  async listGroups(params?: {
    search?: string;
    privacy?: string;
    page?: number;
    page_size?: number;
  }) {
    const q = new URLSearchParams();
    if (params?.search) q.set("search", params.search);
    if (params?.privacy) q.set("privacy", params.privacy);
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    return this.request<{
      groups: Group[];
      total: number;
      page: number;
      page_size: number;
    }>(`/groups?${q.toString()}`);
  }

  async getGroup(id: string) {
    return this.request<Group>(`/groups/${id}`);
  }

  async createGroup(data: {
    name: string;
    description?: string;
    announcement?: string;
    rules?: string;
    tags?: string[];
    privacy?: "public" | "private";
  }) {
    return this.request<Group>("/groups", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async joinGroup(id: string) {
    return this.request<void>(`/groups/${id}/join`, { method: "POST" });
  }

  async updateGroup(
    id: string,
    data: {
      name?: string;
      description?: string;
      announcement?: string;
      rules?: string;
      tags?: string[];
      privacy?: "public" | "private";
    },
  ) {
    return this.request<Group>(`/groups/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async leaveGroup(id: string) {
    return this.request<void>(`/groups/${id}/leave`, { method: "DELETE" });
  }

  async listGroupMembers(id: string, page = 1, pageSize = 20) {
    return this.request<{ members: GroupMember[]; total: number }>(
      `/groups/${id}/members?page=${page}&page_size=${pageSize}`,
    );
  }

  async myGroups(page = 1, pageSize = 20) {
    return this.request<{ groups: Group[]; total: number }>(
      `/users/me/groups?page=${page}&page_size=${pageSize}`,
    );
  }

  async getMyGroupsDashboard() {
    return this.request<GroupDashboard>("/users/me/groups/dashboard");
  }

  async bookmarkPost(id: string) {
    return this.post<{ message: string }>(`/posts/${id}/bookmark`);
  }

  async unbookmarkPost(id: string) {
    return this.delete<{ message: string }>(`/posts/${id}/bookmark`);
  }

  async bookmarkGroup(id: string) {
    return this.post<{ message: string }>(`/groups/${id}/bookmark`);
  }

  async unbookmarkGroup(id: string) {
    return this.delete<{ message: string }>(`/groups/${id}/bookmark`);
  }

  async bookmarkEvent(id: string) {
    return this.post<{ message: string }>(`/events/${id}/bookmark`);
  }

  async unbookmarkEvent(id: string) {
    return this.delete<{ message: string }>(`/events/${id}/bookmark`);
  }

  async bookmarkAudioWork(id: string) {
    return this.post<{ message: string }>(`/audio/works/${id}/bookmark`);
  }

  async unbookmarkAudioWork(id: string) {
    return this.delete<{ message: string }>(`/audio/works/${id}/bookmark`);
  }

  async checkBookmark(
    targetType: "post" | "group" | "event" | "audio_work",
    targetId: string,
  ) {
    return this.get<{ bookmarked: boolean }>(
      `/bookmarks/check?target_type=${targetType}&target_id=${targetId}`,
    );
  }

  async getBookmarkedPosts(page = 1, pageSize = 20) {
    return this.getBookmarkedPostsWithSort(page, pageSize, "latest");
  }

  async getBookmarkedPostsWithSort(
    page = 1,
    pageSize = 20,
    sort: "latest" | "oldest" = "latest",
  ) {
    return this.get<{
      posts: Post[];
      total: number;
      page: number;
      size: number;
    }>(`/bookmarks/posts?page=${page}&page_size=${pageSize}&sort=${sort}`);
  }

  async getBookmarkedGroups(page = 1, pageSize = 20) {
    return this.getBookmarkedGroupsWithSort(page, pageSize, "latest");
  }

  async getBookmarkedGroupsWithSort(
    page = 1,
    pageSize = 20,
    sort: "latest" | "oldest" = "latest",
  ) {
    return this.get<{
      groups: Group[];
      total: number;
      page: number;
      size: number;
    }>(`/bookmarks/groups?page=${page}&page_size=${pageSize}&sort=${sort}`);
  }

  async getBookmarkedEvents(page = 1, pageSize = 20) {
    return this.getBookmarkedEventsWithSort(page, pageSize, "latest");
  }

  async getBookmarkedEventsWithSort(
    page = 1,
    pageSize = 20,
    sort: "latest" | "oldest" = "latest",
  ) {
    return this.get<{
      events: Event[];
      total: number;
      page: number;
      size: number;
    }>(`/bookmarks/events?page=${page}&page_size=${pageSize}&sort=${sort}`);
  }

  async getBookmarkedAudioWorks(page = 1, pageSize = 20) {
    return this.getBookmarkedAudioWorksWithSort(page, pageSize, "latest");
  }

  async getBookmarkedAudioWorksWithSort(
    page = 1,
    pageSize = 20,
    sort: "latest" | "oldest" = "latest",
  ) {
    return this.get<{
      works: AudioWork[];
      total: number;
      page: number;
      size: number;
    }>(`/bookmarks/audio/works?page=${page}&page_size=${pageSize}&sort=${sort}`);
  }

  async batchDeleteBookmarks(
    targetType: "post" | "group" | "event" | "audio_work",
    targetIds: string[],
  ) {
    return this.post<{ message: string }>("/bookmarks/batch-delete", {
      target_type: targetType,
      target_ids: targetIds,
    });
  }

  // ── Leaderboard ──────────────────────────────────────────────────────────

  async getLeaderboard(limit = 20) {
    return this.get<LeaderboardEntry[]>(`/leaderboard?limit=${limit}`);
  }

  async getWeeklyLeaderboard(limit = 20) {
    return this.get<LeaderboardEntry[]>(`/leaderboard/weekly?limit=${limit}`);
  }

  async streamAssistantChat(
    messages: AssistantChatMessage[],
    handlers: AssistantStreamHandlers = {},
    conversationId?: string,
    pageContext?: AssistantPageContextPayload,
  ) {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    };
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}/assistant/chat/stream`, {
      method: "POST",
      headers,
      body: JSON.stringify({
        conversation_id: conversationId,
        messages: messages.map(({ role, content }) => ({ role, content })),
        page_context: pageContext,
      }),
      signal: handlers.signal,
    });

    if (!response.ok || !response.body) {
      let message = "AI 助手暂时不可用";
      try {
        const text = await response.text();
        if (text) {
          const maybeJSON = JSON.parse(text);
          message = maybeJSON?.message || maybeJSON?.error || message;
        }
      } catch {
        // ignore
      }
      throw new Error(message);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true }).replace(/\r\n/g, "\n");
      let boundary = buffer.indexOf("\n\n");
      while (boundary !== -1) {
        const rawBlock = buffer.slice(0, boundary);
        buffer = buffer.slice(boundary + 2);
        boundary = buffer.indexOf("\n\n");

        const parsed = parseSSEBlock(rawBlock);
        if (!parsed) continue;

        try {
          const payload = JSON.parse(parsed.data);
          switch (parsed.event) {
            case "meta":
              handlers.onMeta?.(payload as AssistantMeta);
              break;
            case "token":
              handlers.onToken?.((payload?.content as string) || "");
              break;
            case "done":
              handlers.onDone?.();
              break;
            case "error": {
              const message =
                (payload?.message as string) || "AI 助手暂时不可用";
              handlers.onError?.(message);
              throw new Error(message);
            }
            default:
              break;
          }
        } catch (err) {
          if (err instanceof Error) {
            throw err;
          }
          throw new Error("AI 助手响应解析失败");
        }
      }
    }
  }

  async getAssistantConversations(page = 1, pageSize = 20) {
    return this.get<{
      conversations: AssistantConversation[];
      total: number;
      page: number;
      size: number;
    }>(`/assistant/conversations?page=${page}&page_size=${pageSize}`);
  }

  async getAssistantConversation(id: string, page = 1, pageSize = 100) {
    return this.get<{
      conversation: AssistantConversation;
      messages: AssistantChatMessage[];
      total: number;
      page: number;
      size: number;
    }>(`/assistant/conversations/${id}?page=${page}&page_size=${pageSize}`);
  }

  async getAssistantSettings() {
    return this.get<AssistantSettings>("/admin/assistant/settings");
  }

  async updateAssistantSettings(data: AssistantSettings) {
    return this.put<AssistantSettings>("/admin/assistant/settings", data);
  }

  async submitAssistantFeedback(data: AssistantFeedbackInput) {
    return this.post<{ ok: boolean }>("/assistant/feedback", data);
  }

  async getAssistantOverview() {
    return this.get<AssistantOverviewData>("/admin/assistant/overview");
  }

  async analyzeAssistantMedia(mediaUrls: string[], purpose = "post_create") {
    return this.post<MediaAnalysisResult>("/assistant/media/analyze", {
      media_urls: mediaUrls,
      purpose,
    });
  }

  async getAdminOrders(params?: {
    page?: number;
    page_size?: number;
    status?: string;
    payment_method?: string;
    type?: string;
    search?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.status) q.set("status", params.status);
    if (params?.payment_method) q.set("payment_method", params.payment_method);
    if (params?.type) q.set("type", params.type);
    if (params?.search) q.set("search", params.search);
    return this.get<{
      orders: AdminOrder[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/orders?${q.toString()}`);
  }

  async updateAdminOrderStatus(id: string, status: string) {
    return this.put<{ status: string }>(`/admin/orders/${id}/status`, { status });
  }

  async getAdminGroups(params?: {
    page?: number;
    page_size?: number;
    privacy?: string;
    search?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.privacy) q.set("privacy", params.privacy);
    if (params?.search) q.set("search", params.search);
    return this.get<{
      groups: AdminGroup[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/groups?${q.toString()}`);
  }

  async getAdminUsers(params?: {
    page?: number;
    page_size?: number;
    search?: string;
    status?: string;
    role?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.search) q.set("search", params.search);
    if (params?.status) q.set("status", params.status);
    if (params?.role) q.set("role", params.role);
    return this.get<{
      users: AdminUser[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/users?${q.toString()}`);
  }

  async getAdminReports(params?: {
    page?: number;
    page_size?: number;
    status?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.status) q.set("status", params.status);
    return this.get<{
      reports: AdminReport[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/reports?${q.toString()}`);
  }

  async getAdminAudioWorks(params?: {
    page?: number;
    page_size?: number;
    status?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.status) q.set("status", params.status);
    return this.get<{
      works: AdminAudioWork[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/audio/works?${q.toString()}`);
  }

  async updateAdminAudioWorkModeration(id: string, status: ModerationStatus, note?: string) {
    return this.put<{ status: ModerationStatus; note?: string }>(`/admin/audio/works/${id}/moderation`, {
      status,
      note,
    });
  }

  async updateAdminGroup(id: string, data: { privacy: "public" | "private" }) {
    return this.put<Group>(`/admin/groups/${id}`, data);
  }

  async getAdminEvents(params?: {
    page?: number;
    page_size?: number;
    status?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.status) q.set("status", params.status);
    return this.get<{
      events: AdminEvent[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/events?${q.toString()}`);
  }

  async updateAdminEventStatus(id: string, status: Event["status"]) {
    return this.put<Event>(`/admin/events/${id}/status`, { status });
  }

  async generateAdminReportSummary(reportId: string) {
    return this.post<AdminAIToolResult>("/admin/assistant/tools/report-summary", {
      report_id: reportId,
    });
  }

  async generateAdminWeeklyReport(days = 7) {
    return this.post<AdminAIToolResult>("/admin/assistant/tools/weekly-report", {
      days,
    });
  }

  async generateAdminCreatorRecommendation(userId: string) {
    return this.post<AdminAIToolResult>("/admin/assistant/tools/creator-recommendation", {
      user_id: userId,
    });
  }

  async generateAdminEventCopy(eventId: string) {
    return this.post<AdminAIToolResult>("/admin/assistant/tools/event-copy", {
      event_id: eventId,
    });
  }

  async generateAdminModerationExplanation(postId: string) {
    return this.post<AdminAIToolResult>("/admin/assistant/tools/moderation-explanation", {
      post_id: postId,
    });
  }

  async getAdminAuditLogs(params?: {
    page?: number;
    page_size?: number;
    action?: string;
    resource?: string;
    user_id?: string;
    resource_id?: string;
    start_time?: string;
    end_time?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.page) q.set("page", String(params.page));
    if (params?.page_size) q.set("page_size", String(params.page_size));
    if (params?.action) q.set("action", params.action);
    if (params?.resource) q.set("resource", params.resource);
    if (params?.user_id) q.set("user_id", params.user_id);
    if (params?.resource_id) q.set("resource_id", params.resource_id);
    if (params?.start_time) q.set("start_time", params.start_time);
    if (params?.end_time) q.set("end_time", params.end_time);
    return this.get<{
      logs: AuditLog[];
      total: number;
      page: number;
      size: number;
    }>(`/admin/audit-logs?${q.toString()}`);
  }

  async getAdminPermissionMatrix() {
    return this.get<{
      roles: PermissionMatrixRole[];
      catalog: Record<string, string[]>;
    }>("/admin/permissions/matrix");
  }

  async exportAdminAuditLogsCsv(params?: {
    action?: string;
    resource?: string;
    user_id?: string;
    resource_id?: string;
    start_time?: string;
    end_time?: string;
  }) {
    const q = new URLSearchParams();
    if (params?.action) q.set("action", params.action);
    if (params?.resource) q.set("resource", params.resource);
    if (params?.user_id) q.set("user_id", params.user_id);
    if (params?.resource_id) q.set("resource_id", params.resource_id);
    if (params?.start_time) q.set("start_time", params.start_time);
    if (params?.end_time) q.set("end_time", params.end_time);

    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}/admin/audit-logs/export?${q.toString()}`, {
      method: "GET",
      headers,
    });
    if (!response.ok) {
      throw new Error("导出审计日志失败");
    }
    return response.blob();
  }

  async getAdminSystemConfig() {
    return this.get<AdminSystemConfig>("/admin/system/config");
  }

  async updateAdminSponsorConfig(data: AdminSponsorConfig) {
    return this.put<AdminSponsorConfig>("/admin/system/sponsor", data);
  }

  async getAdminGamesOverview() {
    return this.get<AdminGameOverview>("/admin/games/overview");
  }
}

export const apiClient = new ApiClient(API_BASE_URL);

// ── Domain types ───────────────────────────────────────────────────────────

export interface Event {
  id: string;
  organizer_id: string;
  title: string;
  description: string;
  location: string;
  is_online: boolean;
  start_time: string;
  end_time: string;
  max_capacity: number;
  tags: string[];
  status: "draft" | "published" | "cancelled" | "completed";
  attendee_count: number;
  created_at: string;
  updated_at: string;
}

export interface AdminEvent extends Event {
  organizer_username?: string;
  organizer_email?: string;
}

export interface AdminAudioWork extends AudioWork {
  author_username?: string;
}

export interface AdminUser {
  id: string;
  username: string;
  email: string;
  nickname?: string;
  furry_name?: string;
  role: string;
  status: string;
  created_at: string;
}

export interface EventAttendee {
  event_id: string;
  user_id: string;
  status: "attending" | "maybe" | "not_going";
  joined_at: string;
}

export interface Group {
  id: string;
  owner_id: string;
  name: string;
  description: string;
  announcement: string;
  rules: string;
  avatar_key?: string;
  featured_post_id?: string;
  tags: string[];
  privacy: "public" | "private";
  member_count: number;
  post_count: number;
  created_at: string;
  updated_at: string;
}

export interface AdminGroup extends Group {
  owner_username?: string;
  owner_email?: string;
}

export interface GroupMember {
  group_id: string;
  user_id: string;
  role: "owner" | "moderator" | "member";
  joined_at: string;
  username?: string;
  furry_name?: string;
  avatar_key?: string;
}

export interface AdminOrder {
  id: string;
  order_no: string;
  user_id: string;
  status: string;
  total_cents: number;
  currency: string;
  discount_cents: number;
  payment_method?: string;
  metadata?: Record<string, any>;
  paid_at?: string | null;
  created_at: string;
  expires_at?: string | null;
  updated_at: string;
  order_type?: string;
  recipient_user_id?: string;
  payer_username?: string;
  payer_email?: string;
  recipient_username?: string;
  recipient_email?: string;
}

export interface AdminReport {
  id: string;
  reporter_id: string;
  reporter_username?: string;
  target_type: "post" | "comment" | "user" | "audio_work";
  target_id: string;
  reason: string;
  description?: string;
  status: "pending" | "reviewed" | "dismissed";
  action_taken?: string | null;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  created_at: string;
}

export interface AuditLog {
  id: string;
  user_id?: string | null;
  username: string;
  action: string;
  resource: string;
  resource_id?: string | null;
  ip_address: string;
  user_agent?: string | null;
  before_data?: string | null;
  after_data?: string | null;
  error_message?: string | null;
  created_at: string;
}

export interface PermissionMatrixRole {
  role: string;
  permissions: string[];
  count: number;
}

export interface AdminSponsorConfig {
  monthly_goal: number;
  current_raised: number;
  alipay_qr_url: string;
  wechat_qr_url: string;
  message: string;
}

export interface AdminSystemConfig {
  server: {
    mode: string;
    port: number;
    frontend_url: string;
    allow_origins: string[];
  };
  ratelimit: {
    unauthenticated: number;
    authenticated: number;
    admin: number;
  };
  sponsor: AdminSponsorConfig;
  assistant: {
    provider: string;
    base_url: string;
    model: string;
    embedding_base_url?: string;
    embedding_model?: string;
    embedding_dims?: number;
    vision_base_url?: string;
    vision_model?: string;
    timeout_sec: number;
    vision_timeout_sec?: number;
    max_context_items: number;
    persona_name: string;
    configured: boolean;
    embedding_configured?: boolean;
    vision_configured?: boolean;
    retrieval_limit?: number;
    vector_scan_limit?: number;
    sync_interval_sec?: number;
  };
  oss: {
    provider: string;
    bucket: string;
    endpoint: string;
    region: string;
    allowed_hosts: string[];
  };
  email: {
    configured: boolean;
    host: string;
    from: string;
  };
  payment: {
    alipay_configured: boolean;
    wechat_configured: boolean;
  };
  grpc: {
    stats_addr: string;
    notification_addr: string;
    moderation_addr: string;
    stats_port: number;
    notification_port: number;
    moderation_port: number;
  };
  audio_metrics: {
    playback_events_total: number;
    events_by_type: Record<string, number>;
    last_event?: string;
    last_source_kind?: string;
    last_position_sec: number;
  };
}

export interface GroupAnnouncement {
  id: string;
  group_id: string;
  author_id: string;
  content: string;
  created_at: string;
  author_name?: string;
  furry_name?: string;
  avatar_key?: string;
}

export interface GroupDashboardItem {
  group: Group;
  role: "owner" | "moderator" | "member";
  active_members: GroupMember[];
}

export interface GroupDashboardStats {
  created_count: number;
  managed_count: number;
  total_members: number;
  total_posts: number;
  featured_group_count: number;
}

export interface GroupDashboard {
  stats: GroupDashboardStats;
  created_groups: GroupDashboardItem[];
  managed_groups: GroupDashboardItem[];
}

export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  score: number;
}
