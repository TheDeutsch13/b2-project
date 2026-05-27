import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowRight, Headphones, MousePointer, Shield, Zap } from "lucide-react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { productApi, type Category, type Product } from "../../api/productApi";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { ProductCard } from "../../components/ProductCard/ProductCard";
import { PromoCarousel } from "../../components/PromoCarousel/PromoCarousel";
import { addToCart } from "../../features/cart/cartSlice";
import { toggleFavorite } from "../../features/favorites/favoritesSlice";
import { carouselSlides } from "../../data/carousel.generated";
import { sortCategoriesByFixedOrder } from "../../data/categories";
import {
  fallbackPromoSlides,
  getCategoryImageSources,
  heroMiceImage,
} from "../../data/homeAssets";
import { getPopularProducts } from "../../utils/popularProducts";
import { getProductPrimaryImage } from "../../utils/productDisplay";
import styles from "./HomePage.module.css";

const features = [
  {
    icon: MousePointer,
    title: "Подбор под стиль игры",
    text: "Поможем выбрать мышь под хват и жанр",
  },
  {
    icon: Shield,
    title: "Премиальные бренды",
    text: "Только проверенные производители",
  },
  {
    icon: Zap,
    title: "Быстрая доставка",
    text: "Отправка в день заказа",
  },
  {
    icon: Headphones,
    title: "Поддержка после покупки",
    text: "Помощь с настройкой и гарантией",
  },
];

export function HomePage() {
  const dispatch = useAppDispatch();
  const favoriteIds = useAppSelector((state) => state.favorites.ids);
  const [allProducts, setAllProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);

  const promoSlides = carouselSlides.length > 0 ? carouselSlides : fallbackPromoSlides;

  useEffect(() => {
    Promise.all([productApi.getProducts(), productApi.getCategories()])
      .then(([productsData, categoriesData]) => {
        setAllProducts(productsData);
        setCategories(sortCategoriesByFixedOrder(categoriesData));
      })
      .catch(() => {
        setAllProducts([]);
        setCategories([]);
      });
  }, []);

  const productCountByCategory = useMemo(() => {
    const counts = new Map<number, number>();

    for (const product of allProducts) {
      if (!product.category_id) {
        continue;
      }

      counts.set(product.category_id, (counts.get(product.category_id) ?? 0) + 1);
    }

    return counts;
  }, [allProducts]);

  const popularProducts = useMemo(
    () => getPopularProducts(allProducts),
    [allProducts]
  );

  return (
    <div className="page">
      <Header />
      <main>
        <section className={styles.hero}>
          <img
            src="/home/hero.png"
            alt=""
            className={styles.heroBg}
            onError={(event) => {
              const img = event.currentTarget;
              if (!img.dataset.fallback) {
                img.dataset.fallback = "1";
                img.src = heroMiceImage;
              }
            }}
          />
          <div className={styles.heroOverlay} aria-hidden="true" />
          <div className={styles.heroContent}>
            <h1>
              LEVEL UP
              <br />
              <span>YOUR SETUP</span>
            </h1>
            <p>
              Премиальная игровая периферия для тех, кто ценит точность,
              скорость и стиль.
            </p>
            <Link to="/catalog" className={styles.heroBtn}>
              Смотреть каталог
            </Link>
          </div>
        </section>

        <section className={`container ${styles.features}`}>
          <h2 className={styles.sectionTitle}>Почему GAMEGEAR</h2>
          <div className={styles.featureGrid}>
            {features.map(({ icon: Icon, title, text }) => (
              <div key={title} className={styles.featureCard}>
                <div className={styles.featureIcon}>
                  <Icon size={22} />
                </div>
                <h3>{title}</h3>
                <p>{text}</p>
              </div>
            ))}
          </div>
        </section>

        <section className={`container ${styles.categories}`}>
          <div className={styles.sectionHead}>
            <h2 className={styles.sectionTitle}>Категории</h2>
            <Link to="/catalog" className={styles.sectionLink}>
              Смотреть все →
            </Link>
          </div>
          <div className={styles.categoryGrid}>
            {categories.map((category) => {
              const count = productCountByCategory.get(category.id) ?? 0;
              const { primary, fallback } = getCategoryImageSources(
                category.name
              );

              return (
                <Link
                  key={category.id}
                  to={`/catalog?category=${category.id}`}
                  className={styles.categoryCard}
                >
                  <div className={styles.categoryImage}>
                    <img
                      src={primary}
                      alt=""
                      onError={(event) => {
                        const img = event.currentTarget;
                        if (fallback && img.src !== fallback) {
                          img.src = fallback;
                        }
                      }}
                    />
                  </div>
                  <div className={styles.categoryBody}>
                    <div>
                      <h3>{category.name}</h3>
                      <span>
                        {count > 0 ? `${count}+ товаров` : "Смотреть каталог"}
                      </span>
                    </div>
                    <span className={styles.categoryArrow} aria-hidden="true">
                      <ArrowRight size={18} color="#ffffff" strokeWidth={2.5} />
                    </span>
                  </div>
                </Link>
              );
            })}
          </div>
        </section>

        <section className={`container ${styles.carouselSection}`}>
          <div className={styles.sectionHead}>
            <h2 className={styles.sectionTitle}>Новинки</h2>
          </div>
          <PromoCarousel slides={promoSlides} />
        </section>

        {popularProducts.length > 0 && (
          <section className={`container ${styles.catalogPreview}`}>
            <div className={styles.sectionHead}>
              <h2 className={styles.sectionTitle}>Популярные товары</h2>
              <Link to="/catalog" className={styles.sectionLink}>
                Смотреть все →
              </Link>
            </div>
            <p className={styles.popularHint}>
              Высокий рейтинг и много отзывов от покупателей
            </p>
            <div className={styles.productGrid}>
              {popularProducts.map((product) => {
                const imageSrc = getProductPrimaryImage(product);

                return (
                  <ProductCard
                    key={product.id}
                    id={product.id}
                    title={product.title}
                    description={product.description}
                    price={product.price}
                    imageSrc={imageSrc}
                    ratingAvg={product.rating_avg}
                    ratingCount={product.rating_count}
                    isFavorite={favoriteIds.includes(product.id)}
                    onToggleFavorite={() =>
                      dispatch(toggleFavorite(product.id))
                    }
                    onAddToCart={() =>
                      dispatch(
                        addToCart({
                          productId: product.id,
                          title: product.title,
                          price: product.price,
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
