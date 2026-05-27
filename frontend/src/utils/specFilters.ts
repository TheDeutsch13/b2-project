import type { Product } from "../api/productApi";
import {
  getCategorySpecTemplate,
  WEIGHT_SPEC_KEY,
} from "../data/categorySpecifications";

export interface DynamicSpecFilter {
  key: string;
  label: string;
  options: string[];
}

export function getProductSpecValue(
  product: Product,
  specKey: string
): string {
  const match = product.specifications?.find(
    (item) => item.label.toLowerCase() === specKey.toLowerCase()
  );

  return match?.value?.trim() ?? "";
}

/** Собирает фильтры из реальных значений характеристик товаров категории */
export function buildDynamicSpecFilters(
  products: Product[],
  categoryName?: string
): DynamicSpecFilter[] {
  const template = getCategorySpecTemplate(categoryName);
  const filterableFields = template.filter(
    (field) => field.filterable && field.filterMode !== "range"
  );

  if (!categoryName || filterableFields.length === 0) {
    return [];
  }

  const valuesByKey = new Map<string, Set<string>>();

  for (const product of products) {
    for (const spec of product.specifications ?? []) {
      const specLabel = spec.label.trim();
      const specValue = spec.value.trim();
      if (!specLabel || !specValue) {
        continue;
      }

      const field = filterableFields.find(
        (item) => item.key.toLowerCase() === specLabel.toLowerCase()
      );
      if (!field) {
        continue;
      }

      if (!valuesByKey.has(field.key)) {
        valuesByKey.set(field.key, new Set());
      }
      valuesByKey.get(field.key)!.add(specValue);
    }
  }

  return filterableFields
    .map((field) => {
      const values = valuesByKey.get(field.key);
      if (!values?.size) {
        return null;
      }

      return {
        key: field.key,
        label: field.label,
        options: [...values].sort((a, b) => a.localeCompare(b, "ru")),
      };
    })
    .filter((field): field is DynamicSpecFilter => field !== null);
}

export function collectSpecSuggestions(
  products: Product[],
  categoryName?: string
): Record<string, string[]> {
  const result: Record<string, string[]> = {};
  const filters = buildDynamicSpecFilters(products, categoryName);

  for (const field of filters) {
    result[field.key] = field.options;
  }

  const weightValues = new Set<string>();
  for (const product of products) {
    const raw = getProductSpecValue(product, WEIGHT_SPEC_KEY);
    if (raw) {
      weightValues.add(raw);
    }
  }
  if (weightValues.size > 0) {
    result[WEIGHT_SPEC_KEY] = [...weightValues].sort((a, b) =>
      a.localeCompare(b, "ru")
    );
  }

  return result;
}

export function parseWeightGrams(value: string): number | null {
  const normalized = value.trim().toLowerCase().replace(",", ".");
  const match = normalized.match(/(\d+(?:\.\d+)?)/);
  if (!match) {
    return null;
  }

  const amount = parseFloat(match[1]);
  if (Number.isNaN(amount)) {
    return null;
  }

  if (normalized.includes("кг") || /\bkg\b/.test(normalized)) {
    return amount * 1000;
  }

  return amount;
}

export function categoryHasWeightFilter(categoryName?: string): boolean {
  if (!categoryName) {
    return false;
  }

  return getCategorySpecTemplate(categoryName).some(
    (field) => field.key === WEIGHT_SPEC_KEY && field.filterable
  );
}

export function productMatchesWeightRange(
  product: Product,
  weightMin: string,
  weightMax: string,
  specKey: string = WEIGHT_SPEC_KEY
): boolean {
  const min = weightMin.trim() ? Number(weightMin) : null;
  const max = weightMax.trim() ? Number(weightMax) : null;

  if (min === null && max === null) {
    return true;
  }

  const grams = parseWeightGrams(getProductSpecValue(product, specKey));
  if (grams === null) {
    return false;
  }

  if (min !== null && !Number.isNaN(min) && grams < min) {
    return false;
  }

  if (max !== null && !Number.isNaN(max) && grams > max) {
    return false;
  }

  return true;
}

export function buildBrandFilterOptions(products: Product[]): string[] {
  const brands = new Set<string>();

  for (const product of products) {
    const brand = product.brand?.trim();
    if (brand) {
      brands.add(brand);
    }
  }

  return [...brands].sort((a, b) => a.localeCompare(b, "ru"));
}

export function productMatchesBrand(
  product: Product,
  selectedBrand: string
): boolean {
  const brand = selectedBrand.trim();
  if (!brand) {
    return true;
  }

  return (product.brand?.trim() ?? "") === brand;
}

export function productMatchesSpecFilters(
  product: Product,
  filters: DynamicSpecFilter[],
  selected: Record<string, string>
): boolean {
  for (const field of filters) {
    const filterValue = selected[field.key]?.trim();
    if (!filterValue) {
      continue;
    }

    const productValue = getProductSpecValue(product, field.key);
    if (!productValue || productValue !== filterValue) {
      return false;
    }
  }

  return true;
}

export function productMatchesSearch(product: Product, query: string): boolean {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return true;
  }

  const haystack = [
    product.title,
    product.description,
    product.brand,
    product.category_name ?? "",
    ...(product.specifications?.flatMap((item) => [item.label, item.value]) ??
      []),
  ]
    .join(" ")
    .toLowerCase();

  return haystack.includes(normalized);
}

export function productMatchesPriceRange(
  product: Product,
  priceMin: string,
  priceMax: string
): boolean {
  const min = priceMin.trim() ? Number(priceMin) : null;
  const max = priceMax.trim() ? Number(priceMax) : null;

  if (min !== null && !Number.isNaN(min) && product.price < min) {
    return false;
  }

  if (max !== null && !Number.isNaN(max) && product.price > max) {
    return false;
  }

  return true;
}
