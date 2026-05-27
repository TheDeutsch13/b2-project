import { useEffect, useMemo, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { FilterSheet } from "../../components/FilterSheet/FilterSheet";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { productApi, type Category, type Product } from "../../api/productApi";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { Pagination } from "../../components/Pagination/Pagination";
import { ProductCard } from "../../components/ProductCard/ProductCard";
import { addToCart } from "../../features/cart/cartSlice";
import { toggleFavorite } from "../../features/favorites/favoritesSlice";
import { sortCategoriesByFixedOrder } from "../../data/categories";
import { getProductPrimaryImage } from "../../utils/productDisplay";
import {
  buildBrandFilterOptions,
  buildDynamicSpecFilters,
  categoryHasWeightFilter,
  productMatchesBrand,
  productMatchesPriceRange,
  productMatchesSearch,
  productMatchesSpecFilters,
  productMatchesWeightRange,
} from "../../utils/specFilters";
import {
  CATALOG_PAGE_SIZE,
  getTotalPages,
  paginateItems,
} from "../../utils/pagination";
import styles from "./CatalogPage.module.css";

type SortOption = "popular" | "price-asc" | "price-desc";

export function CatalogPage() {
  const dispatch = useAppDispatch();
  const favoriteIds = useAppSelector((state) => state.favorites.ids);
  const [searchParams] = useSearchParams();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<number | "all">("all");
  const [specFilters, setSpecFilters] = useState<Record<string, string>>({});
  const [priceMin, setPriceMin] = useState("");
  const [priceMax, setPriceMax] = useState("");
  const [weightMin, setWeightMin] = useState("");
  const [weightMax, setWeightMax] = useState("");
  const [selectedBrand, setSelectedBrand] = useState("");
  const [sortBy, setSortBy] = useState<SortOption>("popular");
  const [currentPage, setCurrentPage] = useState(1);
  const [error, setError] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);

  const searchQuery = searchParams.get("q") ?? "";

  const selectedCategoryName =
    selectedCategory === "all"
      ? undefined
      : categories.find((item) => item.id === selectedCategory)?.name;

  const dynamicFilters = useMemo(
    () => buildDynamicSpecFilters(products, selectedCategoryName),
    [products, selectedCategoryName]
  );

  const brandOptions = useMemo(
    () => buildBrandFilterOptions(products),
    [products]
  );

  const showWeightFilter = categoryHasWeightFilter(selectedCategoryName);

  const filteredProducts = useMemo(() => {
    let list = products.filter(
      (product) =>
        productMatchesSearch(product, searchQuery) &&
        productMatchesPriceRange(product, priceMin, priceMax) &&
        productMatchesBrand(product, selectedBrand) &&
        (!showWeightFilter ||
          productMatchesWeightRange(product, weightMin, weightMax)) &&
        productMatchesSpecFilters(product, dynamicFilters, specFilters)
    );

    if (sortBy === "price-asc") {
      list = [...list].sort((a, b) => a.price - b.price);
    } else if (sortBy === "price-desc") {
      list = [...list].sort((a, b) => b.price - a.price);
    } else {
      list = [...list].sort(
        (a, b) =>
          (b.rating_count ?? 0) - (a.rating_count ?? 0) ||
          (b.rating_avg ?? 0) - (a.rating_avg ?? 0)
      );
    }

    return list;
  }, [
    products,
    priceMin,
    priceMax,
    weightMin,
    weightMax,
    selectedBrand,
    showWeightFilter,
    specFilters,
    dynamicFilters,
    sortBy,
    searchQuery,
  ]);

  const totalPages = getTotalPages(
    filteredProducts.length,
    CATALOG_PAGE_SIZE
  );

  const paginatedProducts = useMemo(
    () => paginateItems(filteredProducts, currentPage, CATALOG_PAGE_SIZE),
    [filteredProducts, currentPage]
  );

  useEffect(() => {
    setCurrentPage(1);
  }, [
    selectedCategory,
    searchQuery,
    priceMin,
    priceMax,
    weightMin,
    weightMax,
    selectedBrand,
    specFilters,
    sortBy,
  ]);

  useEffect(() => {
    if (currentPage > totalPages) {
      setCurrentPage(totalPages);
    }
  }, [currentPage, totalPages]);

  useEffect(() => {
    const categoryFromUrl = searchParams.get("category");
    if (categoryFromUrl) {
      const parsed = Number(categoryFromUrl);
      if (!Number.isNaN(parsed)) {
        setSelectedCategory(parsed);
      }
    }
  }, [searchParams]);

  useEffect(() => {
    const load = async () => {
      try {
        const [productsData, categoriesData] = await Promise.all([
          productApi.getProducts(
            selectedCategory === "all" ? undefined : selectedCategory
          ),
          productApi.getCategories(),
        ]);
        setProducts(productsData);
        setCategories(sortCategoriesByFixedOrder(categoriesData));
        setError("");
      } catch {
        setError("Не удалось загрузить каталог");
      }
    };

    load();
  }, [selectedCategory]);

  useEffect(() => {
    setSpecFilters({});
    setWeightMin("");
    setWeightMax("");
    setSelectedBrand("");
  }, [selectedCategory]);

  const selectCategory = (categoryId: number | "all") => {
    setSelectedCategory(categoryId);
    setSpecFilters({});
    setWeightMin("");
    setWeightMax("");
    setSelectedBrand("");
  };

  const resetFilters = () => {
    setSpecFilters({});
    setPriceMin("");
    setPriceMax("");
    setWeightMin("");
    setWeightMax("");
    setSelectedBrand("");
  };

  const hasActiveFilters =
    priceMin.trim() !== "" ||
    priceMax.trim() !== "" ||
    weightMin.trim() !== "" ||
    weightMax.trim() !== "" ||
    selectedBrand.trim() !== "" ||
    Object.values(specFilters).some((value) => value.trim() !== "");

  const filtersPanel = (
    <>
      <h2 className={styles.sidebarTitle}>Категории</h2>
      <ul className={styles.categoryList}>
        <li>
          <button
            type="button"
            className={`${styles.categoryItem} ${selectedCategory === "all" ? styles.categoryActive : ""}`}
            onClick={() => selectCategory("all")}
          >
            <span>Все товары</span>
          </button>
        </li>
        {categories.map((category) => (
          <li key={category.id}>
            <button
              type="button"
              className={`${styles.categoryItem} ${selectedCategory === category.id ? styles.categoryActive : ""}`}
              onClick={() => selectCategory(category.id)}
            >
              <span>{category.name}</span>
            </button>
          </li>
        ))}
      </ul>

      <div className={styles.filterBlock}>
        <h3>Цена</h3>
        <div className={styles.priceRow}>
          <input
            type="number"
            min="0"
            placeholder="От"
            className={styles.filterInput}
            value={priceMin}
            onChange={(e) => setPriceMin(e.target.value)}
          />
          <input
            type="number"
            min="0"
            placeholder="До"
            className={styles.filterInput}
            value={priceMax}
            onChange={(e) => setPriceMax(e.target.value)}
          />
        </div>
      </div>

      {brandOptions.length > 0 && (
        <div className={styles.filterBlock}>
          <h3>Бренд</h3>
          <select
            className={styles.filterSelect}
            value={selectedBrand}
            onChange={(e) => setSelectedBrand(e.target.value)}
          >
            <option value="">Любой</option>
            {brandOptions.map((brand) => (
              <option key={brand} value={brand}>
                {brand}
              </option>
            ))}
          </select>
        </div>
      )}

      {showWeightFilter && (
        <div className={styles.filterBlock}>
          <h3>Вес (г)</h3>
          <div className={styles.priceRow}>
            <input
              type="number"
              min="0"
              placeholder="От"
              className={styles.filterInput}
              value={weightMin}
              onChange={(e) => setWeightMin(e.target.value)}
            />
            <input
              type="number"
              min="0"
              placeholder="До"
              className={styles.filterInput}
              value={weightMax}
              onChange={(e) => setWeightMax(e.target.value)}
            />
          </div>
        </div>
      )}

      {dynamicFilters.map((field) => (
        <div key={field.key} className={styles.filterBlock}>
          <h3>{field.label}</h3>
          <select
            className={styles.filterSelect}
            value={specFilters[field.key] ?? ""}
            onChange={(e) =>
              setSpecFilters((current) => ({
                ...current,
                [field.key]: e.target.value,
              }))
            }
          >
            <option value="">Любой</option>
            {field.options?.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </div>
      ))}

      {hasActiveFilters && (
        <button
          type="button"
          className={styles.resetFilters}
          onClick={resetFilters}
        >
          Сбросить фильтры
        </button>
      )}
    </>
  );

  return (
    <div className="page">
      <Header />
      <main className={`container ${styles.main}`}>
        <nav className={styles.breadcrumbs}>
          <Link to="/">Главная</Link>
          <span>/</span>
          <span>Каталог</span>
        </nav>

        <div className={styles.layout}>
          <aside className={`${styles.sidebar} ${styles.sidebarDesktop}`}>
            {filtersPanel}
          </aside>

          <FilterSheet
            open={filtersOpen}
            onClose={() => setFiltersOpen(false)}
            footer={
              <button
                type="button"
                className={styles.filterSheetFooterBtn}
                onClick={() => setFiltersOpen(false)}
              >
                Показать {filteredProducts.length} товаров
              </button>
            }
          >
            {filtersPanel}
          </FilterSheet>

          <div className={styles.content}>
            <div className={styles.top}>
              <div>
                <h1 className={styles.title}>
                  {selectedCategory === "all"
                    ? "Каталог"
                    : selectedCategoryName ?? "Каталог"}
                </h1>
                <p className={styles.subtitle}>
                  Найдено товаров: {filteredProducts.length}
                  {filteredProducts.length !== products.length && (
                    <> из {products.length}</>
                  )}
                  {totalPages > 1 && (
                    <>
                      {" "}
                      · страница {currentPage} из {totalPages}
                    </>
                  )}
                  {searchQuery.trim() && (
                    <> · поиск: «{searchQuery.trim()}»</>
                  )}
                </p>
              </div>
              <select
                className={styles.sort}
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
              >
                <option value="popular">Сначала популярные</option>
                <option value="price-asc">Сначала дешевле</option>
                <option value="price-desc">Сначала дороже</option>
              </select>
            </div>

            <div className={styles.mobileToolbar}>
              <button
                type="button"
                className={styles.filterOpenBtn}
                onClick={() => setFiltersOpen(true)}
              >
                <SlidersHorizontal size={18} />
                Фильтры
              </button>
              <select
                className={styles.sort}
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
                aria-label="Сортировка"
              >
                <option value="popular">Сортировка: популярные</option>
                <option value="price-asc">Сортировка: дешевле</option>
                <option value="price-desc">Сортировка: дороже</option>
              </select>
            </div>

            {error && <div className={styles.error}>{error}</div>}

            {filteredProducts.length === 0 && !error ? (
              <p className={styles.empty}>
                По выбранным фильтрам товаров не найдено. Попробуйте изменить
                параметры.
              </p>
            ) : (
              <div className={styles.grid}>
                {paginatedProducts.map((product) => {
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
            )}

            <Pagination
              currentPage={currentPage}
              totalItems={filteredProducts.length}
              pageSize={CATALOG_PAGE_SIZE}
              onPageChange={setCurrentPage}
            />
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
