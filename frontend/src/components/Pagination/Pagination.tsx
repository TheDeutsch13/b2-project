import {
  buildPaginationRange,
  getTotalPages,
} from "../../utils/pagination";
import styles from "./Pagination.module.css";

interface PaginationProps {
  currentPage: number;
  totalItems: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

export function Pagination({
  currentPage,
  totalItems,
  pageSize,
  onPageChange,
}: PaginationProps) {
  const totalPages = getTotalPages(totalItems, pageSize);

  if (totalPages <= 1) {
    return null;
  }

  const safePage = Math.min(Math.max(1, currentPage), totalPages);
  const pages = buildPaginationRange(safePage, totalPages);

  return (
    <nav className={styles.pagination} aria-label="Пагинация">
      <button
        type="button"
        className={styles.pageBtn}
        disabled={safePage <= 1}
        onClick={() => onPageChange(safePage - 1)}
        aria-label="Предыдущая страница"
      >
        ‹
      </button>

      {pages.map((token, index) =>
        token === "ellipsis" ? (
          <span key={`ellipsis-${index}`} className={styles.ellipsis}>
            …
          </span>
        ) : (
          <button
            key={token}
            type="button"
            className={`${styles.pageBtn} ${token === safePage ? styles.pageActive : ""}`}
            onClick={() => onPageChange(token)}
            aria-current={token === safePage ? "page" : undefined}
          >
            {token}
          </button>
        )
      )}

      <button
        type="button"
        className={styles.pageBtn}
        disabled={safePage >= totalPages}
        onClick={() => onPageChange(safePage + 1)}
        aria-label="Следующая страница"
      >
        ›
      </button>
    </nav>
  );
}
