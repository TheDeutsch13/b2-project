import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import {
  authApi,
  type UpdateProfilePayload,
  type UserProfile,
} from "../../api/authApi";

interface ProfileState {
  firstName: string;
  lastName: string;
  nickname: string;
  birthDate: string;
  gender: "" | "male" | "female";
  phone: string;
  avatarUrl: string;
  isLoading: boolean;
  isSaving: boolean;
  error: string | null;
}

const initialState: ProfileState = {
  firstName: "",
  lastName: "",
  nickname: "",
  birthDate: "",
  gender: "",
  phone: "",
  avatarUrl: "",
  isLoading: false,
  isSaving: false,
  error: null,
};

function applyUserProfile(state: ProfileState, profile: UserProfile) {
  state.firstName = profile.first_name ?? "";
  state.lastName = profile.last_name ?? "";
  state.nickname = profile.nickname ?? "";
  state.birthDate = profile.birth_date?.slice(0, 10) ?? "";
  state.gender = profile.gender ?? "";
  state.phone = profile.phone ?? "";
  state.avatarUrl = profile.avatar_url ?? "";
}

export const fetchProfile = createAsyncThunk(
  "profile/fetchProfile",
  async (_, { rejectWithValue }) => {
    try {
      return await authApi.me();
    } catch {
      return rejectWithValue("Не удалось загрузить профиль");
    }
  }
);

export const saveProfile = createAsyncThunk(
  "profile/saveProfile",
  async (payload: UpdateProfilePayload, { rejectWithValue }) => {
    try {
      return await authApi.updateProfile(payload);
    } catch {
      return rejectWithValue("Не удалось сохранить профиль");
    }
  }
);

export const uploadAvatar = createAsyncThunk(
  "profile/uploadAvatar",
  async (file: File, { rejectWithValue }) => {
    try {
      return await authApi.uploadAvatar(file);
    } catch {
      return rejectWithValue("Не удалось загрузить фото");
    }
  }
);

export const removeAvatar = createAsyncThunk(
  "profile/removeAvatar",
  async (payload: UpdateProfilePayload, { rejectWithValue }) => {
    try {
      const profile = await authApi.updateProfile({
        ...payload,
        avatar_url: "",
      });
      return profile.avatar_url;
    } catch {
      return rejectWithValue("Не удалось удалить фото");
    }
  }
);

const profileSlice = createSlice({
  name: "profile",
  initialState,
  reducers: {
    clearProfile(state) {
      Object.assign(state, initialState);
    },
    setAvatarPreview(state, action: { payload: string }) {
      state.avatarUrl = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchProfile.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchProfile.fulfilled, (state, action) => {
        state.isLoading = false;
        applyUserProfile(state, action.payload);
      })
      .addCase(fetchProfile.rejected, (state, action) => {
        state.isLoading = false;
        state.error =
          typeof action.payload === "string"
            ? action.payload
            : "Ошибка загрузки профиля";
      })
      .addCase(saveProfile.pending, (state) => {
        state.isSaving = true;
        state.error = null;
      })
      .addCase(saveProfile.fulfilled, (state, action) => {
        state.isSaving = false;
        applyUserProfile(state, action.payload);
      })
      .addCase(saveProfile.rejected, (state, action) => {
        state.isSaving = false;
        state.error =
          typeof action.payload === "string"
            ? action.payload
            : "Ошибка сохранения";
      })
      .addCase(uploadAvatar.pending, (state) => {
        state.isSaving = true;
        state.error = null;
      })
      .addCase(uploadAvatar.fulfilled, (state, action) => {
        state.isSaving = false;
        state.avatarUrl = action.payload;
      })
      .addCase(uploadAvatar.rejected, (state, action) => {
        state.isSaving = false;
        state.error =
          typeof action.payload === "string"
            ? action.payload
            : "Ошибка загрузки фото";
      })
      .addCase(removeAvatar.pending, (state) => {
        state.isSaving = true;
        state.error = null;
      })
      .addCase(removeAvatar.fulfilled, (state) => {
        state.isSaving = false;
        state.avatarUrl = "";
      })
      .addCase(removeAvatar.rejected, (state, action) => {
        state.isSaving = false;
        state.error =
          typeof action.payload === "string"
            ? action.payload
            : "Ошибка удаления фото";
      });
  },
});

export const { clearProfile, setAvatarPreview } = profileSlice.actions;
export default profileSlice.reducer;
