import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Star } from "lucide-react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { authApi, type PublicUserProfile } from "../../api/authApi";
import { productApi, type Product } from "../../api/productApi";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { AccountAvatar } from "../../components/AccountAvatar/AccountAvatar";
import { ProductCard } from "../../components/ProductCard/ProductCard";
import { ProductNoImage } from "../../components/ProductNoImage/ProductNoImage";
import { addToCart } from "../../features/cart/cartSlice";
import { toggleFavorite } from "../../features/favorites/favoritesSlice";
import {
  getProductGalleryImages,
  getProductPrimaryImage,
  isProductInStock,
} from "../../utils/productDisplay";
import { userAvatarUrl } from "../../utils/userDisplay";
import styles from "./ProductPage.module.css";

type TabKey = "description" | "specs" | "reviews";

const defaultVariants = ["Стандарт"];

export function ProductPage() {
  const { id } = useParams();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const favoriteIds = useAppSelector((state) => state.favorites.ids);
  const [product, setProduct] = useState<Product | null>(null);
  const [related, setRelated] = useState<Product[]>([]);
  const [activeImage, setActiveImage] = useState(0);
  const [selectedVariant, setSelectedVariant] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [activeTab, setActiveTab] = useState<TabKey>("description");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [reviewUsers, setReviewUsers] = useState<Record<number, PublicUserProfile>>(
    {}
  );

  const productId = Number(id);

  useEffect(() => {
    if (!productId || Number.isNaN(productId)) {
      setError("Товар не найден");
      setLoading(false);
      return;
    }

    const load = async () => {
      setLoading(true);

      try {
        const [productData, allProducts] = await Promise.all([
          productApi.getProduct(productId),
          productApi.getProducts(),
        ]);

        setProduct(productData);
        setRelated(
          allProducts
            .filter(
              (item) =>
                item.id !== productData.id &&
                item.category_id === productData.category_id
            )
            .slice(0, 10)
        );
        setError("");
      } catch {
        setProduct(null);
        setRelated([]);
        setError("Не удалось загрузить товар");
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [productId]);

  const reviewUserIds = useMemo(() => {
    const ids = (product?.reviews ?? [])
      .map((review) => review.user_id)
      .filter((value): value is number => typeof value === "number" && value > 0);

    return [...new Set(ids)];
  }, [product?.reviews]);

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

  const getReviewDisplayName = (review: { author: string; user_id?: number }) => {
    const user = review.user_id ? reviewUsers[review.user_id] : undefined;
    if (!user) {
      return review.author || "?";
    }

    const fullName = `${user.first_name ?? ""} ${user.last_name ?? ""}`.trim();
    return fullName || user.nickname || review.author || "?";
  };

  const getReviewAvatarUrl = (review: { user_id?: number }) => {
    const user = review.user_id ? reviewUsers[review.user_id] : undefined;
    return userAvatarUrl(user?.avatar_url);
  };

  const galleryImages = useMemo(
    () => (product ? getProductGalleryImages(product) : []),
    [product]
  );

  const hasGallery = galleryImages.length > 0;

  const variants = useMemo(() => {
    if (!product?.variants?.length) {
      return defaultVariants;
    }

    return product.variants;
  }, [product]);

  const inStock = product ? isProductInStock(product) : false;

  const handleAddToCart = () => {
    if (!product) {
      return;
    }

    const imageSrc = getProductPrimaryImage(product);

    for (let index = 0; index < quantity; index += 1) {
      dispatch(
        addToCart({
          productId: product.id,
          title: product.title,
          price: product.price,
          imageSrc,
        })
      );
    }
  };

  if (loading) {
    return (
      <div className="page">
        <Header />
        <main className={`container ${styles.main}`}>
          <p className={styles.loading}>Загрузка товара...</p>
        </main>
        <Footer />
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="page">
        <Header />
        <main className={`container ${styles.main}`}>
          <p className={styles.error}>{error || "Товар не найден"}</p>
          <Link to="/catalog">Вернуться в каталог</Link>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <div className="page">
      <Header />
      <main className={`container ${styles.main}`}>
        <nav className={styles.breadcrumbs}>
          <Link to="/">Главная</Link>
          <span>/</span>
          <Link to="/catalog">{product.category_name || "Каталог"}</Link>
          <span>/</span>
          <span>{product.title}</span>
        </nav>

        <section className={styles.hero}>
          <div
            className={`${styles.gallery} ${!hasGallery ? styles.gallerySingle : ""}`}
          >
            {hasGallery && (
              <div className={styles.thumbs}>
                {galleryImages.map((image, index) => (
                  <button
                    key={`${image}-${index}`}
                    type="button"
                    className={`${styles.thumbBtn} ${activeImage === index ? styles.thumbActive : ""}`}
                    onClick={() => setActiveImage(index)}
                  >
                    <img src={image} alt="" />
                  </button>
                ))}
              </div>
            )}
            <div className={styles.mainImage}>
              {hasGallery ? (
                <img src={galleryImages[activeImage]} alt={product.title} />
              ) : (
                <ProductNoImage />
              )}
            </div>
          </div>

          <div className={styles.info}>
            <h1>{product.title}</h1>

            <span className={styles.variantsLabel}>Доступные варианты:</span>
            <div className={styles.variants}>
              {variants.map((variant, index) => (
                <button
                  key={variant}
                  type="button"
                  className={
                    selectedVariant === index
                      ? styles.variantBtnActive
                      : styles.variantBtn
                  }
                  onClick={() => setSelectedVariant(index)}
                >
                  {variant}
                </button>
              ))}
            </div>

            <p className={styles.price}>
              {product.price.toLocaleString("ru-RU")} ₽
            </p>
            <span
              className={`${styles.stock} ${inStock ? "" : styles.stockOut}`}
            >
              {inStock ? "В наличии | Москва" : "Нет в наличии"}
            </span>

            <div className={styles.actions}>
              <button
                type="button"
                className={styles.addBtn}
                disabled={!inStock}
                onClick={handleAddToCart}
              >
                Добавить в корзину
              </button>
              <div className={styles.qty}>
                <button
                  type="button"
                  onClick={() => setQuantity((value) => Math.max(1, value - 1))}
                >
                  −
                </button>
                <span>{quantity}</span>
                <button
                  type="button"
                  onClick={() => setQuantity((value) => value + 1)}
                >
                  +
                </button>
              </div>
            </div>
          </div>
        </section>

        <section className={styles.tabs}>
          <div className={styles.tabList}>
            <button
              type="button"
              className={
                activeTab === "description" ? styles.tabBtnActive : styles.tabBtn
              }
              onClick={() => setActiveTab("description")}
            >
              Описание
            </button>
            <button
              type="button"
              className={
                activeTab === "specs" ? styles.tabBtnActive : styles.tabBtn
              }
              onClick={() => setActiveTab("specs")}
            >
              Характеристики
            </button>
            <button
              type="button"
              className={
                activeTab === "reviews" ? styles.tabBtnActive : styles.tabBtn
              }
              onClick={() => setActiveTab("reviews")}
            >
              Отзывы
            </button>
          </div>

          <div className={styles.tabPanel}>
            {activeTab === "description" && (
              <p className={styles.description}>
                {product.description ||
                  "Описание товара пока не добавлено."}
              </p>
            )}

            {activeTab === "specs" && (
              <>
                <h3 className={styles.specsTitle}>Технические характеристики</h3>
                {product.specifications?.length ? (
                  <div className={styles.specsList}>
                    {product.specifications.map((item) => (
                      <div key={`${item.label}-${item.value}`} className={styles.specRow}>
                        <span>{item.label}</span>
                        <span>{item.value}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className={styles.description}>
                    Характеристики пока не добавлены.
                  </p>
                )}
              </>
            )}

            {activeTab === "reviews" && (
              <>
                <h3 className={styles.reviewsTitle}>Отзывы</h3>
                {product.reviews?.length ? (
                  <div className={styles.reviewsList}>
                    {product.reviews.map((review, index) => (
                      <article
                        key={`${review.user_id ?? review.author}-${index}`}
                        className={styles.reviewCard}
                      >
                        <AccountAvatar
                          name={getReviewDisplayName(review)}
                          avatarUrl={getReviewAvatarUrl(review)}
                          size="sm"
                        />
                        <div>
                          <div className={styles.reviewHead}>
                            <strong>{getReviewDisplayName(review)}</strong>
                            <div className={styles.stars}>
                              {Array.from({ length: 5 }).map((_, starIndex) => (
                                <Star
                                  key={starIndex}
                                  size={14}
                                  fill={
                                    starIndex < review.rating
                                      ? "currentColor"
                                      : "none"
                                  }
                                  stroke="currentColor"
                                />
                              ))}
                            </div>
                          </div>
                          <p className={styles.reviewText}>{review.text}</p>
                        </div>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className={styles.description}>Отзывов пока нет.</p>
                )}
              </>
            )}
          </div>
        </section>

        {related.length > 0 && (
          <section className={styles.related}>
            <h2>Вам так же может понравиться</h2>
            <div className={styles.relatedGrid}>
              {related.map((item) => {
                const imageSrc = getProductPrimaryImage(item);

                return (
                  <ProductCard
                    key={item.id}
                    id={item.id}
                    title={item.title}
                    price={item.price}
                    imageSrc={imageSrc}
                    ratingAvg={item.rating_avg}
                    ratingCount={item.rating_count}
                    isFavorite={favoriteIds.includes(item.id)}
                    onToggleFavorite={
                      user
                        ? () => dispatch(toggleFavorite(item.id))
                        : undefined
                    }
                    onAddToCart={() =>
                      dispatch(
                        addToCart({
                          productId: item.id,
                          title: item.title,
                          price: item.price,
                          imageSrc,
                        })
                      )
                    }
                  />
                );
              })}
            </div>
          </section>
        )}
      </main>
      <Footer />
    </div>
  );
}
