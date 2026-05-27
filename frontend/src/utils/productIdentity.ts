import type { Product, ProductPayload } from "../api/productApi";

export function getPrimaryVariant(variants?: string[]): string {
  const first = variants?.map((item) => item.trim()).find(Boolean);
  return (first ?? "стандарт").toLowerCase();
}

export function productIdentityKey(input: {
  title: string;
  brand?: string;
  category_id?: number;
  variants?: string[];
}): string {
  return [
    input.title.trim().toLowerCase(),
    (input.brand ?? "").trim().toLowerCase(),
    String(input.category_id ?? ""),
    getPrimaryVariant(input.variants),
  ].join("|");
}

export function isDuplicateProduct(
  products: Product[],
  payload: ProductPayload,
  excludeId?: number
): boolean {
  const key = productIdentityKey(payload);

  return products.some(
    (product) =>
      product.id !== excludeId &&
      productIdentityKey({
        title: product.title,
        brand: product.brand,
        category_id: product.category_id,
        variants: product.variants,
      }) === key
  );
}
