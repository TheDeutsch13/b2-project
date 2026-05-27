import { useEffect, useState } from "react";
import { ShoppingBag } from "lucide-react";
import { productApi, type Order } from "../../api/productApi";
import styles from "./ProfileOrdersPage.module.css";

const statusLabels: Record<string, { label: string; className: string }> = {
  pending: { label: "В пути", className: styles.statusTransit },
  confirmed: { label: "В пути", className: styles.statusTransit },
  shipped: { label: "В пути", className: styles.statusTransit },
  delivered: { label: "Доставлен", className: styles.statusDelivered },
  received: { label: "Получен", className: styles.statusReceived },
  cancelled: { label: "Отменен", className: styles.statusCancelled },
};

export function ProfileOrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    productApi
      .getMyOrders()
      .then(setOrders)
      .catch(() => setError("Не удалось загрузить заказы"));
  }, []);

  return (
    <section>
      <h1 className={styles.title}>Мои покупки</h1>
      <p className={styles.subtitle}>
        Показано {orders.length > 0 ? `1-${orders.length}` : "0"} из{" "}
        {orders.length} заказов
      </p>

      {error && <div className={styles.error}>{error}</div>}

      {orders.length === 0 && !error ? (
        <p className={styles.empty}>У вас пока нет заказов</p>
      ) : (
        <div className={styles.list}>
          {orders.map((order) => {
            const status = statusLabels[order.status] ?? statusLabels.pending;

            return (
              <article key={order.id} className={styles.orderCard}>
                <div className={styles.orderHead}>
                  <div>
                    <span className={styles.orderDate}>
                      Заказ от{" "}
                      {new Date(order.created_at).toLocaleString("ru-RU", {
                        day: "2-digit",
                        month: "2-digit",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                    <strong className={styles.orderId}>№ {order.id}</strong>
                  </div>
                  <span className={status.className}>{status.label}</span>
                  <span className={styles.qty}>
                    Кол-во товаров: {order.items?.length ?? 0}
                  </span>
                </div>

                <p className={styles.deliveryLine}>
                  {order.delivery_type === "cdek"
                    ? `СДЭК${order.cdek_pvz_code ? ` · ${order.cdek_pvz_code}` : ""}`
                    : "Доставка по адресу"}
                  {order.delivery_cost
                    ? ` · ${order.delivery_cost.toLocaleString("ru-RU")} ₽ (${
                        order.delivery_payment === "online"
                          ? "оплачена на сайте"
                          : "при получении"
                      })`
                    : ""}
                  {order.delivery_address
                    ? ` — ${order.delivery_address}`
                    : ""}
                </p>

                {order.items?.map((item) => (
                  <div key={item.id} className={styles.orderRow}>
                    <div className={styles.orderIcon}>
                      <ShoppingBag size={22} />
                    </div>
                    <div className={styles.orderInfo}>
                      <strong>{item.title}</strong>
                      <p>
                        {item.quantity} шт. ·{" "}
                        {(item.price * item.quantity).toLocaleString("ru-RU")} ₽
                      </p>
                    </div>
                  </div>
                ))}

                <div className={styles.orderTotal}>
                  Итого: {order.total_amount.toLocaleString("ru-RU")} ₽
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
