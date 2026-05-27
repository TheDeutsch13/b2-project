import axios from "axios";
import { useEffect, useState } from "react";
import {
  productApi,
  productImageUrl,
  type Category,
  type Product,
  type ProductPayload,
  type ProductReview,
  type ProductSpecification,
} from "../../api/productApi";
import {
  getCategorySpecTemplate,
  mergeSpecsWithTemplate,
  specsFromTemplate,
} from "../../data/categorySpecifications";
import { collectSpecSuggestions } from "../../utils/specFilters";
import { isDuplicateProduct } from "../../utils/productIdentity";
import styles from "./AdminProductModal.module.css";

interface AdminProductModalProps {
  open: boolean;
  product: Product | null;
  categories: Category[];
  existingProducts: Product[];
  onClose: () => void;
  onSaved: () => void;
}

const emptyReview = (): ProductReview => ({ author: "", rating: 5, text: "" });

function categoryNameById(
  categories: Category[],
  id: string
): string | undefined {
  return categories.find((item) => String(item.id) === id)?.name;
}

export function AdminProductModal({
  open,
  product,
  categories,
  existingProducts,
  onClose,
  onSaved,
}: AdminProductModalProps) {
  const isEdit = product !== null;

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");
  const [brand, setBrand] = useState("");
  const [stock, setStock] = useState("0");
  const [categoryId, setCategoryId] = useState("");
  const [images, setImages] = useState<string[]>([]);
  const [variantsText, setVariantsText] = useState("");
  const [specifications, setSpecifications] = useState<ProductSpecification[]>(
    []
  );
  const [reviews, setReviews] = useState<ProductReview[]>([]);
  const [uploading, setUploading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [specSuggestions, setSpecSuggestions] = useState<
    Record<string, string[]>
  >({});

  useEffect(() => {
    if (!open) {
      return;
    }

    if (product) {
      setTitle(product.title);
      setDescription(product.description);
      setPrice(String(product.price));
      setBrand(product.brand || "");
      setStock(String(product.stock ?? 0));
      setCategoryId(product.category_id ? String(product.category_id) : "");
      setImages(product.images ?? []);
      setVariantsText((product.variants ?? []).join(", "));
      const categoryName = categoryNameById(
        categories,
        product.category_id ? String(product.category_id) : ""
      );
      setSpecifications(
        categoryName
          ? mergeSpecsWithTemplate(
              categoryName,
              product.specifications ?? []
            )
          : product.specifications ?? []
      );
      setReviews(product.reviews ?? []);
    } else {
      setTitle("");
      setDescription("");
      setPrice("");
      setBrand("");
      setStock("10");
      const firstCategoryId = categories[0] ? String(categories[0].id) : "";
      setCategoryId(firstCategoryId);
      setImages([]);
      setVariantsText("Стандарт");
      setSpecifications(
        categoryNameById(categories, firstCategoryId)
          ? specsFromTemplate(categoryNameById(categories, firstCategoryId)!)
          : []
      );
      setReviews([]);
    }

    setError("");
  }, [open, product, categories]);

  useEffect(() => {
    if (!open || !categoryId) {
      return;
    }

    const categoryName = categoryNameById(categories, categoryId);
    if (!categoryName) {
      return;
    }

    const loadSuggestions = async () => {
      try {
        const categoryProducts = await productApi.getProducts(
          Number(categoryId)
        );
        setSpecSuggestions(
          collectSpecSuggestions(categoryProducts, categoryName)
        );
      } catch {
        setSpecSuggestions({});
      }
    };

    void loadSuggestions();
  }, [open, categoryId, categories]);

  if (!open) {
    return null;
  }

  const selectedCategoryName = categoryNameById(categories, categoryId);
  const specTemplate = getCategorySpecTemplate(selectedCategoryName);

  const handleCategoryChange = (nextCategoryId: string) => {
    setCategoryId(nextCategoryId);
    const nextName = categoryNameById(categories, nextCategoryId);
    if (nextName) {
      setSpecifications(specsFromTemplate(nextName));
    }
  };

  const updateSpecValue = (key: string, value: string) => {
    setSpecifications((current) => {
      const hasKey = current.some((item) => item.label === key);
      if (!hasKey) {
        return [...current, { label: key, value }];
      }

      return current.map((item) =>
        item.label === key ? { ...item, value } : item
      );
    });
  };

  const handleUpload = async (files: FileList | null) => {
    if (!files?.length) {
      return;
    }

    setUploading(true);
    setError("");

    try {
      const uploadedUrls: string[] = [];

      for (const file of Array.from(files)) {
        const url = await productApi.uploadImage(file);
        uploadedUrls.push(url);
      }

      setImages((current) => [...current, ...uploadedUrls]);
    } catch {
      setError("Не удалось загрузить фото");
    } finally {
      setUploading(false);
    }
  };

  const buildPayload = (): ProductPayload => {
    const variants = variantsText
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);

    const specs = specifications.filter(
      (item) => item.label.trim() || item.value.trim()
    );

    const reviewItems = reviews.filter(
      (item) => item.author.trim() || item.text.trim()
    );

    return {
      title: title.trim(),
      description: description.trim(),
      price: Number(price),
      category_id: categoryId ? Number(categoryId) : undefined,
      brand: brand.trim(),
      stock: Number(stock),
      images,
      variants,
      specifications: specs,
      reviews: reviewItems,
    };
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setError("");

    try {
      const payload = buildPayload();

      if (
        isDuplicateProduct(
          existingProducts,
          payload,
          isEdit && product ? product.id : undefined
        )
      ) {
        setError(
          "Товар с таким названием, брендом, категорией и вариантом уже существует"
        );
        return;
      }

      if (isEdit && product) {
        await productApi.updateProduct(product.id, payload);
      } else {
        await productApi.createProduct(payload);
      }

      onSaved();
      onClose();
    } catch (err) {
      const apiMessage =
        axios.isAxiosError(err) &&
        typeof err.response?.data === "object" &&
        err.response?.data !== null &&
        "error" in err.response.data
          ? String((err.response.data as { error: string }).error)
          : null;

      setError(
        apiMessage ??
          (isEdit ? "Не удалось сохранить товар" : "Не удалось создать товар")
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div
        className={styles.modal}
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
      >
        <div className={styles.head}>
          <h2>{isEdit ? "Редактировать товар" : "Добавить товар"}</h2>
          <button type="button" className={styles.closeBtn} onClick={onClose}>
            ×
          </button>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          <div className={styles.field}>
            <label>Название</label>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          <div className={styles.row2}>
            <div className={styles.field}>
              <label>Категория</label>
              <select
                value={categoryId}
                onChange={(e) => handleCategoryChange(e.target.value)}
                required
              >
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </div>
            <div className={styles.field}>
              <label>Бренд</label>
              <input
                value={brand}
                onChange={(e) => setBrand(e.target.value)}
                placeholder="Logitech"
              />
            </div>
          </div>

          <div className={styles.row2}>
            <div className={styles.field}>
              <label>Цена</label>
              <input
                type="number"
                min="0"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>
            <div className={styles.field}>
              <label>Остаток</label>
              <input
                type="number"
                min="0"
                value={stock}
                onChange={(e) => setStock(e.target.value)}
                required
              />
            </div>
          </div>

          <div className={styles.field}>
            <label>Варианты (через запятую)</label>
            <input
              value={variantsText}
              onChange={(e) => setVariantsText(e.target.value)}
              placeholder="Чёрный, Белый"
            />
          </div>

          <div>
            <p className={styles.sectionTitle}>Фото товара</p>
            <p className={styles.hint}>
              {uploading ? "Загрузка..." : "JPG, PNG или WEBP до 5 МБ"}
            </p>
            <div className={styles.imagesGrid}>
              {images.map((url) => (
                <div key={url} className={styles.imagePreview}>
                  <img src={productImageUrl(url)} alt="" />
                  <button
                    type="button"
                    className={styles.removeImage}
                    onClick={() =>
                      setImages((current) => current.filter((item) => item !== url))
                    }
                  >
                    ×
                  </button>
                </div>
              ))}
              <label className={styles.uploadLabel}>
                + Фото
                <input
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  multiple
                  disabled={uploading}
                  onChange={(e) => {
                    void handleUpload(e.target.files);
                    e.target.value = "";
                  }}
                />
              </label>
            </div>
          </div>

          <div className={styles.field}>
            <label>Описание</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Текст для вкладки «Описание»"
            />
          </div>

          <div>
            <p className={styles.sectionTitle}>
              Характеристики
              {selectedCategoryName ? ` (${selectedCategoryName})` : ""}
            </p>
            <p className={styles.hint}>
              Указанные значения автоматически появятся в фильтрах каталога
            </p>
            {specTemplate.length === 0 ? (
              <p className={styles.hint}>Выберите категорию товара</p>
            ) : (
              <div className={styles.specList}>
                {specTemplate.map((field) => {
                  const specValue =
                    specifications.find((item) => item.label === field.key)
                      ?.value ?? "";
                  const listId = `spec-${field.key.replace(/\s+/g, "-")}`;
                  const suggestions = specSuggestions[field.key] ?? [];

                  const labelText = field.unit
                    ? `${field.label}, ${field.unit}`
                    : field.label;

                  return (
                    <div key={field.key} className={styles.specField}>
                      <label>{labelText}</label>
                      {field.options && field.options.length > 0 ? (
                        <select
                          value={specValue}
                          onChange={(e) =>
                            updateSpecValue(field.key, e.target.value)
                          }
                        >
                          <option value="">Не выбрано</option>
                          {field.options.map((option) => (
                            <option key={option} value={option}>
                              {option}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <>
                          <input
                            type={field.type === "number" ? "number" : "text"}
                            list={suggestions.length > 0 ? listId : undefined}
                            value={specValue}
                            placeholder={field.placeholder}
                            onChange={(e) =>
                              updateSpecValue(field.key, e.target.value)
                            }
                          />
                          {suggestions.length > 0 && (
                            <datalist id={listId}>
                              {suggestions.map((option) => (
                                <option key={option} value={option} />
                              ))}
                            </datalist>
                          )}
                        </>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div>
            <p className={styles.sectionTitle}>Отзывы</p>
            <div className={styles.dynamicList}>
              {reviews.map((item, index) => (
                <div key={index} className={styles.dynamicRowReview}>
                  <input
                    value={item.author}
                    placeholder="Имя"
                    onChange={(e) => {
                      const next = [...reviews];
                      next[index] = { ...item, author: e.target.value };
                      setReviews(next);
                    }}
                  />
                  <input
                    type="number"
                    min={1}
                    max={5}
                    value={item.rating}
                    onChange={(e) => {
                      const next = [...reviews];
                      next[index] = {
                        ...item,
                        rating: Number(e.target.value),
                      };
                      setReviews(next);
                    }}
                  />
                  <input
                    value={item.text}
                    placeholder="Текст отзыва"
                    onChange={(e) => {
                      const next = [...reviews];
                      next[index] = { ...item, text: e.target.value };
                      setReviews(next);
                    }}
                  />
                  <button
                    type="button"
                    className={styles.smallBtn}
                    onClick={() =>
                      setReviews((current) =>
                        current.filter((_, rowIndex) => rowIndex !== index)
                      )
                    }
                  >
                    −
                  </button>
                </div>
              ))}
            </div>
            <button
              type="button"
              className={styles.smallBtn}
              onClick={() => setReviews((current) => [...current, emptyReview()])}
            >
              + Отзыв
            </button>
          </div>

          {error && <p className={styles.hint}>{error}</p>}

          <div className={styles.actions}>
            <button type="button" className={styles.cancelBtn} onClick={onClose}>
              Отмена
            </button>
            <button type="submit" className={styles.saveBtn} disabled={saving}>
              {saving ? "Сохранение..." : isEdit ? "Сохранить" : "Создать"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
