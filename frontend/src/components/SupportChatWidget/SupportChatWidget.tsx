import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { MessageCircle, Send, X } from "lucide-react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { supportApi } from "../../api/supportApi";
import {
  appendSupportMessage,
  clearSupportUnread,
  setSupportThread,
} from "../../features/support/supportSlice";
import { getProfileDisplayName } from "../../utils/userDisplay";
import { scrollContainerToBottom } from "../../utils/scrollContainerToBottom";
import styles from "./SupportChatWidget.module.css";

const AUTH_PATHS = ["/login", "/register"];

function formatMessageTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }

  return date.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function SupportChatWidget() {
  const dispatch = useAppDispatch();
  const location = useLocation();
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);
  const messages = useAppSelector((state) => state.support.messages);
  const unreadCount = useAppSelector((state) => state.support.unreadCount);
  const lastEventAt = useAppSelector((state) => state.support.lastEventAt);

  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(false);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const shouldScrollRef = useRef(false);
  const prevCountRef = useRef(0);

  const isAuthPage = AUTH_PATHS.includes(location.pathname);
  const displayName = user
    ? getProfileDisplayName(user, profile)
    : "Пользователь";

  const loadThread = useCallback(async () => {
    if (!user) {
      return;
    }

    setLoading(true);
    try {
      const data = await supportApi.getMyThread();
      dispatch(
        setSupportThread({
          threadId: data.thread.id,
          messages: data.messages,
        })
      );
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [dispatch, user]);

  useEffect(() => {
    if (user && open) {
      shouldScrollRef.current = true;
      void loadThread();
      dispatch(clearSupportUnread());
    }
  }, [user, open, loadThread, dispatch]);

  useEffect(() => {
    if (open && user) {
      shouldScrollRef.current = false;
      void loadThread();
    }
  }, [lastEventAt, open, user, loadThread]);

  useEffect(() => {
    if (!open) {
      return;
    }

    const grew = messages.length > prevCountRef.current;
    if (!shouldScrollRef.current && !grew) {
      return;
    }

    requestAnimationFrame(() => {
      scrollContainerToBottom(messagesContainerRef.current, grew);
      shouldScrollRef.current = false;
      prevCountRef.current = messages.length;
    });
  }, [messages, open]);

  const handleSend = async () => {
    const body = draft.trim();
    if (!user || !body || sending) {
      return;
    }

    setSending(true);
    try {
      const message = await supportApi.sendMyMessage(body, displayName);
      dispatch(appendSupportMessage(message));
      setDraft("");
      requestAnimationFrame(() => {
        scrollContainerToBottom(messagesContainerRef.current, true);
      });
    } catch {
      // ignore
    } finally {
      setSending(false);
    }
  };

  if (isAuthPage) {
    return null;
  }

  return (
    <div className={styles.root}>
      {open && (
        <div className={styles.panel} role="dialog" aria-label="Чат поддержки">
          <div className={styles.header}>
            <div>
              <h3>Поддержка GAMEGEAR</h3>
              <p>Ответим в рабочее время</p>
            </div>
            <button
              type="button"
              className={styles.iconBtn}
              aria-label="Свернуть чат"
              onClick={() => setOpen(false)}
            >
              <X size={18} />
            </button>
          </div>

          {!user ? (
            <div className={styles.loginPrompt}>
              <p>Войдите в аккаунт, чтобы написать в поддержку.</p>
              <Link to="/login" className={styles.loginLink}>
                Войти
              </Link>
            </div>
          ) : (
            <>
              <div className={styles.messages} ref={messagesContainerRef}>
                {loading && messages.length === 0 ? (
                  <p className={styles.empty}>Загрузка...</p>
                ) : messages.length === 0 ? (
                  <p className={styles.empty}>
                    Здравствуйте! Опишите вопрос — команда поддержки ответит в этом
                    чате.
                  </p>
                ) : (
                  messages.map((message) => {
                    const isUser = message.sender_role === "user";

                    return (
                      <div
                        key={message.id}
                        className={`${styles.bubble} ${
                          isUser ? styles.bubbleUser : styles.bubbleStaff
                        }`}
                      >
                        {!isUser && (
                          <span className={styles.bubbleMeta}>
                            {message.sender_name || "Поддержка"}
                          </span>
                        )}
                        {message.body}
                        <span className={styles.bubbleTime}>
                          {formatMessageTime(message.created_at)}
                        </span>
                      </div>
                    );
                  })
                )}
              </div>

              <div className={styles.composer}>
                <input
                  type="text"
                  placeholder="Ваше сообщение..."
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      void handleSend();
                    }
                  }}
                />
                <button
                  type="button"
                  aria-label="Отправить"
                  disabled={sending || !draft.trim()}
                  onClick={() => void handleSend()}
                >
                  <Send size={18} />
                </button>
              </div>
            </>
          )}
        </div>
      )}

      <button
        type="button"
        className={styles.fab}
        aria-label="Открыть чат поддержки"
        onClick={() => {
          setOpen((value) => !value);
          if (!open) {
            dispatch(clearSupportUnread());
          }
        }}
      >
        <MessageCircle size={24} />
        {!open && unreadCount > 0 && (
          <span className={styles.fabBadge}>
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>
    </div>
  );
}
