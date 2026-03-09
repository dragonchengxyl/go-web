const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
  timestamp?: number
}

export interface Achievement {
  id: string
  name: string
  description: string
  icon: string
  points: number
  unlocked_at?: string
}

export interface LeaderboardEntry {
  rank: number
  user_id: string
  username: string
  avatar?: string
  points: number
  level: number
}

export interface Post {
  id: string
  author_id: string
  title?: string
  content: string
  media_urls?: string[]
  tags?: string[]
  visibility: 'public' | 'followers_only' | 'private'
  like_count: number
  comment_count: number
  is_pinned: boolean
  created_at: string
  updated_at: string
  author_username?: string
  author_avatar_key?: string
  is_liked_by_me?: boolean
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

class ApiClient {
  private baseUrl: string
  private token: string | null = null

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
      } else {
        localStorage.removeItem('access_token')
      }
    }
  }

  getToken(): string | null {
    return this.token
  }

  private async request<T = any>(
    endpoint: string,
    options: RequestInit = {}
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

  async getExplore(page?: number, pageSize?: number) {
    const q = new URLSearchParams()
    if (page) q.set('page', String(page))
    if (pageSize) q.set('page_size', String(pageSize))
    return this.get<{ posts: Post[]; total: number; page: number; size: number }>(`/explore?${q}`)
  }

  async getPost(id: string) {
    return this.get<Post>(`/posts/${id}`)
  }

  async createPost(data: { title?: string; content: string; media_urls?: string[]; tags?: string[]; visibility?: string }) {
    return this.post<Post>('/posts', data)
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

  async getReceivedTips(userId: string, page?: number, pageSize?: number) {
    const q = new URLSearchParams()
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

  // ── Music ─────────────────────────────────────────────────────────────

  async getAlbums(params?: { page?: number; page_size?: number; search?: string }) {
    const q = new URLSearchParams()
    if (params?.page) q.set('page', String(params.page))
    if (params?.page_size) q.set('page_size', String(params.page_size))
    if (params?.search) q.set('search', params.search)
    return this.get<{ albums: any[]; total: number; page: number; size: number }>(`/albums?${q}`)
  }

  async getAlbumBySlug(slug: string) {
    return this.get<any>(`/albums/slug/${slug}`)
  }

  // ── Search ────────────────────────────────────────────────────────────

  async searchAll(query: string) {
    return this.get<{ albums: any[]; query: string }>(`/search?q=${encodeURIComponent(query)}`)
  }

  async searchAlbums(query: string) {
    return this.get<any[]>(`/search/albums?q=${encodeURIComponent(query)}`)
  }

  async getPopularSearches(): Promise<string[]> {
    return this.get<string[]>('/search/popular')
  }

  // ── Achievements ──────────────────────────────────────────────────────

  async getUserAchievements(userId: string): Promise<Achievement[]> {
    return this.get<Achievement[]>(`/users/${userId}/achievements`)
  }

  async getMyAchievements(): Promise<Achievement[]> {
    return this.get<Achievement[]>('/users/me/achievements')
  }

  async getMyPoints(): Promise<{ total: number; level: number }> {
    return this.get<{ total: number; level: number }>('/users/me/points')
  }

  async getLeaderboard(type?: 'all' | 'weekly'): Promise<LeaderboardEntry[]> {
    if (type === 'weekly') {
      return this.get<LeaderboardEntry[]>('/leaderboard/weekly')
    }
    return this.get<LeaderboardEntry[]>('/leaderboard')
  }

  // ── User Profile ──────────────────────────────────────────────────────

  async getMe() {
    return this.get<any>('/users/me')
  }

  async updateProfile(data: {
    bio?: string
    website?: string
    location?: string
    furry_name?: string
    species?: string
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
}

export const apiClient = new ApiClient(API_BASE_URL)
