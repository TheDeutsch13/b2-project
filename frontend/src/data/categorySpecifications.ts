import type { ProductSpecification } from "../api/productApi";
import type { FixedCategoryName } from "./categories";
import { FIXED_CATEGORIES } from "./categories";

export type SpecFieldType = "text" | "number";
export type SpecFilterMode = "select" | "range";

/** Ключ характеристики веса (фильтр «от — до» в граммах) */
export const WEIGHT_SPEC_KEY = "Вес";

export interface CategorySpecField {
  /** Ключ в БД (поле label у характеристики товара) */
  key: string;
  label: string;
  type: SpecFieldType;
  /** Участвует в фильтрах каталога (значения берутся из товаров) */
  filterable?: boolean;
  /** select — выпадающий список; range — поля «от» и «до» */
  filterMode?: SpecFilterMode;
  /** Фиксированный список в админке (например подключение) */
  options?: string[];
  placeholder?: string;
  unit?: string;
}

export const CONNECTION_OPTIONS = ["Проводная", "Беспроводная"] as const;

export const CATEGORY_SPEC_TEMPLATES: Record<
  FixedCategoryName,
  CategorySpecField[]
> = {
  Мыши: [
    {
      key: "Сенсор",
      label: "Сенсор",
      type: "text",
      filterable: true,
      placeholder: "например HERO 2",
    },
    {
      key: WEIGHT_SPEC_KEY,
      label: "Вес",
      type: "text",
      filterable: true,
      filterMode: "range",
      placeholder: "например 63",
      unit: "г",
    },
    {
      key: "Подключение",
      label: "Подключение",
      type: "text",
      filterable: true,
      options: [...CONNECTION_OPTIONS],
    },
    {
      key: "Частота опроса",
      label: "Частота опроса",
      type: "text",
      filterable: true,
      placeholder: "например 8000",
      unit: "Hz",
    },
    {
      key: "Материал",
      label: "Материал",
      type: "text",
      filterable: true,
      placeholder: "например пластик",
    },
    {
      key: "Энкодер",
      label: "Энкодер",
      type: "text",
      filterable: true,
      placeholder: "например TTC Gold",
    },
    {
      key: "Переключатели",
      label: "Переключатели",
      type: "text",
      filterable: true,
      placeholder: "например Omron",
    },
  ],
  Коврики: [
    {
      key: "Размер",
      label: "Размер",
      type: "text",
      filterable: true,
      placeholder: "490x420",
      unit: "мм",
    },
    {
      key: "Толщина",
      label: "Толщина",
      type: "text",
      filterable: true,
      placeholder: "например 4 мм",
    },
    {
      key: "Поверхность",
      label: "Поверхность",
      type: "text",
      filterable: true,
      placeholder: "Speed, Control…",
    },
    {
      key: "Материал",
      label: "Материал",
      type: "text",
      filterable: true,
      placeholder: "Ткань, стекло…",
    },
  ],
  Клавиатуры: [
    {
      key: "Переключатели",
      label: "Переключатели",
      type: "text",
      filterable: true,
      placeholder: "Linear, tactile…",
    },
    {
      key: "Формфактор",
      label: "Формфактор",
      type: "text",
      filterable: true,
      placeholder: "65%, 70%, 75%…",
    },
    {
      key: "Подключение",
      label: "Подключение",
      type: "text",
      filterable: true,
      options: [...CONNECTION_OPTIONS],
    },
    {
      key: "Раскладка",
      label: "Раскладка",
      type: "text",
      filterable: true,
      placeholder: "ANSI, ISO…",
    },
  ],
  Аксессуары: [
    {
      key: "Тип",
      label: "Тип",
      type: "text",
      filterable: true,
      placeholder: "Глайды, грипсы, рукава…",
    },
    {
      key: "Материал",
      label: "Материал",
      type: "text",
      filterable: true,
      placeholder: "PTFE, силикон…",
    },
  ],
};

export function getCategorySpecTemplate(
  categoryName: string | undefined
): CategorySpecField[] {
  if (!categoryName) {
    return [];
  }

  return CATEGORY_SPEC_TEMPLATES[categoryName as FixedCategoryName] ?? [];
}

export function specsFromTemplate(categoryName: string): ProductSpecification[] {
  return getCategorySpecTemplate(categoryName).map((field) => ({
    label: field.key,
    value: "",
  }));
}

/** Старые ключи характеристик → новые (при редактировании товара) */
const SPEC_LEGACY_ALIASES: Record<string, string> = {
  Формфактор: "Формат",
};

export function mergeSpecsWithTemplate(
  categoryName: string,
  existing: ProductSpecification[] = []
): ProductSpecification[] {
  const template = getCategorySpecTemplate(categoryName);
  const byLabel = new Map(existing.map((item) => [item.label, item.value]));

  return template.map((field) => {
    const legacyKey = SPEC_LEGACY_ALIASES[field.key];
    const value =
      byLabel.get(field.key) ??
      (legacyKey ? byLabel.get(legacyKey) : undefined) ??
      "";

    return {
      label: field.key,
      value,
    };
  });
}

export function isKnownCategoryName(name: string): name is FixedCategoryName {
  return FIXED_CATEGORIES.includes(name as FixedCategoryName);
}
