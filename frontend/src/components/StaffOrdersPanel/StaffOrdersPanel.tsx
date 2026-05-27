import type { Order } from "../../api/productApi";
import {
  formatOrderDate,
  getDeliveryPaymentLabel,
  getDeliveryTypeLabel,
  getItemsSubtotal,
  getOrderStatusLabel,
} from "../../utils/orderDisplay";
import { COURIER_ORDER_STATUS_OPTIONS } from "../../utils/orderStatuses";
import styles from "../../pages/AdminPage/AdminPage.module.css";

interface StaffOrdersPanelProps {
  orders: Order[];
  loadFailed: boolean;
  courierMode?: boolean;
  onStatusChange: (orderId: number, status: string) => void | Promise<void>;
}

export function StaffOrdersPanel({
  orders,
  loadFailed,
  courierMode = false,
  onStatusChange,
}: StaffOrdersPanelProps) {
  if (loadFailed) {
    return (
      <p className={styles.ordersEmpty}>
        Список заказов временно недоступен. Перезапустите product-service и
        выполните миграции: <code>.\scripts\migrate.ps1</code>
      </p>
    );
  }

  if (orders.length === 0) {
    return <p className={styles.ordersEmpty}>Заказов пока нет</p>;
  }

  return (
    <div className={styles.orders}>
      {orders.map((order) => {
        const itemsSubtotal = getItemsSubtotal(order);
        const statusOptions = courierMode
          ? COURIER_ORDER_STATUS_OPTIONS
          : [
              { value: "pending", label: "Ожидает" },
              { value: "confirmed", label: "Подтверждён" },
              { value: "shipped", label: "Отправлен" },
              { value: "delivered", label: "Доставлен" },
              { value: "received", label: "Получен" },
              { value: "cancelled", label: "Отменён" },
            ];

        return (
          <article key={order.id} className={styles.orderCard}>
            <header className={styles.orderCardHead}>
              <div>
                <strong className={styles.orderId}>Заказ #{order.id}</strong>
                <span className={styles.orderDate}>
                  {formatOrderDate(order.created_at)}
                </span>
              </div>
              <select
                className={styles.orderStatusSelect}
                value={order.status}
                title={getOrderStatusLabel(order.status)}
                onChange={(event) =>
                  void onStatusChange(order.id, event.target.value)
                }
              >
                {!statusOptions.some((option) => option.value === order.status) && (
                  <option value={order.status} disabled>
                    {getOrderStatusLabel(order.status)}
                  </option>
                )}
                {statusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </header>

            <div className={styles.orderDetails}>
              <div className={styles.orderDetailBlock}>
                <span className={styles.orderDetailLabel}>Клиент</span>
                <p>{order.contact_name}</p>
                <p>{order.contact_phone}</p>
                <p>{order.contact_email}</p>
              </div>

              <div className={styles.orderDetailBlock}>
                <span className={styles.orderDetailLabel}>Доставка</span>
                <p>
                  <strong>{getDeliveryTypeLabel(order)}</strong>
                </p>
                {order.delivery_type === "cdek" && order.cdek_pvz_code && (
                  <p className={styles.orderHighlight}>
                    ПВЗ: {order.cdek_pvz_code}
                  </p>
                )}
                {order.delivery_city && <p>Город: {order.delivery_city}</p>}
                <p className={styles.orderAddress}>
                  {order.delivery_address || "—"}
                </p>
                {(order.delivery_cost ?? 0) > 0 && (
                  <p>
                    Стоимость доставки:{" "}
                    {order.delivery_cost?.toLocaleString("ru-RU")} ₽ ·{" "}
                    {getDeliveryPaymentLabel(order)}
                  </p>
                )}
              </div>

              <div className={styles.orderDetailBlock}>
                <span className={styles.orderDetailLabel}>Оплата</span>
                <p>Товары: на сайте (карта)</p>
                <p>{getDeliveryPaymentLabel(order)}</p>
                <p className={styles.orderTotal}>
                  Итого: {order.total_amount.toLocaleString("ru-RU")} ₽
                </p>
                <p className={styles.orderSubtotalHint}>
                  Товары: {itemsSubtotal.toLocaleString("ru-RU")} ₽
                  {(order.delivery_cost ?? 0) > 0 &&
                    ` + доставка ${order.delivery_cost?.toLocaleString("ru-RU")} ₽`}
                </p>
              </div>
            </div>

            {(order.items?.length ?? 0) > 0 && (
              <div className={styles.orderItems}>
                <span className={styles.orderDetailLabel}>Состав заказа</span>
                <ul>
                  {order.items.map((item) => (
                    <li key={item.id}>
                      <span>{item.title}</span>
                      <span>
                        {item.quantity} × {item.price.toLocaleString("ru-RU")} ₽
                      </span>
                      <strong>
                        {(item.price * item.quantity).toLocaleString("ru-RU")} ₽
                      </strong>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {order.comment?.trim() && (
              <p className={styles.orderComment}>
                <span className={styles.orderDetailLabel}>Комментарий</span>
                {order.comment}
              </p>
            )}
          </article>
        );
      })}
    </div>
  );
}
