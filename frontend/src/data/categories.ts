/** Фиксированный набор категорий магазина (совпадает с seed-миграцией product). */
export const FIXED_CATEGORIES = [
  "Мыши",
  "Коврики",
  "Клавиатуры",
  "Аксессуары",
] as const;

export type FixedCategoryName = (typeof FIXED_CATEGORIES)[number];

const fixedCategorySet = new Set<string>(FIXED_CATEGORIES);

export function filterFixedCategories<T extends { name: string }>(items: T[]): T[] {
  return items.filter((item) => fixedCategorySet.has(item.name));
}

export function sortCategoriesByFixedOrder<T extends { name: string }>(
  items: T[]
): T[] {
  const order = new Map(
    FIXED_CATEGORIES.map((name, index) => [name, index])
  );

  return filterFixedCategories(items).sort((a, b) => {
    const left = order.get(a.name as FixedCategoryName) ?? 999;
    const right = order.get(b.name as FixedCategoryName) ?? 999;
    return left - right;
  });
}
