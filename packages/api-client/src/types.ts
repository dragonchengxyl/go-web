export interface Game {
  id: string;
  title: string;
  slug: string;
  description: string;
  cover_image: string;
  price_cents: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  description: string;
  product_type: 'game' | 'dlc' | 'ost' | 'bundle' | 'membership';
  entity_id?: string;
  price_cents: number;
  currency: string;
  original_price_cents?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  order_no: string;
  user_id: string;
  status: 'pending_payment' | 'paid' | 'fulfilled' | 'cancelled' | 'failed' | 'refunded';
  total_cents: number;
  currency: string;
  discount_cents: number;
  coupon_code?: string;
  payment_method?: 'alipay' | 'wechat' | 'stripe' | 'paypal';
  paid_at?: string;
  created_at: string;
  expires_at?: string;
  items?: OrderItem[];
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  price_cents: number;
  quantity: number;
  created_at: string;
}

export interface Album {
  id: string;
  title: string;
  slug: string;
  description: string;
  cover_image: string;
  release_date: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface Track {
  id: string;
  album_id: string;
  title: string;
  track_number: number;
  duration_seconds: number;
  audio_url: string;
  is_published: boolean;
  created_at: string;
}

export interface Comment {
  id: string;
  commentable_type: 'game' | 'album' | 'track';
  commentable_id: string;
  user_id: string;
  parent_id?: string;
  content: string;
  like_count: number;
  reply_count: number;
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  request_id: string;
  timestamp: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}
