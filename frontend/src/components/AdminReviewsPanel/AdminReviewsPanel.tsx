import { Search, Star } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { authApi, type PublicUserProfile } from "../../api/authApi";
import {
  productApi,
  type AdminProductReview,
  type Product,
} from "../../api/productApi";
import { getProfileFullName } from "../../utils/userDisplay";
import styles from "./AdminReviewsPanel.module.css";

interface AdminReviewsPanelProps {
  products: Product[];
  onCountChange?: (count: number) => void;
}

function formatReviewDate(value?: string) {
  if (!value) {
    return "—";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "—";
  }

  return date.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function AdminReviewsPanel({
  products,
  onCountChange,
}: AdminReviewsPanelProps) {
  const [reviews, setReviews] = useState<AdminProductReview[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [ratingFilter, setRatingFilter] = useState("all");
  const [productFilter, setProductFilter] = useState("all");
  const [reviewUsers, setReviewUsers] = useState<
    Record<number, PublicUserProfile>
  >({});

  const productOptions = useMemo(
    () =>
      [...products]
        .filter((product) => product.rating_count > 0)
        .sort((a, b) => a.title.localeCompare(b.title, "ru")),
    [products]
  );

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError("");

      try {
        const data = await productApi.getAdminReviews({
          rating: ratingFilter === "all" ? undefined : Number(ratingFilter),
          productId:
            productFilter === "all" ? undefined : Number(productFilter),
          q: search.trim() || undefined,
        });

        if (!cancelled) {
          setReviews(data);
          onCountChange?.(data.length);
        }
      } catch {
        if (!cancelled) {
          setReviews([]);
          setError("Не удалось загрузить отзывы");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    const timer = window.setTimeout(() => {
      void load();
    }, search.trim() ? 300 : 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [search, ratingFilter, productFilter]);

  const reviewUserIds = useMemo(() => {
    const ids = reviews
      .map((review) => review.user_id)
      .filter((id): id is number => typeof id === "number" && id > 0);
    return [...new Set(ids)];
  }, [reviews]);

  useEffect(() => {
    if (reviewUserIds.length === 0) {
      setReviewUsers({});
      return;
    }

    let cancelled = false;

    authApi
      .getPublicUsersByIds(reviewUserIds)
      .then((users) => {
        if (cancelled) {
          return;
        }

        const map: Record<number, PublicUserProfile> = {};
        for (const user of users) {
          map[user.id] = user;
        }
        setReviewUsers(map);
      })
      .catch(() => {
        if (!cancelled) {
          setReviewUsers({});
        }
      });

    return () => {
      cancelled = true;
    };
  }, [reviewUserIds]);

  const getAuthorName = (review: AdminProductReview) => {
    const user = review.user_id ? reviewUsers[review.user_id] : undefined;
    if (user) {
      const fullName = getProfileFullName({
        firstName: user.first_name,
        lastName: user.last_name,
      });
      return fullName || user.nickname?.trim() || review.author || "Покупатель";
    }

    return review.author || "Покупатель";
  };

  return (
    <>
      <p className={styles.hint}>
        Показаны отзывы покупателей, получивших заказ. Оставить отзыв можно
        только после статуса «Получен».
      </p>

      <div className={styles.filtersRow}>
        <select
          className={styles.filterSelect}
          value={ratingFilter}
          onChange={(event) => setRatingFilter(event.target.value)}
        >
          <option value="all">Все оценки</option>
          {[5, 4, 3, 2, 1].map((rating) => (
            <option key={rating} value={String(rating)}>
              {rating} ★
            </option>
          ))}
        </select>

        <select
          className={styles.filterSelect}
          value={productFilter}
          onChange={(event) => setProductFilter(event.target.value)}
        >
          <option value="all">Все товары</option>
          {productOptions.map((product) => (
            <option key={product.id} value={String(product.id)}>
              {product.title}
            </option>
          ))}
        </select>

        <div className={styles.search}>
          <Search size={16} />
          <input
            placeholder="Поиск по товару, автору, тексту..."
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
      </div>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Товар</th>
              <th>Покупатель</th>
              <th>Оценка</th>
              <th>Отзыв</th>
              <th>Дата</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={5} className={styles.emptyCell}>
                  Загрузка...
                </td>
              </tr>
            ) : reviews.length === 0 ? (
              <tr>
                <td colSpan={5} className={styles.emptyCell}>
                  Отзывы не найдены
                </td>
              </tr>
            ) : (
              reviews.map((review) => (
                <tr
                  key={`${review.product_id}-${review.user_id ?? review.author}-${review.created_at}`}
                >
                  <td>
                    <Link
                      to={`/product/${review.product_id}`}
                      className={styles.productLink}
                    >
                      {review.product_title}
                    </Link>
                  </td>
                  <td>
                    <div className={styles.authorCell}>
                      <span className={styles.verifiedBadge}>Покупатель</span>
                      <strong>{getAuthorName(review)}</strong>
                    </div>
                  </td>
                  <td>
                    <div
                      className={styles.stars}
                      aria-label={`Оценка ${review.rating} из 5`}
                    >
                      {Array.from({ length: 5 }, (_, index) => (
                        <Star
                          key={index}
                          size={14}
                          fill={
                            index < review.rating ? "currentColor" : "none"
                          }
                          className={
                            index < review.rating
                              ? styles.starFilled
                              : styles.starEmpty
                          }
                        />
                      ))}
                    </div>
                  </td>
                  <td className={styles.textCell}>{review.text}</td>
                  <td className={styles.dateCell}>
                    {formatReviewDate(review.created_at)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </>
  );
}
