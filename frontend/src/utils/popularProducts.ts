import type { Product } from "../api/productApi";

/** Минимум отзывов, чтобы считать товар популярным */
export const MIN_POPULAR_REVIEWS = 1;

/** Минимальная средняя оценка (из 5) */
export const MIN_POPULAR_RATING = 4;

const HOME_POPULAR_LIMIT = 8;

export function getProductPopularityScore(product: Product): number {
  const count = product.rating_count ?? 0;
  const avg = product.rating_avg ?? 0;

  if (count < MIN_POPULAR_REVIEWS || avg < MIN_POPULAR_RATING) {
    return 0;
  }

  // Учитываем и оценку, и число отзывов
  return avg * count;
}

export function isPopularProduct(product: Product): boolean {
  return getProductPopularityScore(product) > 0;
}

export function sortByPopularity(products: Product[]): Product[] {
  return [...products].sort((a, b) => {
    const scoreDiff = getProductPopularityScore(b) - getProductPopularityScore(a);
    if (scoreDiff !== 0) {
      return scoreDiff;
    }

    const countDiff = (b.rating_count ?? 0) - (a.rating_count ?? 0);
    if (countDiff !== 0) {
      return countDiff;
    }

    return (b.rating_avg ?? 0) - (a.rating_avg ?? 0);
  });
}

export function getPopularProducts(
  products: Product[],
  limit = HOME_POPULAR_LIMIT
): Product[] {
  return sortByPopularity(products.filter(isPopularProduct)).slice(0, limit);
}
