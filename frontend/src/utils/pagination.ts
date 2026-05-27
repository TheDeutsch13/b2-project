export const CATALOG_PAGE_SIZE = 30;

export function getTotalPages(totalItems: number, pageSize: number): number {
  if (totalItems <= 0) {
    return 1;
  }

  return Math.ceil(totalItems / pageSize);
}

export function paginateItems<T>(
  items: T[],
  page: number,
  pageSize: number
): T[] {
  if (items.length === 0) {
    return [];
  }

  const totalPages = getTotalPages(items.length, pageSize);
  const safePage = Math.min(Math.max(1, page), totalPages);
  const start = (safePage - 1) * pageSize;

  return items.slice(start, start + pageSize);
}

export type PaginationToken = number | "ellipsis";

export function buildPaginationRange(
  currentPage: number,
  totalPages: number
): PaginationToken[] {
  if (totalPages <= 1) {
    return totalPages === 1 ? [1] : [];
  }

  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const pages: PaginationToken[] = [1];

  if (currentPage > 3) {
    pages.push("ellipsis");
  }

  const rangeStart = Math.max(2, currentPage - 1);
  const rangeEnd = Math.min(totalPages - 1, currentPage + 1);

  for (let page = rangeStart; page <= rangeEnd; page += 1) {
    pages.push(page);
  }

  if (currentPage < totalPages - 2) {
    pages.push("ellipsis");
  }

  pages.push(totalPages);

  return pages;
}
