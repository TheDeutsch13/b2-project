import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

const STORAGE_PREFIX = "gamegear_favorites_";

function readFavorites(storageKey: string): number[] {
  try {
    const raw = localStorage.getItem(storageKey);
    if (!raw) {
      return [];
    }

    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter((id): id is number => typeof id === "number");
  } catch {
    return [];
  }
}

function writeFavorites(storageKey: string, ids: number[]) {
  localStorage.setItem(storageKey, JSON.stringify(ids));
}

interface FavoritesState {
  storageKey: string;
  ids: number[];
}

const initialState: FavoritesState = {
  storageKey: "",
  ids: [],
};

const favoritesSlice = createSlice({
  name: "favorites",
  initialState,
  reducers: {
    setFavoritesUser(state, action: PayloadAction<string>) {
      const storageKey = `${STORAGE_PREFIX}${action.payload}`;
      state.storageKey = storageKey;
      state.ids = readFavorites(storageKey);
    },
    clearFavoritesUser(state) {
      state.storageKey = "";
      state.ids = [];
    },
    toggleFavorite(state, action: PayloadAction<number>) {
      const productId = action.payload;
      const exists = state.ids.includes(productId);

      state.ids = exists
        ? state.ids.filter((id) => id !== productId)
        : [...state.ids, productId];

      if (state.storageKey) {
        writeFavorites(state.storageKey, state.ids);
      }
    },
  },
});

export const { setFavoritesUser, clearFavoritesUser, toggleFavorite } =
  favoritesSlice.actions;

export default favoritesSlice.reducer;
