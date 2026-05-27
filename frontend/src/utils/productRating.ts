export function formatProductRating(ratingAvg: number, ratingCount: number): string {
  const avg = Number.isFinite(ratingAvg) ? ratingAvg : 0;
  const count = ratingCount > 0 ? ratingCount : 0;

  if (count === 0) {
    return "0 (0)";
  }

  return `${avg.toFixed(1)} (${count})`;
}

/** Рейтинг для таблицы админки: 4.9★(1360) */
export function formatAdminProductRating(
  ratingAvg: number,
  ratingCount: number
): string {
  const count = ratingCount > 0 ? ratingCount : 0;

  if (count === 0) {
    return "—";
  }

  const avg = Number.isFinite(ratingAvg) ? ratingAvg : 0;

  return `${avg.toFixed(1)}★(${count})`;
}
