import { axiosInstance } from "./axiosInstance";

export interface ProductSpecification {
  label: string;
  value: string;
}

export interface ProductReview {
  user_id?: number;
  author: string;
  rating: number;
  text: string;
  created_at?: string;
}

export interface MyProductReview {
  product_id: number;
  product_title: string;
  product_image?: string;
  author: string;
  rating: number;
  text: string;
  created_at?: string;
}

export interface AdminProductReview {
  product_id: number;
  product_title: string;
  user_id?: number;
  author: string;
  rating: number;
  text: string;
  created_at?: string;
}

export interface AdminReviewsQuery {
  rating?: number;
  productId?: number;
  q?: string;
}

export interface UpsertProductReviewPayload {
  author: string;
  rating: number;
  text: string;
}

export interface Product {
  id: number;
  category_id?: number;
  category_name?: string;
  title: string;
  description: string;
  price: number;
  brand: string;
  stock: number;
  images: string[];
  specifications: ProductSpecification[];
  variants: string[];
  reviews: ProductReview[];
  rating_avg: number;
  rating_count: number;
  created_at: string;
}

export interface Category {
  id: number;
  name: string;
  created_at: string;
}

export interface ProductPayload {
  title: string;
  description: string;
  price: number;
  category_id?: number;
  brand?: string;
  stock?: number;
  images?: string[];
  specifications?: ProductSpecification[];
  variants?: string[];
  reviews?: ProductReview[];
}

export interface OrderItemPayload {
  product_id: number;
  quantity: number;
}

export interface CreateOrderPayload {
  contact_name: string;
  contact_phone: string;
  contact_email: string;
  delivery_address: string;
  delivery_type: "cdek" | "custom";
  delivery_city?: string;
  cdek_pvz_code?: string;
  delivery_payment?: "online" | "on_receipt";
  payment_method: string;
  comment?: string;
  items: OrderItemPayload[];
}

export interface Order {
  id: number;
  user_id: number;
  status: string;
  contact_name: string;
  contact_phone: string;
  contact_email: string;
  delivery_address: string;
  delivery_type?: "cdek" | "custom";
  delivery_cost?: number;
  delivery_city?: string;
  cdek_pvz_code?: string;
  delivery_payment?: "online" | "on_receipt";
  payment_method: string;
  comment: string;
  total_amount: number;
  created_at: string;
  items: Array<{
    id: number;
    product_id: number;
    title: string;
    quantity: number;
    price: number;
  }>;
}

export function productImageUrl(url?: string): string | undefined {
  if (!url) {
    return undefined;
  }

  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }

  return url.startsWith("/") ? url : `/${url}`;
}

export const productApi = {
  getProducts: async (categoryId?: number): Promise<Product[]> => {
    const response = await axiosInstance.get<Product[]>("/api/products", {
      params: categoryId ? { category_id: categoryId } : undefined,
    });
    return response.data;
  },

  getProduct: async (id: number): Promise<Product> => {
    const response = await axiosInstance.get<Product>(`/api/products/${id}`);
    return response.data;
  },

  getCategories: async (): Promise<Category[]> => {
    const response = await axiosInstance.get<Category[]>("/api/categories");
    return response.data;
  },

  createProduct: async (payload: ProductPayload): Promise<Product> => {
    const response = await axiosInstance.post<Product>("/api/products", payload);
    return response.data;
  },

  updateProduct: async (id: number, payload: ProductPayload): Promise<Product> => {
    const response = await axiosInstance.put<Product>(`/api/products/${id}`, payload);
    return response.data;
  },

  deleteProduct: async (id: number): Promise<void> => {
    await axiosInstance.delete(`/api/products/${id}`);
  },

  uploadImage: async (file: File): Promise<string> => {
    const formData = new FormData();
    formData.append("file", file);

    const response = await axiosInstance.post<{ url: string }>("/api/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });

    return response.data.url;
  },

  createOrder: async (payload: CreateOrderPayload): Promise<Order> => {
    const response = await axiosInstance.post<Order>("/api/orders", payload);
    return response.data;
  },

  getMyOrders: async (): Promise<Order[]> => {
    const response = await axiosInstance.get<Order[]>("/api/orders/my");
    return response.data;
  },

  getAllOrders: async (): Promise<Order[]> => {
    const response = await axiosInstance.get<Order[]>("/api/orders");
    return response.data;
  },

  updateOrderStatus: async (orderId: number, status: string): Promise<Order> => {
    const response = await axiosInstance.patch<Order>(
      `/api/orders/${orderId}/status`,
      { status }
    );
    return response.data;
  },

  getCourierOrders: async (): Promise<Order[]> => {
    const response = await axiosInstance.get<Order[]>("/api/courier/orders");
    return response.data;
  },

  updateCourierOrderStatus: async (orderId: number, status: string): Promise<Order> => {
    const response = await axiosInstance.patch<Order>(
      `/api/courier/orders/${orderId}/status`,
      { status }
    );
    return response.data;
  },

  getMyReviews: async (): Promise<MyProductReview[]> => {
    const response = await axiosInstance.get<MyProductReview[]>("/api/reviews/my");
    return response.data;
  },

  getAdminReviews: async (
    params: AdminReviewsQuery = {}
  ): Promise<AdminProductReview[]> => {
    const query = new URLSearchParams();
    if (params.rating) {
      query.set("rating", String(params.rating));
    }
    if (params.productId) {
      query.set("product_id", String(params.productId));
    }
    if (params.q?.trim()) {
      query.set("q", params.q.trim());
    }
    const suffix = query.toString();
    const response = await axiosInstance.get<AdminProductReview[]>(
      suffix ? `/api/reviews?${suffix}` : "/api/reviews"
    );
    return response.data;
  },

  upsertProductReview: async (
    productId: number,
    payload: UpsertProductReviewPayload
  ): Promise<Product> => {
    const response = await axiosInstance.put<Product>(
      `/api/products/${productId}/reviews`,
      payload
    );
    return response.data;
  },

  deleteProductReview: async (productId: number): Promise<Product> => {
    const response = await axiosInstance.delete<Product>(
      `/api/products/${productId}/reviews`
    );
    return response.data;
  },
};
