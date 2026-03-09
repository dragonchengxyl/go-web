const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

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

  // Auth APIs
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

  // Game APIs
  async getGames(params?: {
    page?: number
    page_size?: number
    search?: string
    genre?: string
    tag?: string
  }) {
    const query = new URLSearchParams()
    if (params?.page) query.append('page', params.page.toString())
    if (params?.page_size) query.append('page_size', params.page_size.toString())
    if (params?.search) query.append('search', params.search)
    if (params?.genre) query.append('genre', params.genre)
    if (params?.tag) query.append('tag', params.tag)

    return this.get<{ games: any[]; total: number; page: number; size: number }>(
      `/games?${query.toString()}`
    )
  }

  async getGameBySlug(slug: string) {
    return this.get<any>(`/games/slug/${slug}`)
  }

  // Music APIs
  async getAlbums(params?: {
    page?: number
    page_size?: number
    search?: string
  }) {
    const query = new URLSearchParams()
    if (params?.page) query.append('page', params.page.toString())
    if (params?.page_size) query.append('page_size', params.page_size.toString())
    if (params?.search) query.append('search', params.search)

    return this.get<{ albums: any[]; total: number; page: number; size: number }>(
      `/albums?${query.toString()}`
    )
  }

  async getAlbumBySlug(slug: string) {
    return this.get<any>(`/albums/slug/${slug}`)
  }

  // Product APIs
  async getProducts(params?: {
    product_type?: string
    is_active?: boolean
    page?: number
    page_size?: number
  }) {
    const query = new URLSearchParams()
    if (params?.product_type) query.append('product_type', params.product_type)
    if (params?.is_active !== undefined) query.append('is_active', params.is_active.toString())
    if (params?.page) query.append('page', params.page.toString())
    if (params?.page_size) query.append('page_size', params.page_size.toString())

    return this.get<{ products: any[]; total: number; page: number; size: number }>(
      `/products?${query.toString()}`
    )
  }

  // Order APIs
  async createOrder(items: { product_id: string }[], couponCode?: string, idempotencyKey?: string) {
    return this.post<any>('/orders', {
      items,
      coupon_code: couponCode,
      idempotency_key: idempotencyKey,
    })
  }

  async payOrder(orderId: string, paymentMethod: string) {
    return this.post<any>(`/orders/${orderId}/pay`, {
      payment_method: paymentMethod,
    })
  }

  async getOrders(page?: number, pageSize?: number) {
    const query = new URLSearchParams()
    if (page) query.append('page', page.toString())
    if (pageSize) query.append('page_size', pageSize.toString())

    return this.get<{ orders: any[]; total: number; page: number; size: number }>(
      `/orders?${query.toString()}`
    )
  }

  async getOrder(orderId: string) {
    return this.get<any>(`/orders/${orderId}`)
  }

  // Search APIs
  async searchAll(query: string) {
    return this.get<{ games: any[]; albums: any[]; query: string }>(
      `/search?q=${encodeURIComponent(query)}`
    )
  }

  async searchGames(query: string) {
    return this.get<any[]>(`/search/games?q=${encodeURIComponent(query)}`)
  }

  async searchAlbums(query: string) {
    return this.get<any[]>(`/search/albums?q=${encodeURIComponent(query)}`)
  }

  // Search - Popular
  async getPopularSearches(): Promise<string[]> {
    return this.get<string[]>('/search/popular')
  }

  // Coupon APIs
  async validateCoupon(code: string): Promise<{ valid: boolean; discount: number; discount_type: string }> {
    return this.get<{ valid: boolean; discount: number; discount_type: string }>(
      `/coupons/validate?code=${encodeURIComponent(code)}`
    )
  }

  async redeemCode(code: string): Promise<void> {
    return this.post<void>('/coupons/redeem', { code })
  }

  // Achievement APIs
  async getUserAchievements(userId: string): Promise<Achievement[]> {
    return this.get<Achievement[]>(`/users/${userId}/achievements`)
  }

  async getMyAchievements(): Promise<Achievement[]> {
    return this.get<Achievement[]>('/users/me/achievements')
  }

  async getMyPoints(): Promise<{ total: number; level: number }> {
    return this.get<{ total: number; level: number }>('/users/me/points')
  }

  // Leaderboard APIs
  async getLeaderboard(type?: 'all' | 'weekly'): Promise<LeaderboardEntry[]> {
    if (type === 'weekly') {
      return this.get<LeaderboardEntry[]>('/leaderboard/weekly')
    }
    return this.get<LeaderboardEntry[]>('/leaderboard')
  }

  // Upload APIs
  async uploadAvatar(file: File): Promise<{ url: string }> {
    const formData = new FormData()
    formData.append('file', file)

    const headers: Record<string, string> = {}
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${this.baseUrl}/upload/avatar`, {
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

  // Payment APIs (method-specific)
  async payOrderAlipay(orderId: string, returnUrl?: string): Promise<{ pay_url: string }> {
    return this.post<{ pay_url: string }>(`/orders/${orderId}/pay/alipay`, {
      return_url: returnUrl,
    })
  }

  async payOrderWechat(orderId: string): Promise<{ qr_code: string }> {
    return this.post<{ qr_code: string }>(`/orders/${orderId}/pay/wechat`, {})
  }
}

export const apiClient = new ApiClient(API_BASE_URL)
