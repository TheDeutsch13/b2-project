import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { authApi, type UserProfile } from "../../api/authApi";
import type { LoginRequest, RegisterRequest } from "../../api/authApi";
import { tokenStorage } from "../../api/axiosInstance";
import type { UserRole } from "../../constants/roles";

export type { UserRole };

export type AuthUser = Pick<UserProfile, "id" | "email" | "role">;

interface AuthState {
  user: AuthUser | null;
  token: string | null;
  refreshToken: string | null;
  isLoading: boolean;
  error: string | null;
  isInitialized: boolean;
}

const initialState: AuthState = {
  user: null,
  token: tokenStorage.getAccess(),
  refreshToken: tokenStorage.getRefresh(),
  isLoading: false,
  error: null,
  isInitialized: false,
};

export const registerUser = createAsyncThunk(
  "auth/registerUser",
  async (data: RegisterRequest, { rejectWithValue }) => {
    try {
      const user = await authApi.register(data);
      return user;
    } catch {
      return rejectWithValue("Не удалось зарегистрироваться");
    }
  }
);

export const loginUser = createAsyncThunk(
  "auth/loginUser",
  async (data: LoginRequest, { rejectWithValue }) => {
    try {
      return await authApi.login(data);
    } catch {
      return rejectWithValue("Неверный email или пароль");
    }
  }
);

export const initializeAuth = createAsyncThunk(
  "auth/initializeAuth",
  async (_, { rejectWithValue }) => {
    if (!tokenStorage.getAccess()) {
      return null;
    }

    try {
      return await authApi.me();
    } catch {
      tokenStorage.clear();
      return rejectWithValue("session expired");
    }
  }
);

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    logout: (state) => {
      state.user = null;
      state.token = null;
      state.refreshToken = null;
      tokenStorage.clear();
    },
    clearAuthError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(registerUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(registerUser.fulfilled, (state) => {
        state.isLoading = false;
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error =
          typeof action.payload === "string"
            ? action.payload
            : "Ошибка регистрации";
      })
      .addCase(loginUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.token = action.payload.access_token;
        state.refreshToken = action.payload.refresh_token;
        tokenStorage.setTokens(
          action.payload.access_token,
          action.payload.refresh_token
        );
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error =
          typeof action.payload === "string" ? action.payload : "Ошибка входа";
      })
      .addCase(initializeAuth.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(initializeAuth.fulfilled, (state, action) => {
        state.isLoading = false;
        state.isInitialized = true;
        if (action.payload) {
          state.user = action.payload;
        }
      })
      .addCase(initializeAuth.rejected, (state) => {
        state.isLoading = false;
        state.isInitialized = true;
        state.user = null;
        state.token = null;
        state.refreshToken = null;
      });
  },
});

export const { logout, clearAuthError } = authSlice.actions;
export default authSlice.reducer;
