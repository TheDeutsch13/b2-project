import { configureStore } from "@reduxjs/toolkit";
import authReducer from "../features/auth/authSlice";
import cartReducer from "../features/cart/cartSlice";
import favoritesReducer from "../features/favorites/favoritesSlice";
import notificationsReducer from "../features/notifications/notificationsSlice";
import profileReducer from "../features/profile/profileSlice";
import supportReducer from "../features/support/supportSlice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    cart: cartReducer,
    favorites: favoritesReducer,
    notifications: notificationsReducer,
    profile: profileReducer,
    support: supportReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
