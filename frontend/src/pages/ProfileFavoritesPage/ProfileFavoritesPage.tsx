import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { productApi, type Product } from "../../api/productApi";
import { ProductCard } from "../../components/ProductCard/ProductCard";
import { addToCart } from "../../features/cart/cartSlice";
import { toggleFavorite } from "../../features/favorites/favoritesSlice";
import { getProductPrimaryImage } from "../../utils/productDisplay";
import styles from "./ProfileFavoritesPage.module.css";

export function ProfileFavoritesPage() {
  const dispatch = useAppDispatch();
  const favoriteIds = useAppSelector((state) => state.favorites.ids);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    productApi
      .getProducts()
      .then(setProducts)
      .catch(() => setProducts([]))
      .finally(() => setLoading(false));
  }, []);

  const favoriteProducts = useMemo(
    () => products.filter((product) => favoriteIds.includes(product.id)),
    [products, favoriteIds]
  );

  return (
    <section>
      <h1 className={styles.title}>Избранное</h1>

      {loading ? (
        <p className={styles.empty}>Загрузка...</p>
      ) : favoriteProducts.length === 0 ? (
        <p className={styles.empty}>
          Пока нет избранных товаров.{" "}
          <Link to="/catalog">Перейти в каталог</Link>
        </p>
      ) : (
        <div className={styles.grid}>
          {favoriteProducts.map((product) => {
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
                isFavorite
                onToggleFavorite={() => dispatch(toggleFavorite(product.id))}
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
      )}
    </section>
  );
}
