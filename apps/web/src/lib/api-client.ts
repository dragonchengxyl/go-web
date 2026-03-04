const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
  timestamp?: number
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

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
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

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  async post<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async put<T>(endpoint: string, body?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  async delete<T>(endpoint: string): Promise<T> {
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
}

export const apiClient = new ApiClient(API_BASE_URL)
