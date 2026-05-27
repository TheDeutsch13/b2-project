import { Minus, Plus, Trash2 } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import {
  clearCart,
  removeFromCart,
  setQuantity,
} from "../../features/cart/cartSlice";
import styles from "./CartPage.module.css";

export function CartPage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const user = useAppSelector((state) => state.auth.user);
  const items = useAppSelector((state) => state.cart.items);

  const total = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
  const itemsCount = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <div className="page">
      <Header />
      <main className={`container ${styles.main}`}>
        <h1 className={styles.title}>Корзина</h1>

        <div className={styles.layout}>
          <section className={styles.itemsCard}>
            <header className={styles.itemsHead}>
              <h2>Товары в корзине ({itemsCount})</h2>
              {items.length > 0 && (
                <button
                  type="button"
                  className={styles.clearBtn}
                  onClick={() => dispatch(clearCart())}
                >
                  <Trash2 size={14} />
                  Очистить корзину
                </button>
              )}
            </header>

            {items.length === 0 ? (
              <p className={styles.empty}>
                Корзина пуста.{" "}
                <Link to="/catalog">Перейти в каталог</Link>
              </p>
            ) : (
              <ul className={styles.list}>
                {items.map((item) => (
                  <li key={item.productId}>
                    <div className={styles.thumb}>
                      {item.imageSrc ? (
                        <img src={item.imageSrc} alt="" />
                      ) : (
                        <div className={styles.thumbPlaceholder} />
                      )}
                    </div>
                    <div className={styles.itemBody}>
                      <strong>{item.title}</strong>
                      <div className={styles.qtyRow}>
                        <button
                          type="button"
                          onClick={() =>
                            dispatch(
                              setQuantity({
                                productId: item.productId,
                                quantity: item.quantity - 1,
                              })
                            )
                          }
                        >
                          <Minus size={14} />
                        </button>
                        <span>{item.quantity}</span>
                        <button
                          type="button"
                          onClick={() =>
                            dispatch(
                              setQuantity({
                                productId: item.productId,
                                quantity: item.quantity + 1,
                              })
                            )
                          }
                        >
                          <Plus size={14} />
                        </button>
                      </div>
                    </div>
                    <div className={styles.itemPrice}>
                      <strong>
                        {(item.price * item.quantity).toLocaleString("ru-RU")} ₽
                      </strong>
                      <button
                        type="button"
                        className={styles.removeBtn}
                        onClick={() => dispatch(removeFromCart(item.productId))}
                      >
                        ×
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <aside className={styles.summary}>
            <h2>Сводка заказа</h2>
            <p>{itemsCount} товаров</p>
            <div className={styles.totalRow}>
              <span>Итого:</span>
              <strong>{total.toLocaleString("ru-RU")} ₽</strong>
            </div>
            <button
              type="button"
              className={styles.checkoutBtn}
              disabled={items.length === 0}
              onClick={() => {
                if (user) {
                  navigate("/checkout");
                  return;
                }
                navigate("/login", { state: { from: { pathname: "/checkout" } } });
              }}
            >
              {user ? "Оформить заказ" : "Войти для оформления"}
            </button>
            {!user && items.length > 0 && (
              <p className={styles.guestHint}>
                Оплата доступна только после{" "}
                <Link
                  to="/login"
                  state={{ from: { pathname: "/checkout" } }}
                >
                  входа
                </Link>{" "}
                или{" "}
                <Link to="/register">регистрации</Link>.
              </p>
            )}
          </aside>
        </div>
      </main>
      <Footer />
    </div>
  );
}
