import { useEffect } from "react";
import { BrowserRouter } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "./app/hooks";
import { initializeAuth } from "./features/auth/authSlice";
import {
  clearFavoritesUser,
  setFavoritesUser,
} from "./features/favorites/favoritesSlice";
import { clearProfile, fetchProfile } from "./features/profile/profileSlice";
import { useNotificationAutoDismiss } from "./hooks/useNotificationAutoDismiss";
import { useNotificationsWebSocket } from "./hooks/useNotificationsWebSocket";
import { SupportChatWidget } from "./components/SupportChatWidget/SupportChatWidget";
import { resetSupport } from "./features/support/supportSlice";
import { AppRouter } from "./routes/AppRouter";

function AppContent() {
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);

  useNotificationsWebSocket();
  useNotificationAutoDismiss();

  useEffect(() => {
    dispatch(initializeAuth());
  }, [dispatch]);

  useEffect(() => {
    if (user?.email) {
      dispatch(setFavoritesUser(user.email));
      void dispatch(fetchProfile());
    } else {
      dispatch(clearFavoritesUser());
      dispatch(clearProfile());
      dispatch(resetSupport());
    }
  }, [dispatch, user?.id, user?.email]);

  return (
    <>
      <AppRouter />
      <SupportChatWidget />
    </>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}

export default App;
