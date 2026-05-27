import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export interface AppNotification {
  id: string;
  type: string;
  message: string;
  orderId?: number;
  createdAt: string;
}

interface NotificationsState {
  items: AppNotification[];
}

const initialState: NotificationsState = {
  items: [],
};

const notificationsSlice = createSlice({
  name: "notifications",
  initialState,
  reducers: {
    addNotification: (state, action: PayloadAction<Omit<AppNotification, "id" | "createdAt">>) => {
      state.items.unshift({
        ...action.payload,
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
      });

      if (state.items.length > 20) {
        state.items = state.items.slice(0, 20);
      }
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.items = state.items.filter((item) => item.id !== action.payload);
    },
    clearNotifications: (state) => {
      state.items = [];
    },
  },
});

export const { addNotification, removeNotification, clearNotifications } =
  notificationsSlice.actions;

/** Время показа toast-уведомления (мс) */
export const NOTIFICATION_AUTO_DISMISS_MS = 7000;
export default notificationsSlice.reducer;
