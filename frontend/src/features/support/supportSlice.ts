import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { SupportMessage } from "../../api/supportApi";

interface SupportState {
  threadId: number | null;
  messages: SupportMessage[];
  unreadCount: number;
  lastEventAt: number;
}

const initialState: SupportState = {
  threadId: null,
  messages: [],
  unreadCount: 0,
  lastEventAt: 0,
};

const supportSlice = createSlice({
  name: "support",
  initialState,
  reducers: {
    setSupportThread: (
      state,
      action: PayloadAction<{ threadId: number; messages: SupportMessage[] }>
    ) => {
      state.threadId = action.payload.threadId;
      state.messages = action.payload.messages;
    },
    appendSupportMessage: (state, action: PayloadAction<SupportMessage>) => {
      const exists = state.messages.some((item) => item.id === action.payload.id);
      if (!exists) {
        state.messages.push(action.payload);
      }
    },
    clearSupportUnread: (state) => {
      state.unreadCount = 0;
    },
    incrementSupportUnread: (state) => {
      state.unreadCount += 1;
    },
    bumpSupportEvent: (state) => {
      state.lastEventAt = Date.now();
    },
    resetSupport: () => initialState,
  },
});

export const {
  setSupportThread,
  appendSupportMessage,
  clearSupportUnread,
  incrementSupportUnread,
  bumpSupportEvent,
  resetSupport,
} = supportSlice.actions;

export default supportSlice.reducer;
