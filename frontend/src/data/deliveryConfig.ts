export type DeliveryType = "cdek" | "custom";

export const DELIVERY_COSTS: Record<DeliveryType, number> = {
  cdek: 305,
  custom: 700,
};

export const DELIVERY_LABELS: Record<DeliveryType, string> = {
  cdek: "СДЭК — пункт выдачи",
  custom: "Доставка по адресу",
};
