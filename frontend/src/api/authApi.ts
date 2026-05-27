import type { UserRole } from "../constants/roles";
import { axiosInstance } from "./axiosInstance";

export type { UserRole };

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UserProfile {
  id: number;
  email: string;
  role: UserRole;
  first_name: string;
  last_name: string;
  nickname: string;
  birth_date?: string;
  gender: "" | "male" | "female";
  phone: string;
  avatar_url: string;
  created_at: string;
}

export interface PublicUserProfile {
  id: number;
  first_name: string;
  last_name: string;
  nickname: string;
  avatar_url: string;
}

export type AuthUserResponse = UserProfile;

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: UserProfile;
}

export interface RegisterResponse {
  id: number;
  email: string;
  role: UserRole;
}

export interface UpdateProfilePayload {
  first_name: string;
  last_name: string;
  nickname: string;
  birth_date?: string;
  gender: "" | "male" | "female";
  phone: string;
  avatar_url?: string;
}

export const authApi = {
  register: async (data: RegisterRequest): Promise<RegisterResponse> => {
    const response = await axiosInstance.post<RegisterResponse>(
      "/api/auth/register",
      data
    );
    return response.data;
  },

  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>(
      "/api/auth/login",
      data
    );
    return response.data;
  },

  refresh: async (refreshToken: string): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>("/api/auth/refresh", {
      refresh_token: refreshToken,
    });
    return response.data;
  },

  me: async (): Promise<UserProfile> => {
    const response = await axiosInstance.get<UserProfile>("/api/auth/me");
    return response.data;
  },

  updateProfile: async (payload: UpdateProfilePayload): Promise<UserProfile> => {
    const response = await axiosInstance.patch<UserProfile>(
      "/api/auth/profile",
      payload
    );
    return response.data;
  },

  uploadAvatar: async (file: File): Promise<string> => {
    const formData = new FormData();
    formData.append("file", file);

    const response = await axiosInstance.post<{ url: string }>(
      "/api/auth/upload/avatar",
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );

    return response.data.url;
  },

  getUsers: async (): Promise<UserProfile[]> => {
    const response = await axiosInstance.get<UserProfile[]>("/api/auth/users");
    return response.data;
  },

  updateUserRole: async (userId: number, role: UserRole): Promise<UserProfile> => {
    const response = await axiosInstance.patch<UserProfile>(
      `/api/auth/users/${userId}/role`,
      { role }
    );
    return response.data;
  },

  getPublicUsersByIds: async (ids: number[]): Promise<PublicUserProfile[]> => {
    const clean = [...new Set(ids.filter((id) => Number.isFinite(id) && id > 0))];
    if (clean.length === 0) {
      return [];
    }

    const response = await axiosInstance.get<PublicUserProfile[]>(
      "/api/auth/users/public",
      { params: { ids: clean.join(",") } }
    );
    return response.data;
  },
};
