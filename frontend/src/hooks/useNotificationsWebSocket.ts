import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "../app/hooks";
import { addNotification } from "../features/notifications/notificationsSlice";
import {
  appendSupportMessage,
  bumpSupportEvent,
  incrementSupportUnread,
} from "../features/support/supportSlice";
import { tokenStorage } from "../api/axiosInstance";
import { isUserRole } from "../constants/roles";

interface WsPayload {
  type: string;
  message: string;
  order_id?: number;
  support_thread_id?: number;
  support_message_id?: number;
  support_body?: string;
  support_sender_id?: number;
  support_sender_role?: string;
  support_sender_name?: string;
  target_user_id?: number;
}

export function useNotificationsWebSocket() {
  const dispatch = useAppDispatch();
  const token = useAppSelector((state) => state.auth.token);
  const user = useAppSelector((state) => state.auth.user);

  useEffect(() => {
    if (!token) {
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const socket = new WebSocket(
      `${protocol}://${window.location.host}/ws/notifications?token=${tokenStorage.getAccess()}`
    );

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as WsPayload;

        if (data.type === "support_message" && data.support_message_id) {
          const isStaff =
            user &&
            isUserRole(user.role) &&
            (user.role === "admin" || user.role === "moderator");
          const isTargetUser = user?.id === data.target_user_id;
          const isOwnMessage = user?.id === data.support_sender_id;

          if ((isTargetUser || isStaff) && !isOwnMessage) {
            dispatch(
              appendSupportMessage({
                id: data.support_message_id,
                thread_id: data.support_thread_id ?? 0,
                sender_id: data.support_sender_id ?? 0,
                sender_role:
                  data.support_sender_role === "staff" ? "staff" : "user",
                sender_name: data.support_sender_name ?? "",
                body: data.support_body ?? "",
                created_at: new Date().toISOString(),
              })
            );
            dispatch(bumpSupportEvent());
            dispatch(incrementSupportUnread());
          }

          return;
        }

        dispatch(
          addNotification({
            type: data.type,
            message: data.message,
            orderId: data.order_id,
          })
        );
      } catch {
        // ignore malformed messages
      }
    };

    return () => {
      socket.close();
    };
  }, [dispatch, token, user?.id, user?.role]);
}
