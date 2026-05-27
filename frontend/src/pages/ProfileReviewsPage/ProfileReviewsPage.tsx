import { useCallback, useEffect, useMemo, useState } from "react";
import {
  productApi,
  productImageUrl,
  type MyProductReview,
  type Order,
} from "../../api/productApi";
import { useAppSelector } from "../../app/hooks";
import { LeaveReviewModal } from "../../components/LeaveReviewModal/LeaveReviewModal";
import { StarRating } from "../../components/StarRating/StarRating";
import {
  hasReviewForProduct,
  type UserReview,
} from "../../features/profile/userReviewsStorage";
import { canLeaveReviewForOrder } from "../../utils/orderDisplay";
import { getProfileDisplayName } from "../../utils/userDisplay";
import styles from "./ProfileReviewsPage.module.css";

interface PendingReviewItem {
  orderId: number;
  productId: number;
  productTitle: string;
  productImage?: string;
  receivedAt: string;
}

function mapMyReview(review: MyProductReview): UserReview {
  return {
    id: String(review.product_id),
    productId: review.product_id,
    productTitle: review.product_title,
    productImage: review.product_image,
    rating: review.rating,
    text: review.text,
    createdAt: review.created_at ?? new Date().toISOString(),
  };
}

export function ProfileReviewsPage() {
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);
  const [orders, setOrders] = useState<Order[]>([]);
  const [myReviews, setMyReviews] = useState<UserReview[]>([]);
  const [reviewTarget, setReviewTarget] = useState<PendingReviewItem | null>(
    null
  );
  const [editingReview, setEditingReview] = useState<UserReview | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const reloadReviews = useCallback(async () => {
    if (!user) {
      return;
    }

    try {
      const list = await productApi.getMyReviews();
      setMyReviews(list.map(mapMyReview));
    } catch {
      setMyReviews([]);
      setSaveError(
        "Не удалось загрузить отзывы. Перезапустите dev-сервер фронтенда (обновлён proxy для /api/reviews)."
      );
    }
  }, [user]);

  useEffect(() => {
    productApi
      .getMyOrders()
      .then(setOrders)
      .catch(() => setOrders([]));
  }, []);

  useEffect(() => {
    void reloadReviews();
  }, [reloadReviews]);

  const pendingItems = useMemo(() => {
    if (!user) {
      return [];
    }

    const items: PendingReviewItem[] = [];
    const seenProductIds = new Set<number>();

    for (const order of orders) {
      if (!canLeaveReviewForOrder(order.status)) {
        continue;
      }

      for (const item of order.items ?? []) {
        if (
          seenProductIds.has(item.product_id) ||
          hasReviewForProduct(myReviews, item.product_id)
        ) {
          continue;
        }

        seenProductIds.add(item.product_id);

        items.push({
          orderId: order.id,
          productId: item.product_id,
          productTitle: item.title,
          receivedAt: order.created_at,
        });
      }
    }

    return items;
  }, [orders, user, myReviews]);

  const handleSubmitReview = async (rating: number, text: string) => {
    if (!user || isSaving) {
      return;
    }

    const author = getProfileDisplayName(user, profile);
    const productId = editingReview?.productId ?? reviewTarget?.productId;

    if (!productId) {
      return;
    }

    setIsSaving(true);
    setSaveError(null);

    try {
      await productApi.upsertProductReview(productId, {
        author,
        rating,
        text,
      });

      setReviewTarget(null);
      setEditingReview(null);
      await reloadReviews();
    } catch {
      setSaveError(
        "Не удалось сохранить отзыв. Проверьте, что заказ в статусе «Получен», и перезапустите product-service."
      );
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteReview = async (review: UserReview) => {
    if (!user || isSaving) {
      return;
    }

    const confirmed = window.confirm(
      `Удалить отзыв на «${review.productTitle}»?`
    );

    if (!confirmed) {
      return;
    }

    setIsSaving(true);

    try {
      await productApi.deleteProductReview(review.productId);

      if (editingReview?.id === review.id) {
        setEditingReview(null);
      }

      await reloadReviews();
    } catch {
      window.alert("Не удалось удалить отзыв.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <section>
      <h1 className={styles.title}>Отзывы</h1>

      {saveError && <p className={styles.error}>{saveError}</p>}

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Ожидают отзыва</h2>
        <p className={styles.sectionSubtitle}>
          По каждому товару можно оставить только один отзыв
        </p>

        {pendingItems.length === 0 ? (
          <p className={styles.empty}>Нет товаров, ожидающих отзыва</p>
        ) : (
          <div className={styles.pendingList}>
            {pendingItems.map((item) => (
              <article
                key={item.productId}
                className={styles.pendingCard}
              >
                <div className={styles.thumb} />
                <div className={styles.pendingInfo}>
                  <strong>{item.productTitle}</strong>
                  <span>
                    Получен{" "}
                    {new Date(item.receivedAt).toLocaleDateString("ru-RU")}
                  </span>
                </div>
                <button
                  type="button"
                  className={styles.reviewBtn}
                  disabled={isSaving}
                  onClick={() => setReviewTarget(item)}
                >
                  Оставить отзыв
                </button>
              </article>
            ))}
          </div>
        )}
      </div>

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Мои отзывы</h2>
        <p className={styles.sectionSubtitle}>
          Опубликовано отзывов: {myReviews.length}
        </p>

        {myReviews.length === 0 ? (
          <p className={styles.empty}>Вы ещё не оставляли отзывов</p>
        ) : (
          <div className={styles.reviewsList}>
            {myReviews.map((review) => (
              <article key={review.id} className={styles.reviewCard}>
                <div className={styles.reviewHead}>
                  <div className={styles.thumb}>
                    {review.productImage && (
                      <img
                        src={productImageUrl(review.productImage)}
                        alt=""
                      />
                    )}
                  </div>
                  <div>
                    <strong>{review.productTitle}</strong>
                    <StarRating value={review.rating} />
                  </div>
                  <div className={styles.reviewActions}>
                    <button
                      type="button"
                      className={styles.editReviewBtn}
                      disabled={isSaving}
                      onClick={() => setEditingReview(review)}
                    >
                      Редактировать
                    </button>
                    <button
                      type="button"
                      className={styles.deleteReviewBtn}
                      disabled={isSaving}
                      onClick={() => void handleDeleteReview(review)}
                    >
                      Удалить
                    </button>
                  </div>
                </div>
                <p className={styles.reviewText}>{review.text}</p>
              </article>
            ))}
          </div>
        )}
      </div>

      <LeaveReviewModal
        open={reviewTarget !== null || editingReview !== null}
        productTitle={
          reviewTarget?.productTitle ?? editingReview?.productTitle ?? ""
        }
        initialRating={editingReview?.rating ?? 5}
        initialText={editingReview?.text ?? ""}
        submitLabel={editingReview ? "Сохранить" : "Опубликовать"}
        isSubmitting={isSaving}
        onClose={() => {
          if (isSaving) {
            return;
          }

          setReviewTarget(null);
          setEditingReview(null);
          setSaveError(null);
        }}
        onSubmit={(rating, text) => {
          void handleSubmitReview(rating, text);
        }}
      />
    </section>
  );
}
