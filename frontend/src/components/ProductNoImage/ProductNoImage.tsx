import { ImageOff } from "lucide-react";
import styles from "./ProductNoImage.module.css";

interface ProductNoImageProps {
  compact?: boolean;
  className?: string;
}

export function ProductNoImage({ compact = false, className = "" }: ProductNoImageProps) {
  return (
    <div
      className={`${styles.placeholder} ${compact ? styles.compact : ""} ${className}`}
      aria-label="Нет изображения"
    >
      <ImageOff size={compact ? 20 : 28} className={styles.icon} strokeWidth={1.5} />
      <span className={styles.text}>NO IMAGE</span>
    </div>
  );
}
