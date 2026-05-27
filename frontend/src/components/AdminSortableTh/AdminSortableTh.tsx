import { ChevronDown, ChevronUp } from "lucide-react";
import styles from "./AdminSortableTh.module.css";

export type AdminSortColumn =
  | "title"
  | "category"
  | "brand"
  | "price"
  | "rating"
  | "status"
  | "stock"
  | "name"
  | "email"
  | "phone"
  | "role"
  | "created";

export type SortDirection = "asc" | "desc";

interface AdminSortableThProps {
  label: string;
  column: AdminSortColumn;
  activeColumn: AdminSortColumn | null;
  direction: SortDirection;
  onSort: (column: AdminSortColumn) => void;
}

export function AdminSortableTh({
  label,
  column,
  activeColumn,
  direction,
  onSort,
}: AdminSortableThProps) {
  const isActive = activeColumn === column;

  return (
    <th>
      <button
        type="button"
        className={styles.sortBtn}
        onClick={() => onSort(column)}
      >
        <span>{label}</span>
        <span className={styles.sortIcons} aria-hidden>
          <ChevronUp
            size={14}
            className={
              isActive && direction === "asc" ? styles.sortIconActive : ""
            }
          />
          <ChevronDown
            size={14}
            className={
              isActive && direction === "desc" ? styles.sortIconActive : ""
            }
          />
        </span>
      </button>
    </th>
  );
}

export function getDefaultSortDirection(
  column: AdminSortColumn
): SortDirection {
  switch (column) {
    case "price":
    case "rating":
    case "stock":
    case "status":
    case "created":
    case "role":
      return "desc";
    default:
      return "asc";
  }
}
