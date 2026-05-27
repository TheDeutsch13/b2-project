import { Star } from "lucide-react";
import styles from "./StarRating.module.css";

interface StarRatingProps {
  value: number;
  onChange?: (value: number) => void;
  size?: number;
}

export function StarRating({ value, onChange, size = 18 }: StarRatingProps) {
  return (
    <div className={styles.stars} role={onChange ? "radiogroup" : undefined}>
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          className={`${styles.star} ${star <= value ? styles.filled : ""}`}
          onClick={onChange ? () => onChange(star) : undefined}
          disabled={!onChange}
          aria-label={`${star} из 5`}
        >
          <Star size={size} fill={star <= value ? "currentColor" : "none"} />
        </button>
      ))}
    </div>
  );
}
