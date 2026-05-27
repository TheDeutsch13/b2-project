import { Heart, ShoppingCart, Star } from "lucide-react";
import { Link } from "react-router-dom";
import { ProductNoImage } from "../ProductNoImage/ProductNoImage";
import { formatProductRating } from "../../utils/productRating";
import styles from "./ProductCard.module.css";

interface ProductCardProps {
  id: number;
  title: string;
  description?: string;
  price: number;
  imageSrc?: string;
  ratingAvg?: number;
  ratingCount?: number;
  isFavorite?: boolean;
  onToggleFavorite?: () => void;
  onAddToCart?: () => void;
}

export function ProductCard({
  id,
  title,
  description,
  price,
  imageSrc,
  ratingAvg = 0,
  ratingCount = 0,
  isFavorite = false,
  onToggleFavorite,
  onAddToCart,
}: ProductCardProps) {
  const hasRating = ratingCount > 0;
  const starFill = hasRating ? "#fbbf24" : "none";
  const starStroke = hasRating ? "#fbbf24" : "var(--color-muted)";

  return (
    <article className={styles.card}>
      <div className={styles.imageWrap}>
        <Link to={`/product/${id}`} className={styles.imageLink}>
          {imageSrc ? (
            <img src={imageSrc} alt="" className={styles.image} />
          ) : (
            <ProductNoImage />
          )}
        </Link>
        {onToggleFavorite && (
          <button
            type="button"
            className={`${styles.favoriteBtn} ${isFavorite ? styles.favoriteActive : ""}`}
            aria-label={isFavorite ? "Убрать из избранного" : "Добавить в избранное"}
            onClick={(event) => {
              event.preventDefault();
              event.stopPropagation();
              onToggleFavorite();
            }}
          >
            <Heart
              size={18}
              fill={isFavorite ? "currentColor" : "none"}
              stroke="currentColor"
            />
          </button>
        )}
        <button
          type="button"
          className={styles.cartBtn}
          aria-label="В корзину"
          onClick={onAddToCart}
        >
          <ShoppingCart size={16} />
        </button>
      </div>
      <div className={styles.body}>
        <Link to={`/product/${id}`} className={styles.titleLink}>
          <h3 className={styles.title}>{title}</h3>
        </Link>
        {description && <p className={styles.desc}>{description}</p>}
        <div className={styles.footer}>
          <strong className={styles.price}>
            {price.toLocaleString("ru-RU")} ₽
          </strong>
          <span className={styles.rating}>
            <Star size={14} fill={starFill} stroke={starStroke} />
            {formatProductRating(ratingAvg, ratingCount)}
          </span>
        </div>
      </div>
    </article>
  );
}
