import axios from "axios";

const ACCESS_TOKEN_KEY = "gamegear_token";
const REFRESH_TOKEN_KEY = "gamegear_refresh_token";

export const tokenStorage = {
  getAccess: () => localStorage.getItem(ACCESS_TOKEN_KEY),
  getRefresh: () => localStorage.getItem(REFRESH_TOKEN_KEY),
  setTokens: (access: string, refresh: string) => {
    localStorage.setItem(ACCESS_TOKEN_KEY, access);
    localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
  },
  clear: () => {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  },
};

export const axiosInstance = axios.create({
  baseURL: "",
  headers: {
    "Content-Type": "application/json",
  },
});

axiosInstance.interceptors.request.use((config) => {
  const token = tokenStorage.getAccess();

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

let refreshPromise: Promise<string | null> | null = null;

axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    const newAccessToken = await refreshPromise;

    if (!newAccessToken) {
      tokenStorage.clear();
      return Promise.reject(error);
    }

    originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;

    return axiosInstance(originalRequest);
  }
);

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = tokenStorage.getRefresh();

  if (!refreshToken) {
    return null;
  }

  try {
    const response = await axios.post("/api/auth/refresh", {
      refresh_token: refreshToken,
    });

    const { access_token, refresh_token } = response.data;
    tokenStorage.setTokens(access_token, refresh_token);

    return access_token;
  } catch {
    tokenStorage.clear();
    return null;
  }
}
