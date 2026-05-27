import { productImageUrl, type Product } from "../api/productApi";

/** URL первого загруженного фото или undefined, если фото нет. */
export function getProductPrimaryImage(
  product: Pick<Product, "images">
): string | undefined {
  const uploaded = product.images?.[0];
  return uploaded ? productImageUrl(uploaded) : undefined;
}

/** Все загруженные фото товара (пустой массив, если нет). */
export function getProductGalleryImages(product: Pick<Product, "images">): string[] {
  if (!product.images?.length) {
    return [];
  }

  return product.images
    .map((url) => productImageUrl(url))
    .filter((url): url is string => Boolean(url));
}

export function isProductInStock(product: Pick<Product, "stock">): boolean {
  return product.stock > 0;
}
