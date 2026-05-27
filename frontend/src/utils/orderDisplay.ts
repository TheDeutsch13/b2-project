import type { Order } from "../api/productApi";

const STATUS_LABELS: Record<string, string> = {
  pending: "Ожидает",
  confirmed: "Подтверждён",
  shipped: "Отправлен",
  delivered: "Доставлен",
  received: "Получен",
  cancelled: "Отменён",
};

/** Заказ в этом статусе — покупатель может оставить отзыв */
export const ORDER_STATUS_FOR_REVIEWS = "received";

export function canLeaveReviewForOrder(status: string): boolean {
  return status === ORDER_STATUS_FOR_REVIEWS;
}

export function getOrderStatusLabel(status: string): string {
  return STATUS_LABELS[status] ?? status;
}

export function formatOrderDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function getDeliveryTypeLabel(order: Order): string {
  if (order.delivery_type === "cdek") {
    return "СДЭК — пункт выдачи";
  }

  if (order.delivery_type === "custom") {
    return "Доставка по адресу";
  }

  return "Доставка";
}

export function getDeliveryPaymentLabel(order: Order): string {
  if (order.delivery_payment === "online") {
    return "Доставка оплачена на сайте";
  }

  if (order.delivery_payment === "on_receipt") {
    return "Доставка при получении";
  }

  return "—";
}

export function getOrderDeliverySummary(order: Order): string {
  const parts: string[] = [];

  if (order.delivery_city) {
    parts.push(order.delivery_city);
  }

  if (order.delivery_type === "cdek" && order.cdek_pvz_code) {
    parts.push(`ПВЗ: ${order.cdek_pvz_code}`);
  }

  if (order.delivery_address) {
    parts.push(order.delivery_address);
  }

  return parts.join(" · ") || "—";
}

export function getItemsSubtotal(order: Order): number {
  return (order.items ?? []).reduce(
    (sum, item) => sum + item.price * item.quantity,
    0
  );
}
