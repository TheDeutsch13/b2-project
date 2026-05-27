import { useEffect, useMemo, useState } from "react";
import { Eye, Pencil, Search, SlidersHorizontal, Trash2 } from "lucide-react";
import { FilterSheet } from "../../components/FilterSheet/FilterSheet";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAppSelector } from "../../app/hooks";
import { authApi, type UserProfile } from "../../api/authApi";
import {
  ASSIGNABLE_ROLES,
  getRoleLabel,
  isUserRole,
} from "../../constants/roles";
import { StaffOrdersPanel } from "../../components/StaffOrdersPanel/StaffOrdersPanel";
import { AdminReviewsPanel } from "../../components/AdminReviewsPanel/AdminReviewsPanel";
import { SupportStaffPanel } from "../../components/SupportStaffPanel/SupportStaffPanel";
import {
  productApi,
  type Category,
  type Order,
  type Product,
} from "../../api/productApi";
import {
  AdminSortableTh,
  getDefaultSortDirection,
  type AdminSortColumn,
  type SortDirection,
} from "../../components/AdminSortableTh/AdminSortableTh";
import {
  getProfileDisplayName,
  getProfileFullName,
  userAvatarUrl,
} from "../../utils/userDisplay";
import { AccountAvatar } from "../../components/AccountAvatar/AccountAvatar";
import { AdminProductModal } from "../../components/AdminProductModal/AdminProductModal";
import { AdminSidebar } from "../../components/AdminSidebar/AdminSidebar";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { sortCategoriesByFixedOrder } from "../../data/categories";
import { ProductNoImage } from "../../components/ProductNoImage/ProductNoImage";
import {
  getProductPrimaryImage,
  isProductInStock,
} from "../../utils/productDisplay";
import { formatAdminProductRating } from "../../utils/productRating";
import styles from "./AdminPage.module.css";

const MODERATOR_TABS = new Set(["users", "support"]);

export function AdminPage() {
  const navigate = useNavigate();
  const currentUser = useAppSelector((state) => state.auth.user);
  const isAdmin = currentUser?.role === "admin";
  const isModerator = currentUser?.role === "moderator";

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [message, setMessage] = useState("");
  const [ordersLoadFailed, setOrdersLoadFailed] = useState(false);
  const [search, setSearch] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [brandFilter, setBrandFilter] = useState<string>("all");
  const [stockFilter, setStockFilter] = useState<string>("all");
  const [sortColumn, setSortColumn] = useState<AdminSortColumn | "id">("id");
  const [sortDirection, setSortDirection] = useState<SortDirection>("desc");
  const [userSearch, setUserSearch] = useState("");
  const [userRoleFilter, setUserRoleFilter] = useState<string>("all");
  const [userSortColumn, setUserSortColumn] = useState<AdminSortColumn | "id">(
    "created"
  );
  const [userSortDirection, setUserSortDirection] =
    useState<SortDirection>("desc");
  const [modalOpen, setModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [productFiltersOpen, setProductFiltersOpen] = useState(false);
  const [reviewsCount, setReviewsCount] = useState(0);

  const [searchParams] = useSearchParams();
  const tab = searchParams.get("tab") ?? "products";
  const isProductsTab = tab === "products";
  const isOrdersTab = tab === "orders";
  const isUsersTab = tab === "users";
  const isReviewsTab = tab === "reviews";
  const isSupportTab = tab === "support";

  const loadData = async () => {
    const errors: string[] = [];

    if (isAdmin) {
      try {
        setProducts(await productApi.getProducts());
      } catch {
        setProducts([]);
        errors.push("товары");
      }

      try {
        const categoriesData = sortCategoriesByFixedOrder(
          await productApi.getCategories()
        );
        setCategories(categoriesData);
      } catch {
        setCategories([]);
        errors.push("категории");
      }

      try {
        setOrders(await productApi.getAllOrders());
        setOrdersLoadFailed(false);
      } catch {
        setOrders([]);
        setOrdersLoadFailed(true);
      }
    }

    if (isAdmin || isModerator) {
      try {
        setUsers(await authApi.getUsers());
      } catch {
        setUsers([]);
        errors.push("пользователи");
      }
    }

    if (errors.length > 0) {
      const needsAuthService = errors.includes("пользователи");
      const needsProductService = errors.some((item) =>
        ["товары", "категории", "заказы"].includes(item)
      );

      const hint = needsAuthService && !needsProductService
        ? "Проверьте auth-service."
        : needsProductService && !needsAuthService
          ? "Проверьте product-service."
          : "Проверьте auth-service и product-service.";

      setMessage(
        `Не удалось загрузить: ${errors.join(", ")}. ${hint}`
      );
    }
  };

  useEffect(() => {
    void loadData();
  }, [isAdmin, isModerator]);

  useEffect(() => {
    if (isModerator && !MODERATOR_TABS.has(tab)) {
      navigate("/admin?tab=users", { replace: true });
    }
  }, [isModerator, tab, navigate]);

  const brandOptions = useMemo(
    () =>
      [...new Set(products.map((product) => product.brand).filter(Boolean))].sort(
        (a, b) => a.localeCompare(b, "ru")
      ),
    [products]
  );

  const filteredProducts = useMemo(() => {
    const query = search.trim().toLowerCase();

    let list = products.filter((product) => {
      if (query) {
        const matchesSearch =
          product.title.toLowerCase().includes(query) ||
          product.brand?.toLowerCase().includes(query) ||
          product.category_name?.toLowerCase().includes(query);

        if (!matchesSearch) {
          return false;
        }
      }

      if (
        categoryFilter !== "all" &&
        String(product.category_id ?? "") !== categoryFilter
      ) {
        return false;
      }

      if (brandFilter !== "all" && product.brand !== brandFilter) {
        return false;
      }

      const inStock = isProductInStock(product);
      if (stockFilter === "in_stock" && !inStock) {
        return false;
      }
      if (stockFilter === "out_of_stock" && inStock) {
        return false;
      }

      return true;
    });

    const directionMultiplier = sortDirection === "asc" ? 1 : -1;

    list = [...list].sort((a, b) => {
      let result = 0;

      switch (sortColumn) {
        case "title":
          result = a.title.localeCompare(b.title, "ru");
          break;
        case "category":
          result = (a.category_name ?? "").localeCompare(
            b.category_name ?? "",
            "ru"
          );
          break;
        case "brand":
          result = (a.brand ?? "").localeCompare(b.brand ?? "", "ru");
          break;
        case "price":
          result = a.price - b.price;
          break;
        case "rating":
          result =
            (a.rating_avg ?? 0) - (b.rating_avg ?? 0) ||
            (a.rating_count ?? 0) - (b.rating_count ?? 0);
          break;
        case "status": {
          const aInStock = isProductInStock(a) ? 1 : 0;
          const bInStock = isProductInStock(b) ? 1 : 0;
          result = aInStock - bInStock;
          break;
        }
        case "stock":
          result = (a.stock ?? 0) - (b.stock ?? 0);
          break;
        default:
          result = a.id - b.id;
      }

      if (result !== 0) {
        return result * directionMultiplier;
      }

      return b.id - a.id;
    });

    return list;
  }, [
    products,
    search,
    categoryFilter,
    brandFilter,
    stockFilter,
    sortColumn,
    sortDirection,
  ]);

  const handleSort = (column: AdminSortColumn) => {
    if (sortColumn === column) {
      setSortDirection((current) => (current === "asc" ? "desc" : "asc"));
      return;
    }

    setSortColumn(column);
    setSortDirection(getDefaultSortDirection(column));
  };

  const activeSortColumn: AdminSortColumn | null =
    sortColumn === "id" ? null : sortColumn;

  const filteredUsers = useMemo(() => {
    const query = userSearch.trim().toLowerCase();

    let list = users.filter((item) => {
      if (userRoleFilter !== "all" && item.role !== userRoleFilter) {
        return false;
      }

      if (!query) {
        return true;
      }

      const fullName = getProfileFullName({
        firstName: item.first_name,
        lastName: item.last_name,
      }).toLowerCase();

      return (
        item.email.toLowerCase().includes(query) ||
        fullName.includes(query) ||
        item.phone.toLowerCase().includes(query) ||
        item.nickname.toLowerCase().includes(query)
      );
    });

    const directionMultiplier = userSortDirection === "asc" ? 1 : -1;

    list = [...list].sort((a, b) => {
      let result = 0;

      switch (userSortColumn) {
        case "name":
          result = getProfileFullName({
            firstName: a.first_name,
            lastName: a.last_name,
          }).localeCompare(
            getProfileFullName({
              firstName: b.first_name,
              lastName: b.last_name,
            }),
            "ru"
          );
          break;
        case "email":
          result = a.email.localeCompare(b.email, "ru");
          break;
        case "phone":
          result = a.phone.localeCompare(b.phone, "ru");
          break;
        case "role":
          result = a.role.localeCompare(b.role, "ru");
          break;
        case "created":
          result =
            new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
        default:
          result = a.id - b.id;
      }

      if (result !== 0) {
        return result * directionMultiplier;
      }

      return b.id - a.id;
    });

    return list;
  }, [
    users,
    userSearch,
    userRoleFilter,
    userSortColumn,
    userSortDirection,
  ]);

  const activeUserSortColumn: AdminSortColumn | null =
    userSortColumn === "id" ? null : userSortColumn;

  const handleUserSort = (column: AdminSortColumn) => {
    if (userSortColumn === column) {
      setUserSortDirection((current) => (current === "asc" ? "desc" : "asc"));
      return;
    }

    setUserSortColumn(column);
    setUserSortDirection(getDefaultSortDirection(column));
  };

  const inStockCount = products.filter((product) => isProductInStock(product)).length;
  const brandsCount = new Set(products.map((product) => product.brand).filter(Boolean))
    .size;

  const openCreate = () => {
    setEditingProduct(null);
    setModalOpen(true);
  };

  const openEdit = (product: Product) => {
    setEditingProduct(product);
    setModalOpen(true);
  };

  const handleDelete = async (product: Product) => {
    const confirmed = window.confirm(
      `Удалить товар «${product.title}»? Это действие нельзя отменить.`
    );

    if (!confirmed) {
      return;
    }

    try {
      await productApi.deleteProduct(product.id);
      setMessage(`Товар «${product.title}» удалён`);
      await loadData();
    } catch (err) {
      const apiMessage =
        typeof err === "object" &&
        err !== null &&
        "response" in err &&
        typeof (err as { response?: { data?: { error?: string } } }).response
          ?.data?.error === "string"
          ? (err as { response: { data: { error: string } } }).response.data.error
          : null;

      if (apiMessage === "product not found" || apiMessage?.includes("404")) {
        setMessage("Не удалось удалить товар. Перезапустите product-service (нужна новая версия API).");
      } else {
        setMessage(apiMessage ?? "Не удалось удалить товар");
      }
    }
  };

  const handleStatusChange = async (orderId: number, status: string) => {
    try {
      await productApi.updateOrderStatus(orderId, status);
      setMessage(`Статус заказа #${orderId} обновлён`);
      await loadData();
    } catch {
      setMessage("Не удалось обновить статус");
    }
  };

  const handleRoleChange = async (userId: number, role: string) => {
    if (!isUserRole(role)) {
      return;
    }

    try {
      await authApi.updateUserRole(userId, role);
      setMessage("Роль обновлена. Пользователю нужно войти заново.");
      await loadData();
    } catch {
      setMessage("Не удалось изменить роль");
    }
  };

  return (
    <div className="page">
      <Header />
      <main className={styles.wrapper}>
        <div className={`container ${styles.layout}`}>
          <AdminSidebar />

          <div className={styles.content}>
            <div className={styles.pageHead}>
              <h1>
                {isOrdersTab
                  ? "Заказы"
                  : isUsersTab
                    ? "Пользователи"
                    : isSupportTab
                      ? "Поддержка"
                      : isReviewsTab
                        ? "Отзывы"
                        : "Товары"}
                {isProductsTab && (
                  <span className={styles.badge}>{products.length}</span>
                )}
                {isOrdersTab && (
                  <span className={styles.badge}>{orders.length}</span>
                )}
                {isUsersTab && (
                  <span className={styles.badge}>{users.length}</span>
                )}
                {isReviewsTab && (
                  <span className={styles.badge}>{reviewsCount}</span>
                )}
              </h1>
              {isProductsTab && (
                <button
                  type="button"
                  className={styles.addBtn}
                  onClick={openCreate}
                >
                  Добавить товар +
                </button>
              )}
            </div>

            {isProductsTab && (
              <>
                <div className={styles.filtersRow}>
                  <button
                    type="button"
                    className={styles.filterOpenBtn}
                    onClick={() => setProductFiltersOpen(true)}
                  >
                    <SlidersHorizontal size={16} />
                    Фильтры
                  </button>

                  <div className={styles.filtersDesktop}>
                    <select
                      className={styles.filterSelect}
                      value={categoryFilter}
                      onChange={(e) => setCategoryFilter(e.target.value)}
                    >
                      <option value="all">Все категории</option>
                      {categories.map((category) => (
                        <option key={category.id} value={String(category.id)}>
                          {category.name}
                        </option>
                      ))}
                    </select>

                    <select
                      className={styles.filterSelect}
                      value={brandFilter}
                      onChange={(e) => setBrandFilter(e.target.value)}
                    >
                      <option value="all">Все бренды</option>
                      {brandOptions.map((brand) => (
                        <option key={brand} value={brand}>
                          {brand}
                        </option>
                      ))}
                    </select>

                    <select
                      className={styles.filterSelect}
                      value={stockFilter}
                      onChange={(e) => setStockFilter(e.target.value)}
                    >
                      <option value="all">Любой остаток</option>
                      <option value="in_stock">В наличии</option>
                      <option value="out_of_stock">Нет в наличии</option>
                    </select>
                  </div>

                  <div className={styles.search}>
                    <Search size={16} />
                    <input
                      placeholder="Поиск товаров..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                    />
                  </div>
                </div>

                <FilterSheet
                  open={productFiltersOpen}
                  onClose={() => setProductFiltersOpen(false)}
                  footer={
                    <button
                      type="button"
                      className={styles.resetFiltersBtn}
                      onClick={() => setProductFiltersOpen(false)}
                    >
                      Применить
                    </button>
                  }
                >
                  <div className={styles.sheetFilters}>
                    <label className={styles.sheetLabel}>
                      Категория
                      <select
                        className={styles.filterSelect}
                        value={categoryFilter}
                        onChange={(e) => setCategoryFilter(e.target.value)}
                      >
                        <option value="all">Все категории</option>
                        {categories.map((category) => (
                          <option key={category.id} value={String(category.id)}>
                            {category.name}
                          </option>
                        ))}
                      </select>
                    </label>

                    <label className={styles.sheetLabel}>
                      Бренд
                      <select
                        className={styles.filterSelect}
                        value={brandFilter}
                        onChange={(e) => setBrandFilter(e.target.value)}
                      >
                        <option value="all">Все бренды</option>
                        {brandOptions.map((brand) => (
                          <option key={brand} value={brand}>
                            {brand}
                          </option>
                        ))}
                      </select>
                    </label>

                    <label className={styles.sheetLabel}>
                      Остаток
                      <select
                        className={styles.filterSelect}
                        value={stockFilter}
                        onChange={(e) => setStockFilter(e.target.value)}
                      >
                        <option value="all">Любой остаток</option>
                        <option value="in_stock">В наличии</option>
                        <option value="out_of_stock">Нет в наличии</option>
                      </select>
                    </label>
                  </div>
                </FilterSheet>
              </>
            )}

            {message && <div className={styles.message}>{message}</div>}

            {isProductsTab && (
              <>
                <div className={styles.stats}>
                  <div className={styles.statCard}>
                    <span>Всего товаров</span>
                    <strong>{products.length}</strong>
                  </div>
                  <div className={styles.statCard}>
                    <span>В наличии</span>
                    <strong>{inStockCount}</strong>
                  </div>
                  <div className={styles.statCard}>
                    <span>Нет в наличии</span>
                    <strong>{products.length - inStockCount}</strong>
                  </div>
                  <div className={styles.statCard}>
                    <span>Категории</span>
                    <strong>{categories.length}</strong>
                  </div>
                  <div className={styles.statCard}>
                    <span>Бренды</span>
                    <strong>{brandsCount}</strong>
                  </div>
                </div>

                <div className={styles.tableWrap}>
                  <table className={styles.table}>
                    <thead>
                      <tr>
                        <AdminSortableTh
                          label="Товар"
                          column="title"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Категория"
                          column="category"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Бренд"
                          column="brand"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Цена"
                          column="price"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Рейтинг"
                          column="rating"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Статус"
                          column="status"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <AdminSortableTh
                          label="Остаток"
                          column="stock"
                          activeColumn={activeSortColumn}
                          direction={sortDirection}
                          onSort={handleSort}
                        />
                        <th>Действия</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredProducts.length === 0 ? (
                        <tr>
                          <td colSpan={8} className={styles.emptyCell}>
                            Товаров пока нет
                          </td>
                        </tr>
                      ) : (
                        filteredProducts.map((product) => {
                          const inStock = isProductInStock(product);
                          const imageSrc = getProductPrimaryImage(product);

                          return (
                            <tr key={product.id}>
                              <td>
                                <div className={styles.productCell}>
                                  <div className={styles.productThumb}>
                                    {imageSrc ? (
                                      <img src={imageSrc} alt="" />
                                    ) : (
                                      <ProductNoImage compact />
                                    )}
                                  </div>
                                  <span>{product.title}</span>
                                </div>
                              </td>
                              <td>{product.category_name || "—"}</td>
                              <td>{product.brand || "—"}</td>
                              <td>
                                {product.price.toLocaleString("ru-RU")} ₽
                              </td>
                              <td className={styles.ratingCell}>
                                {formatAdminProductRating(
                                  product.rating_avg ?? 0,
                                  product.rating_count ?? 0
                                )}
                              </td>
                              <td>
                                <span
                                  className={
                                    inStock ? styles.inStock : styles.outStock
                                  }
                                >
                                  {inStock ? "В наличии" : "Нет в наличии"}
                                </span>
                              </td>
                              <td>{product.stock ?? 0}</td>
                              <td>
                                <div className={styles.rowActions}>
                                  <Link
                                    to={`/product/${product.id}`}
                                    className={styles.iconBtn}
                                    title="Просмотр"
                                  >
                                    <Eye size={16} />
                                  </Link>
                                  <button
                                    type="button"
                                    className={styles.iconBtn}
                                    title="Редактировать"
                                    onClick={() => openEdit(product)}
                                  >
                                    <Pencil size={16} />
                                  </button>
                                  <button
                                    type="button"
                                    className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                                    title="Удалить"
                                    onClick={() => void handleDelete(product)}
                                  >
                                    <Trash2 size={16} />
                                  </button>
                                </div>
                              </td>
                            </tr>
                          );
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              </>
            )}

            {isOrdersTab && (
              <StaffOrdersPanel
                orders={orders}
                loadFailed={ordersLoadFailed}
                onStatusChange={handleStatusChange}
              />
            )}

            {isSupportTab && <SupportStaffPanel />}

            {isUsersTab && (
              <>
                <div className={styles.filtersRow}>
                  <select
                    className={styles.filterSelect}
                    value={userRoleFilter}
                    onChange={(event) => setUserRoleFilter(event.target.value)}
                  >
                    <option value="all">Все роли</option>
                    {ASSIGNABLE_ROLES.map((role) => (
                      <option key={role} value={role}>
                        {getRoleLabel(role)}
                      </option>
                    ))}
                  </select>

                  <div className={styles.search}>
                    <Search size={16} />
                    <input
                      placeholder="Поиск по имени, email, телефону..."
                      value={userSearch}
                      onChange={(event) => setUserSearch(event.target.value)}
                    />
                  </div>
                </div>

                <div className={styles.tableWrap}>
                  <table className={styles.table}>
                    <thead>
                      <tr>
                        <AdminSortableTh
                          label="Пользователь"
                          column="name"
                          activeColumn={activeUserSortColumn}
                          direction={userSortDirection}
                          onSort={handleUserSort}
                        />
                        <AdminSortableTh
                          label="Email"
                          column="email"
                          activeColumn={activeUserSortColumn}
                          direction={userSortDirection}
                          onSort={handleUserSort}
                        />
                        <AdminSortableTh
                          label="Телефон"
                          column="phone"
                          activeColumn={activeUserSortColumn}
                          direction={userSortDirection}
                          onSort={handleUserSort}
                        />
                        <AdminSortableTh
                          label="Роль"
                          column="role"
                          activeColumn={activeUserSortColumn}
                          direction={userSortDirection}
                          onSort={handleUserSort}
                        />
                        <AdminSortableTh
                          label="Регистрация"
                          column="created"
                          activeColumn={activeUserSortColumn}
                          direction={userSortDirection}
                          onSort={handleUserSort}
                        />
                      </tr>
                    </thead>
                    <tbody>
                      {filteredUsers.length === 0 ? (
                        <tr>
                          <td colSpan={5} className={styles.emptyCell}>
                            Пользователи не найдены
                          </td>
                        </tr>
                      ) : (
                        filteredUsers.map((item) => {
                          const displayName = getProfileDisplayName(
                            {
                              id: item.id,
                              email: item.email,
                              role: item.role,
                            },
                            {
                              firstName: item.first_name,
                              lastName: item.last_name,
                              nickname: item.nickname,
                            }
                          );

                          return (
                            <tr key={item.id}>
                              <td>
                                <div className={styles.productCell}>
                                  <AccountAvatar
                                    name={displayName}
                                    avatarUrl={userAvatarUrl(item.avatar_url)}
                                    size="xs"
                                  />
                                  <span>{displayName}</span>
                                </div>
                              </td>
                              <td>{item.email}</td>
                              <td>{item.phone || "—"}</td>
                              <td>
                                {isAdmin && item.id !== currentUser?.id ? (
                                  <select
                                    className={styles.filterSelect}
                                    value={item.role}
                                    onChange={(event) =>
                                      void handleRoleChange(
                                        item.id,
                                        event.target.value
                                      )
                                    }
                                  >
                                    {ASSIGNABLE_ROLES.map((role) => (
                                      <option key={role} value={role}>
                                        {getRoleLabel(role)}
                                      </option>
                                    ))}
                                  </select>
                                ) : (
                                  getRoleLabel(
                                    isUserRole(item.role) ? item.role : "user"
                                  )
                                )}
                              </td>
                              <td>
                                {new Date(item.created_at).toLocaleDateString(
                                  "ru-RU"
                                )}
                              </td>
                            </tr>
                          );
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              </>
            )}

            {isReviewsTab && (
              <AdminReviewsPanel
                products={products}
                onCountChange={setReviewsCount}
              />
            )}
          </div>
        </div>
      </main>

      <AdminProductModal
        open={modalOpen}
        product={editingProduct}
        categories={categories}
        existingProducts={products}
        onClose={() => setModalOpen(false)}
        onSaved={async () => {
          await loadData();
          setMessage(
            editingProduct ? "Товар обновлён" : "Товар создан"
          );
        }}
      />

      <Footer />
    </div>
  );
}
