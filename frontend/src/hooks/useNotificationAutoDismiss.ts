import { useEffect, useRef } from "react";
import { useAppDispatch, useAppSelector } from "../app/hooks";
import {
  NOTIFICATION_AUTO_DISMISS_MS,
  removeNotification,
} from "../features/notifications/notificationsSlice";

export function useNotificationAutoDismiss() {
  const dispatch = useAppDispatch();
  const items = useAppSelector((state) => state.notifications.items);
  const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  useEffect(() => {
    const timers = timersRef.current;

    for (const item of items) {
      if (timers.has(item.id)) {
        continue;
      }

      const timeoutId = setTimeout(() => {
        dispatch(removeNotification(item.id));
        timers.delete(item.id);
      }, NOTIFICATION_AUTO_DISMISS_MS);

      timers.set(item.id, timeoutId);
    }

    for (const [id, timeoutId] of timers) {
      if (!items.some((item) => item.id === id)) {
        clearTimeout(timeoutId);
        timers.delete(id);
      }
    }
  }, [dispatch, items]);

  useEffect(() => {
    const timers = timersRef.current;

    return () => {
      for (const timeoutId of timers.values()) {
        clearTimeout(timeoutId);
      }
      timers.clear();
    };
  }, []);
}
