import type { FixedCategoryName } from "./categories";
import heroMice from "../assets/home/hero-mice.png";
import slideMouseTop from "../assets/home/slide-mouse-top.png";
import slideMouseSide from "../assets/home/slide-mouse-side.png";
import slideMouseAlt from "../assets/home/slide-mouse-alt.png";
import mouseProduct from "../assets/products/mouse-prox2.png";

export const heroMiceImage = heroMice;

/** Имена файлов в frontend/public/home/categories/ — можно заменить без пересборки */
export const categoryPublicFiles: Record<FixedCategoryName, string> = {
  Мыши: "mice.png",
  Коврики: "pads.png",
  Клавиатуры: "keyboards.png",
  Аксессуары: "accessories.png",
};

/** Запасные слайды, если папка Carousel пуста */
export const fallbackPromoSlides = [
  {
    id: "mice-collection",
    image: heroMice,
    title: "Новая коллекция мышей",
    description: "Премиальные модели для FPS, MOBA и универсального гейминга",
    link: "/catalog",
  },
  {
    id: "pro-x2",
    image: slideMouseTop,
    title: "Logitech PRO X2",
    description: "Лёгкий корпус, точный сенсор и фирменный хват под соревнования",
    link: "/catalog",
  },
] as const;

/** Запасные картинки (если в public/home/categories/ нет своего файла) */
export const categoryImages: Record<string, string> = {
  Мыши: slideMouseTop,
  Коврики: slideMouseSide,
  Клавиатуры: slideMouseAlt,
  Аксессуары: mouseProduct,
};

export function getCategoryImageSources(categoryName: string): {
  primary: string;
  fallback: string;
} {
  const fileName =
    categoryPublicFiles[categoryName as FixedCategoryName] ?? null;
  const fallback = categoryImages[categoryName] ?? slideMouseTop;

  if (fileName) {
    return {
      primary: `/home/categories/${fileName}`,
      fallback,
    };
  }

  return { primary: fallback, fallback };
}

export function getProductImage(index: number): string {
  const images = [mouseProduct, slideMouseTop, slideMouseSide, slideMouseAlt];
  return images[index % images.length];
}
