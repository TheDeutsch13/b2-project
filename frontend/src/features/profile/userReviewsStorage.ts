export interface UserReview {
  id: string;
  productId: number;
  productTitle: string;
  productImage?: string;
  rating: number;
  text: string;
  createdAt: string;
  orderId?: number;
}

const REVIEWS_PREFIX = "gamegear_user_reviews_";

function reviewsKey(email: string): string {
  return `${REVIEWS_PREFIX}${email.toLowerCase()}`;
}

/** Один товар — один отзыв (оставляем самый новый). */
export function dedupeReviewsByProduct(reviews: UserReview[]): UserReview[] {
  const byProduct = new Map<number, UserReview>();

  for (const review of reviews) {
    const existing = byProduct.get(review.productId);
    if (
      !existing ||
      new Date(review.createdAt).getTime() >
        new Date(existing.createdAt).getTime()
    ) {
      byProduct.set(review.productId, review);
    }
  }

  return [...byProduct.values()].sort(
    (a, b) =>
      new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

export function readUserReviews(email: string): UserReview[] {
  try {
    const raw = localStorage.getItem(reviewsKey(email));
    if (!raw) {
      return [];
    }

    const parsed = JSON.parse(raw) as UserReview[];
    if (!Array.isArray(parsed)) {
      return [];
    }

    const normalized = dedupeReviewsByProduct(parsed);
    if (normalized.length !== parsed.length) {
      writeUserReviews(email, normalized);
    }

    return normalized;
  } catch {
    return [];
  }
}

export function writeUserReviews(email: string, reviews: UserReview[]): void {
  localStorage.setItem(
    reviewsKey(email),
    JSON.stringify(dedupeReviewsByProduct(reviews))
  );
}

export function hasReviewForProduct(
  reviews: UserReview[],
  productId: number
): boolean {
  return reviews.some((review) => review.productId === productId);
}

export function upsertUserReview(email: string, review: UserReview): void {
  const withoutProduct = readUserReviews(email).filter(
    (item) => item.productId !== review.productId
  );
  writeUserReviews(email, [review, ...withoutProduct]);
}

export function deleteUserReview(email: string, reviewId: string): void {
  const next = readUserReviews(email).filter((item) => item.id !== reviewId);
  writeUserReviews(email, next);
}
