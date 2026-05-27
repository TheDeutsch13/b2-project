import { Minus, Plus, X } from "lucide-react";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import {
  closeCart,
  removeFromCart,
  setQuantity,
} from "../../features/cart/cartSlice";
import styles from "./CartDrawer.module.css";

export function CartDrawer() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { items, isOpen } = useAppSelector((state) => state.cart);
  const user = useAppSelector((state) => state.auth.user);

  const total = items.reduce(
    (sum, item) => sum + item.price * item.quantity,
    0
  );
  const itemsCount = items.reduce((sum, item) => sum + item.quantity, 0);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        dispatch(closeCart());
      }
    };

    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKeyDown);

    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [dispatch, isOpen]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className={styles.root}>
      <button
        type="button"
        className={styles.backdrop}
        aria-label="Закрыть корзину"
        onClick={() => dispatch(closeCart())}
      />

      <aside className={styles.drawer} role="dialog" aria-label="Корзина">
        <header className={styles.header}>
          <div>
            <h2>Корзина</h2>
            <p>
              {itemsCount}{" "}
              {itemsCount === 1
                ? "товар"
                : itemsCount >= 2 && itemsCount <= 4
                  ? "товара"
                  : "товаров"}
            </p>
          </div>
          <button
            type="button"
            className={styles.closeBtn}
            aria-label="Закрыть"
            onClick={() => dispatch(closeCart())}
          >
            <X size={18} />
          </button>
        </header>

        <div className={styles.list}>
          {items.length === 0 ? (
            <p className={styles.empty}>Корзина пуста</p>
          ) : (
            items.map((item) => (
              <article key={item.productId} className={styles.item}>
                <div className={styles.itemImage}>
                  {item.imageSrc ? (
                    <img src={item.imageSrc} alt="" />
                  ) : (
                    <div className={styles.itemPlaceholder} />
                  )}
                </div>
                <div className={styles.itemBody}>
                  <div className={styles.itemTop}>
                    <h3>{item.title}</h3>
                    <strong>
                      {(item.price * item.quantity).toLocaleString("ru-RU")} ₽
                    </strong>
                  </div>
                  <div className={styles.qtyRow}>
                    <button
                      type="button"
                      aria-label="Уменьшить количество"
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
                      aria-label="Увеличить количество"
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
                    <button
                      type="button"
                      className={styles.removeBtn}
                      onClick={() => dispatch(removeFromCart(item.productId))}
                    >
                      Удалить
                    </button>
                  </div>
                </div>
              </article>
            ))
          )}
        </div>

        <footer className={styles.footer}>
          <div className={styles.totalRow}>
            <span>Итого:</span>
            <strong>{total.toLocaleString("ru-RU")} ₽</strong>
          </div>
          <button
            type="button"
            className={styles.checkoutBtn}
            disabled={items.length === 0}
            onClick={() => {
              dispatch(closeCart());
              navigate("/cart");
            }}
          >
            В корзину
          </button>
          {!user && items.length > 0 && (
            <p className={styles.guestHint}>
              Для оплаты{" "}
              <Link
                to="/login"
                state={{ from: { pathname: "/checkout" } }}
                onClick={() => dispatch(closeCart())}
              >
                войдите
              </Link>{" "}
              в аккаунт
            </p>
          )}
          {user && items.length > 0 && (
            <button
              type="button"
              className={styles.checkoutSecondaryBtn}
              onClick={() => {
                dispatch(closeCart());
                navigate("/checkout");
              }}
            >
              Оформить заказ
            </button>
          )}
        </footer>
      </aside>
    </div>
  );
}
