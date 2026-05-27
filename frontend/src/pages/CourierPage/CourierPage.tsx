import { useCallback, useEffect, useState } from "react";
import { productApi, type Order } from "../../api/productApi";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { StaffOrdersPanel } from "../../components/StaffOrdersPanel/StaffOrdersPanel";
import styles from "./CourierPage.module.css";

export function CourierPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [message, setMessage] = useState("");
  const [loadFailed, setLoadFailed] = useState(false);

  const loadOrders = useCallback(async () => {
    try {
      setOrders(await productApi.getCourierOrders());
      setLoadFailed(false);
    } catch {
      setOrders([]);
      setLoadFailed(true);
    }
  }, []);

  useEffect(() => {
    void loadOrders();
  }, [loadOrders]);

  const handleStatusChange = async (orderId: number, status: string) => {
    try {
      await productApi.updateCourierOrderStatus(orderId, status);
      setMessage(`Статус заказа #${orderId} обновлён`);
      await loadOrders();
    } catch {
      setMessage("Не удалось обновить статус");
    }
  };

  return (
    <div className="page">
      <Header />
      <main className={styles.wrapper}>
        <div className="container">
          <div className={styles.head}>
            <h1>Доставка заказов</h1>
            <p>Подтверждайте заказы и отмечайте этапы доставки.</p>
          </div>

          {message && <p className={styles.message}>{message}</p>}

          <StaffOrdersPanel
            orders={orders}
            loadFailed={loadFailed}
            courierMode
            onStatusChange={handleStatusChange}
          />
        </div>
      </main>
      <Footer />
    </div>
  );
}
