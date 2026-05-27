import { useEffect, useState } from "react";
import { StarRating } from "../StarRating/StarRating";
import styles from "./LeaveReviewModal.module.css";

interface LeaveReviewModalProps {
  open: boolean;
  productTitle: string;
  initialRating?: number;
  initialText?: string;
  submitLabel?: string;
  isSubmitting?: boolean;
  onClose: () => void;
  onSubmit: (rating: number, text: string) => void;
}

export function LeaveReviewModal({
  open,
  productTitle,
  initialRating = 5,
  initialText = "",
  submitLabel = "Опубликовать",
  isSubmitting = false,
  onClose,
  onSubmit,
}: LeaveReviewModalProps) {
  const [rating, setRating] = useState(initialRating);
  const [text, setText] = useState(initialText);

  useEffect(() => {
    if (open) {
      setRating(initialRating);
      setText(initialText);
    }
  }, [open, initialRating, initialText]);

  if (!open) {
    return null;
  }

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();

    if (isSubmitting || !text.trim()) {
      return;
    }

    onSubmit(rating, text.trim());
  };

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div
        className={styles.modal}
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
      >
        <div className={styles.head}>
          <h2>Оставить отзыв</h2>
          <button type="button" className={styles.closeBtn} onClick={onClose}>
            ×
          </button>
        </div>

        <p className={styles.productTitle}>{productTitle}</p>

        <form className={styles.form} onSubmit={handleSubmit}>
          <label className={styles.field}>
            <span>Оценка</span>
            <StarRating value={rating} onChange={setRating} size={22} />
          </label>

          <label className={styles.field}>
            <span>Текст отзыва</span>
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Поделитесь впечатлениями о товаре"
              required
            />
          </label>

          <div className={styles.actions}>
            <button type="button" className={styles.cancelBtn} onClick={onClose}>
              Отмена
            </button>
            <button
              type="submit"
              className={styles.saveBtn}
              disabled={isSubmitting}
            >
              {isSubmitting ? "Сохранение…" : submitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
