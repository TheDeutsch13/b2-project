import { useCallback, useEffect, useRef, useState } from "react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import {
  supportApi,
  type SupportMessage,
  type SupportThreadListItem,
} from "../../api/supportApi";
import { bumpSupportEvent } from "../../features/support/supportSlice";
import { getProfileDisplayName } from "../../utils/userDisplay";
import { scrollContainerToBottom } from "../../utils/scrollContainerToBottom";
import styles from "./SupportStaffPanel.module.css";

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

export function SupportStaffPanel() {
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);
  const lastEventAt = useAppSelector((state) => state.support.lastEventAt);

  const [threads, setThreads] = useState<SupportThreadListItem[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [selectedStatus, setSelectedStatus] = useState("open");
  const [messages, setMessages] = useState<SupportMessage[]>([]);
  const [selectedEmail, setSelectedEmail] = useState("");
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [showClosed, setShowClosed] = useState(false);
  const [actionMessage, setActionMessage] = useState("");

  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const shouldScrollOnLoadRef = useRef(false);
  const prevMessageCountRef = useRef(0);

  const staffName = user
    ? getProfileDisplayName(user, profile) || "Поддержка"
    : "Поддержка";

  const openCount = threads.filter((thread) => thread.status === "open").length;

  const loadThreads = useCallback(async () => {
    try {
      const data = await supportApi.listThreads(showClosed);
      setThreads(data);

      if (selectedId && !data.some((thread) => thread.id === selectedId)) {
        setSelectedId(data[0]?.id ?? null);
      } else if (selectedId === null && data.length > 0) {
        setSelectedId(data[0].id);
      }
    } catch {
      setThreads([]);
      setSelectedId(null);
    }
  }, [showClosed, selectedId]);

  const loadThread = useCallback(async (threadId: number, scrollToBottom: boolean) => {
    setLoading(true);
    shouldScrollOnLoadRef.current = scrollToBottom;

    try {
      const data = await supportApi.getThread(threadId);
      setMessages(data.messages);
      setSelectedEmail(data.thread.user_email);
      setSelectedStatus(data.thread.status);
      prevMessageCountRef.current = data.messages.length;
    } catch {
      setMessages([]);
      setSelectedStatus("open");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadThreads();
  }, [loadThreads, lastEventAt]);

  useEffect(() => {
    if (selectedId) {
      void loadThread(selectedId, true);
    }
  }, [selectedId, loadThread]);

  useEffect(() => {
    if (!selectedId) {
      return;
    }

    void loadThread(selectedId, false);
  }, [lastEventAt, selectedId, loadThread]);

  useEffect(() => {
    if (!shouldScrollOnLoadRef.current && messages.length <= prevMessageCountRef.current) {
      return;
    }

    requestAnimationFrame(() => {
      scrollContainerToBottom(
        messagesContainerRef.current,
        shouldScrollOnLoadRef.current
      );
      shouldScrollOnLoadRef.current = false;
      prevMessageCountRef.current = messages.length;
    });
  }, [messages]);

  const handleSelectThread = (threadId: number) => {
    setSelectedId(threadId);
    setActionMessage("");
  };

  const handleSend = async () => {
    const body = draft.trim();
    if (!selectedId || !body || sending) {
      return;
    }

    setSending(true);
    try {
      const message = await supportApi.sendStaffMessage(
        selectedId,
        body,
        staffName
      );
      setMessages((current) => [...current, message]);
      setSelectedStatus("open");
      setDraft("");
      dispatch(bumpSupportEvent());
      await loadThreads();
      requestAnimationFrame(() => {
        scrollContainerToBottom(messagesContainerRef.current, true);
      });
    } catch {
      setActionMessage("Не удалось отправить сообщение");
    } finally {
      setSending(false);
    }
  };

  const handleCloseThread = async () => {
    if (!selectedId) {
      return;
    }

    const confirmed = window.confirm(
      "Завершить обращение? Диалог скроется из активных. Клиент сможет написать снова — обращение откроется автоматически."
    );

    if (!confirmed) {
      return;
    }

    try {
      await supportApi.closeThread(selectedId);
      setSelectedStatus("closed");
      setActionMessage("Обращение завершено");
      setSelectedId(null);
      setMessages([]);
      await loadThreads();
    } catch {
      setActionMessage("Не удалось завершить обращение");
    }
  };

  const handleDeleteThread = async () => {
    if (!selectedId) {
      return;
    }

    const confirmed = window.confirm(
      "Удалить диалог безвозвратно? Все сообщения будут удалены."
    );

    if (!confirmed) {
      return;
    }

    try {
      await supportApi.deleteThread(selectedId);
      setActionMessage("Диалог удалён");
      setSelectedId(null);
      setMessages([]);
      await loadThreads();
    } catch {
      setActionMessage("Не удалось удалить диалог");
    }
  };

  const isClosed = selectedStatus === "closed";

  return (
    <div className={styles.wrap}>
      <aside className={styles.threadList}>
        <div className={styles.threadListHead}>
          <span>Активные ({openCount})</span>
          <label className={styles.showClosedToggle}>
            <input
              type="checkbox"
              checked={showClosed}
              onChange={(event) => {
                setShowClosed(event.target.checked);
                setSelectedId(null);
              }}
            />
            Закрытые
          </label>
        </div>
        <div className={styles.threadItems}>
          {threads.length === 0 ? (
            <p className={styles.empty}>Активных обращений нет</p>
          ) : (
            threads.map((thread) => (
              <button
                key={thread.id}
                type="button"
                className={`${styles.threadBtn} ${
                  selectedId === thread.id ? styles.threadBtnActive : ""
                }`}
                onClick={() => handleSelectThread(thread.id)}
              >
                <span className={styles.threadEmail}>
                  {thread.user_email || `Пользователь #${thread.user_id}`}
                  {thread.status === "closed" && (
                    <span className={styles.threadClosedBadge}> · закрыто</span>
                  )}
                </span>
                <span className={styles.threadPreview}>
                  {thread.last_message_body || "Нет сообщений"}
                </span>
              </button>
            ))
          )}
        </div>
      </aside>

      <section className={styles.chat}>
        {selectedId ? (
          <>
            <div className={styles.chatHead}>
              <div>
                <strong>{selectedEmail || "Клиент"}</strong>
                <span>
                  Диалог #{selectedId}
                  {isClosed ? " · завершено" : ""}
                </span>
              </div>
              <div className={styles.chatActions}>
                {!isClosed && (
                  <button
                    type="button"
                    className={styles.actionBtn}
                    onClick={() => void handleCloseThread()}
                  >
                    Завершить
                  </button>
                )}
                <button
                  type="button"
                  className={`${styles.actionBtn} ${styles.actionBtnDanger}`}
                  onClick={() => void handleDeleteThread()}
                >
                  Удалить
                </button>
              </div>
            </div>

            {actionMessage && (
              <p className={styles.actionHint}>{actionMessage}</p>
            )}

            <div className={styles.messages} ref={messagesContainerRef}>
              {loading && messages.length === 0 ? (
                <p className={styles.empty}>Загрузка...</p>
              ) : messages.length === 0 ? (
                <p className={styles.empty}>Сообщений пока нет</p>
              ) : (
                messages.map((message) => {
                  const isStaff = message.sender_role === "staff";

                  return (
                    <div
                      key={message.id}
                      className={`${styles.bubble} ${
                        isStaff ? styles.bubbleStaff : styles.bubbleUser
                      }`}
                    >
                      <span className={styles.bubbleMeta}>
                        {message.sender_name ||
                          (isStaff ? "Поддержка" : "Клиент")}{" "}
                        · {formatMessageTime(message.created_at)}
                      </span>
                      {message.body}
                    </div>
                  );
                })
              )}
            </div>

            <div className={styles.composer}>
              <input
                type="text"
                placeholder={
                  isClosed
                    ? "Обращение закрыто. Ответ откроет его снова."
                    : "Ответ клиенту..."
                }
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
                disabled={sending || !draft.trim()}
                onClick={() => void handleSend()}
              >
                Отправить
              </button>
            </div>
          </>
        ) : (
          <p className={styles.empty}>Выберите диалог слева</p>
        )}
      </section>
    </div>
  );
}
