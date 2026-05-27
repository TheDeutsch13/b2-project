import { MapPin, Search, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { cdekApi, type CdekPickupPoint } from "../../api/cdekApi";
import styles from "./CdekPvzModal.module.css";

interface CdekPvzModalProps {
  open: boolean;
  city: string;
  selectedCode?: string;
  onClose: () => void;
  onSelect: (point: CdekPickupPoint) => void;
}

export function CdekPvzModal({
  open,
  city,
  selectedCode,
  onClose,
  onSelect,
}: CdekPvzModalProps) {
  const [query, setQuery] = useState(city);
  const [points, setPoints] = useState<CdekPickupPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    setQuery(city);
  }, [open, city]);

  useEffect(() => {
    if (!open || !query.trim()) {
      return;
    }

    const timer = window.setTimeout(() => {
      setLoading(true);
      setError("");

      cdekApi
        .getPickupPoints(query.trim())
        .then(setPoints)
        .catch(() => {
          setPoints([]);
          setError("Не удалось загрузить пункты выдачи");
        })
        .finally(() => setLoading(false));
    }, 300);

    return () => window.clearTimeout(timer);
  }, [open, query]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return points;
    }

    return points.filter(
      (point) =>
        point.name.toLowerCase().includes(q) ||
        point.address.toLowerCase().includes(q) ||
        point.code.toLowerCase().includes(q)
    );
  }, [points, query]);

  if (!open) {
    return null;
  }

  return (
    <div className={styles.root}>
      <button
        type="button"
        className={styles.backdrop}
        aria-label="Закрыть"
        onClick={onClose}
      />
      <div className={styles.modal} role="dialog" aria-label="Выбор пункта СДЭК">
        <header className={styles.header}>
          <h2>Выбор пункта выдачи СДЭК</h2>
          <button type="button" className={styles.closeBtn} onClick={onClose}>
            <X size={18} />
          </button>
        </header>

        <div className={styles.searchRow}>
          <Search size={16} />
          <input
            value={query}
            placeholder="Город или адрес пункта"
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        <p className={styles.hint}>
          Пункты выдачи в г. {query.trim() || "—"}
          {loading ? " · загрузка…" : ` · ${filtered.length}`}
        </p>

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.list}>
          {filtered.length === 0 && !loading ? (
            <p className={styles.empty}>Пункты не найдены</p>
          ) : (
            filtered.map((point) => (
              <button
                key={point.code}
                type="button"
                className={`${styles.pointCard} ${
                  point.code === selectedCode ? styles.pointCardActive : ""
                }`}
                onClick={() => {
                  onSelect(point);
                  onClose();
                }}
              >
                <MapPin size={18} />
                <div>
                  <strong>
                    {point.code}, {point.name}
                  </strong>
                  <p>{point.address}</p>
                  <p>
                    Режим работы: {point.work_time || "—"}
                    {point.phone ? ` · ${point.phone}` : ""}
                  </p>
                </div>
              </button>
            ))
          )}
        </div>

        <footer className={styles.footer}>
          <button type="button" className={styles.cancelBtn} onClick={onClose}>
            Отмена
          </button>
        </footer>
      </div>
    </div>
  );
}
