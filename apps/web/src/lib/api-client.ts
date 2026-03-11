const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
  timestamp?: number
}

export type ModerationStatus = 'pending' | 'approved' | 'blocked'

export interface Post {
  id: string
  author_id: string
  title?: string
  content: string
  media_urls?: string[]
  tags?: string[]
  content_labels?: Record<string, boolean>
  visibility: 'public' | 'followers_only' | 'private'
  moderation_status: ModerationStatus
  like_count: number
  comment_count: number
  is_pinned: boolean
  created_at: string
  updated_at: string
  author_username?: string
  author_avatar_key?: string
  is_liked_by_me?: boolean
}

export interface OSSUploadPolicy {
  host: string
  OSSAccessKeyId: string
  policy: string
  signature: string
  expire: number
  dir: string
}

export interface UserFollow {
  follower_id: string
  followee_id: string
  created_at: string
}

export interface FollowStats {
  user_id: string
  follower_count: number
  following_count: number
}

export interface Conversation {
  id: string
  type: 'direct' | 'group'
  name?: string
  members: string[]
  created_at: string
  updated_at: string
  last_message?: Message
  unread_count?: number
}

export interface Message {
  id: string
  conversation_id: string
  sender_id: string
  content: string
  media_url?: string
  is_read: boolean
  created_at: string
  sender_username?: string
  sender_avatar_key?: string
}

export interface TipOrder {
  id: string
  order_no: string
  user_id: string
  status: string
  total_cents: number
  currency: string
  metadata?: {
    type: string
    to_user_id: string
    message: string
  }
  created_at: string
}

export interface Comment {
  id: string
  user_id: string
  commentable_type: string
  commentable_id: string
  parent_id?: string
  content: string
  is_edited: boolean
  like_count: number
  reply_count: number
  created_at: string
  updated_at: string
  author_username?: string
  author_avatar_key?: string
}

export interface Notification {
  id: string
  user_id: string
  actor_id?: string
  type: 'like' | 'comment' | 'follow' | 'tip' | 'system'
  target_id?: string
  target_type?: string
  is_read: boolean
  created_at: string
  actor_username?: string
  actor_avatar_key?: string
}

class ApiClient {
  private baseUrl: string
  private token: string | null = null
  private refreshing: Promise<void> | null = null

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('access_token')
    }
  }

  setToken(token: string | null) {
    this.token = token
    if (typeof window !== 'undefined') {
      if (token) {
        localStorage.setItem('access_token', token)
        document.cookie = `_auth=1; path=/; max-age=${7 * 24 * 3600}; SameSite=Lax`
      } else {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        document.cookie = '_auth=; path=/; max-age=0'
      }
    }
  }

  setRefreshToken(token: string) {
    if (typeof window !== 'undefined') {
      localStorage.setItem('refresh_token', token)
    }
  }

  getToken(): string | null {
    return this.token
  }

  private async tryRefresh(): Promise<boolean> {
    const refreshToken = typeof window !== 'undefined'
      ? localStorage.getItem('refresh_token')
      : null
    if (!refreshToken) return false

    try {
      const res = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      const data: ApiResponse<{ access_token: string; refresh_token: string }> = await res.json()
      if (data.code !== 0) return false
      this.setToken(data.data.access_token)
      this.setRefreshToken(data.data.refresh_token)
      return true
    } catch {
      return false
    }
  }

  private async request<T = any>(
    endpoint: string,
    options: RequestInit = {},
    isRetry = false
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    })

    if (response.status === 401 && !isRetry) {
      if (!this.refreshing) {
        this.refreshing = this.tryRefresh().then(ok => {
          this.refreshing = null
          if (!ok) {
            this.setToken(null)
            if (typeof window !== 'undefined') {
              window.location.href = '/login'
            }
          }
        })
      }
      await this.refreshing
      if (this.token) {
        return this.request<T>(endpoint, options, true)
      }
      throw new Error('登录已过期，请重新登录')
    }

    const data: ApiResponse<T> = await response.json()

    if (data.code !== 0) {
      throw new Error(data.message || 'Request failed')
    }

    return data.data
  }

  async get<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  async post<T = any>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async put<T = any>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async delete<T = any>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }

  // ── Auth ──────────────────────────────────────────────────────────────

  async login(email: string, password: string) {
    return this.post<{ access_token: string; refresh_token: string; user: any }>('/auth/login', {
      email,
      password,
    })
  }

  async register(username: string, email: string, password: string) {
    return this.post<{ access_token: string; refresh_token: string; user: any }>('/auth/register', {
      username,
      email,
      password,
    })
  }

  async logout() {
    return this.post<void>('/auth/logout')
  }

  // ── Posts ─────────────────────────────────────────────────────────────

  async getFeed(page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ posts: Post[]; total: number; page: number; size: number }>(`/feed?${q}`)
  }

  async getExplore(page?: number, pageSize?: number, tag?: string) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    if (tag) q.set('tag', tag)
    return this.get<{ posts: Post[]; total: number; page: number; size: number }>(`/explore?${q}`)
  }

  async getHotTags(): Promise<string[]> {
    return this.get<string[]>('/explore/tags')
  }

  async getPost(id: string) {
    return this.get<Post>(`/posts/${id}`)
  }

  async createPost(data: { title?: string; content: string; media_urls?: string[]; tags?: string[]; visibility?: string; is_ai_generated?: boolean }) {
    return this.post<Post>('/posts', data)
  }

  async getOSSPolicy(purpose?: string): Promise<OSSUploadPolicy> {
    return this.post<OSSUploadPolicy>('/upload/oss-policy', { purpose: purpose ?? 'post' })
  }

  async updatePost(id: string, data: { title?: string; content: string; media_urls?: string[]; tags?: string[]; visibility?: string }) {
    return this.put<Post>(`/posts/${id}`, data)
  }

  async deletePost(id: string) {
    return this.delete<void>(`/posts/${id}`)
  }

  async likePost(id: string) {
    return this.post<void>(`/posts/${id}/like`)
  }

  async unlikePost(id: string) {
    return this.delete<void>(`/posts/${id}/like`)
  }

  async getComments(postId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams({ commentable_type: 'post', commentable_id: postId })
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ comments: Comment[]; total: number; page: number; size: number }>(`/comments?${q}`)
  }

  async createComment(postId: string, content: string) {
    return this.post<Comment>('/comments', {
      commentable_type: 'post',
      commentable_id: postId,
      content,
    })
  }

  async getUserPosts(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ posts: Post[]; total: number; page: number; size: number }>(`/users/${userId}/posts?${q}`)
  }

  // ── Follow ────────────────────────────────────────────────────────────

  async followUser(userId: string) {
    return this.post<void>(`/users/${userId}/follow`)
  }

  async unfollowUser(userId: string) {
    return this.delete<void>(`/users/${userId}/follow`)
  }

  async getFollowers(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ followers: UserFollow[]; total: number }>(`/users/${userId}/followers?${q}`)
  }

  async getFollowing(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ following: UserFollow[]; total: number }>(`/users/${userId}/following?${q}`)
  }

  async getFollowStats(userId: string) {
    return this.get<FollowStats>(`/users/${userId}/follow-stats`)
  }

  // ── Chat ──────────────────────────────────────────────────────────────

  async getConversations(page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ conversations: Conversation[]; total: number }>(`/conversations?${q}`)
  }

  async createDirectConversation(otherUserId: string) {
    return this.post<Conversation>('/conversations', { other_user_id: otherUserId })
  }

  async getMessages(conversationId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ messages: Message[]; total: number }>(`/conversations/${conversationId}/messages?${q}`)
  }

  async sendMessage(conversationId: string, content: string, mediaUrl?: string) {
    return this.post<Message>(`/conversations/${conversationId}/messages`, { content, media_url: mediaUrl })
  }

  async markRead(conversationId: string) {
    return this.put<void>(`/conversations/${conversationId}/read`)
  }

  // WebSocket connection for chat
  connectWebSocket(onMessage: (msg: any) => void): WebSocket | null {
    if (typeof window === 'undefined') return null
    const token = this.token
    if (!token) return null
    const ws = new WebSocket(`${WS_BASE_URL}/ws/chat?token=${token}`)
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        onMessage(msg)
      } catch {
        // ignore
      }
    }
    return ws
  }

  // ── Tips ──────────────────────────────────────────────────────────────

  async createTip(toUserId: string, amount: number, message?: string) {
    return this.post<TipOrder>('/tips', { to_user_id: toUserId, amount, message })
  }

  async getReceivedTips(userId: string, page?: number, pageSize?: number) {    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ tips: TipOrder[]; total: number }>(`/users/${userId}/tips/received?${q}`)
  }

  async payTipAlipay(orderId: string, returnUrl?: string) {
    return this.post<{ pay_url: string }>(`/orders/${orderId}/pay/alipay`, { return_url: returnUrl })
  }

  async payTipWechat(orderId: string) {
    return this.post<{ qr_code: string }>(`/orders/${orderId}/pay/wechat`, {})
  }

  // ── Notifications ─────────────────────────────────────────────────────

  async getNotifications(page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ notifications: Notification[]; total: number; page: number; size: number }>(`/notifications?${q}`)
  }

  async markNotificationsRead(ids?: string[]) {
    return this.post<void>('/notifications/read', { ids: ids ?? [] })
  }

  async getUnreadCount() {
    return this.get<{ count: number }>('/notifications/unread-count')
  }

  // ── Search ────────────────────────────────────────────────────────────

  async searchAll(query: string) {
    return this.get<{ albums: any[]; games?: any[]; users?: any[]; posts?: any[]; query: string }>(`/search?q=${encodeURIComponent(query)}`)
  }

  async getPopularSearches(): Promise<string[]> {
    return this.get<string[]>('/search/popular')
  }

  // ── Block / Report ────────────────────────────────────────────────────

  async getBlockedUsers() {
    return this.get<{ users: any[]; total: number }>('/users/me/blocked')
  }

  async blockUser(userId: string) {
    return this.post<{ message: string }>(`/users/${userId}/block`)
  }

  async unblockUser(userId: string) {
    return this.delete<{ message: string }>(`/users/${userId}/block`)
  }

  async createReport(targetType: string, targetId: string, reason: string, description?: string) {
    return this.post<{ message: string }>('/reports', { target_type: targetType, target_id: targetId, reason, description })
  }

  async getCreatorStats() {
    return this.get<{
      post_count: number
      total_likes: number
      total_comments: number
      follower_count: number
      following_count: number
      tip_total_cents: number
      tip_count: number
    }>('/creator/stats')
  }

  // ── User Profile ──────────────────────────────────────────────────────

  async getMe() {
    return this.get<any>('/users/me')
  }

  async getUser(userId: string) {
    return this.get<any>(`/users/${userId}`)
  }

  async getSponsorInfo(): Promise<{
    monthly_goal: number
    current_raised: number
    alipay_qr_url: string
    wechat_qr_url: string
    message: string
  }> {
    return this.get('/sponsor')
  }

  async updateProfile(data: {
    bio?: string
    website?: string
    location?: string
    furry_name?: string
    species?: string
    avatar_key?: string
  }) {
    return this.put<any>('/users/me', data)
  }

  // ── File Upload ───────────────────────────────────────────────────────

  async uploadFile(endpoint: string, file: File): Promise<{ url: string }> {
    const formData = new FormData()
    formData.append('file', file)

    const headers: Record<string, string> = {}
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers,
      body: formData,
    })

    const data: ApiResponse<{ url: string }> = await response.json()
    if (data.code !== 0) {
      throw new Error(data.message || 'Upload failed')
    }
    return data.data
  }

  // ── Events ───────────────────────────────────────────────────────────────

  async listEvents(page = 1, pageSize = 20) {
    return this.request<{ events: Event[]; total: number; page: number; page_size: number }>(
      `/events?page=${page}&page_size=${pageSize}`
    )
  }

  async getEvent(id: string) {
    return this.request<Event>(`/events/${id}`)
  }

  async createEvent(data: {
    title: string
    description?: string
    location?: string
    is_online?: boolean
    start_time: string
    end_time: string
    max_capacity?: number
    tags?: string[]
  }) {
    return this.request<Event>('/events', { method: 'POST', body: JSON.stringify(data) })
  }

  async attendEvent(eventId: string, status = 'attending') {
    return this.request<void>(`/events/${eventId}/attend`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    })
  }

  async listEventAttendees(eventId: string, page = 1, pageSize = 20) {
    return this.request<{ attendees: EventAttendee[]; total: number }>(
      `/events/${eventId}/attendees?page=${page}&page_size=${pageSize}`
    )
  }

  async myEvents(page = 1, pageSize = 20) {
    return this.request<{ events: Event[]; total: number }>(
      `/users/me/events?page=${page}&page_size=${pageSize}`
    )
  }

  async myAttending(page = 1, pageSize = 20) {
    return this.request<{ events: Event[]; total: number }>(
      `/users/me/attending?page=${page}&page_size=${pageSize}`
    )
  }

  // ── Groups ───────────────────────────────────────────────────────────────

  async listGroups(params?: { search?: string; privacy?: string; page?: number; page_size?: number }) {
    const q = new URLSearchParams()
    if (params?.search) q.set('search', params.search)
    if (params?.privacy) q.set('privacy', params.privacy)
    if (params?.page) q.set('page', String(params.page))
    if (params?.page_size) q.set('page_size', String(params.page_size))
    return this.request<{ groups: Group[]; total: number; page: number; page_size: number }>(
      `/groups?${q.toString()}`
    )
  }

  async getGroup(id: string) {
    return this.request<Group>(`/groups/${id}`)
  }

  async createGroup(data: { name: string; description?: string; tags?: string[]; privacy?: 'public' | 'private' }) {
    return this.request<Group>('/groups', { method: 'POST', body: JSON.stringify(data) })
  }

  async joinGroup(id: string) {
    return this.request<void>(`/groups/${id}/join`, { method: 'POST' })
  }

  async leaveGroup(id: string) {
    return this.request<void>(`/groups/${id}/leave`, { method: 'DELETE' })
  }

  async listGroupMembers(id: string, page = 1, pageSize = 20) {
    return this.request<{ members: GroupMember[]; total: number }>(
      `/groups/${id}/members?page=${page}&page_size=${pageSize}`
    )
  }

  async myGroups(page = 1, pageSize = 20) {
    return this.request<{ groups: Group[]; total: number }>(
      `/users/me/groups?page=${page}&page_size=${pageSize}`
    )
  }

  // ── Leaderboard ──────────────────────────────────────────────────────────

  async getLeaderboard(limit = 20) {
    return this.get<LeaderboardEntry[]>(`/leaderboard?limit=${limit}`)
  }

  async getWeeklyLeaderboard(limit = 20) {
    return this.get<LeaderboardEntry[]>(`/leaderboard/weekly?limit=${limit}`)
  }
}

export const apiClient = new ApiClient(API_BASE_URL)

// ── Domain types ───────────────────────────────────────────────────────────

export interface Event {
  id: string
  organizer_id: string
  title: string
  description: string
  location: string
  is_online: boolean
  start_time: string
  end_time: string
  max_capacity: number
  tags: string[]
  status: 'draft' | 'published' | 'cancelled' | 'completed'
  attendee_count: number
  created_at: string
  updated_at: string
}

export interface EventAttendee {
  event_id: string
  user_id: string
  status: 'attending' | 'maybe' | 'not_going'
  joined_at: string
}

export interface Group {
  id: string
  owner_id: string
  name: string
  description: string
  avatar_key?: string
  tags: string[]
  privacy: 'public' | 'private'
  member_count: number
  post_count: number
  created_at: string
  updated_at: string
}

export interface GroupMember {
  group_id: string
  user_id: string
  role: 'owner' | 'moderator' | 'member'
  joined_at: string
}

export interface LeaderboardEntry {
  rank: number
  user_id: string
  username: string
  score: number
}

