import { apiClient } from './client';
import type { Game, ApiResponse, PaginatedResponse } from './types';

export class GameService {
  async listGames(params?: {
    page?: number;
    page_size?: number;
    search?: string;
  }): Promise<ApiResponse<PaginatedResponse<Game>>> {
    return apiClient.get('/games', { params });
  }

  async getGame(id: string): Promise<ApiResponse<Game>> {
    return apiClient.get(`/games/${id}`);
  }

  async getGameBySlug(slug: string): Promise<ApiResponse<Game>> {
    return apiClient.get(`/games/slug/${slug}`);
  }
}

export const gameService = new GameService();
