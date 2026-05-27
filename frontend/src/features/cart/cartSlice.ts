import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export interface CartItem {
  productId: number;
  title: string;
  price: number;
  quantity: number;
  imageSrc?: string;
}

interface CartState {
  items: CartItem[];
  isOpen: boolean;
}

const initialState: CartState = {
  items: [],
  isOpen: false,
};

const cartSlice = createSlice({
  name: "cart",
  initialState,
  reducers: {
    openCart: (state) => {
      state.isOpen = true;
    },
    closeCart: (state) => {
      state.isOpen = false;
    },
    toggleCart: (state) => {
      state.isOpen = !state.isOpen;
    },
    addToCart: (state, action: PayloadAction<Omit<CartItem, "quantity">>) => {
      const existing = state.items.find(
        (item) => item.productId === action.payload.productId
      );

      if (existing) {
        existing.quantity += 1;
      } else {
        state.items.push({ ...action.payload, quantity: 1 });
      }

      state.isOpen = true;
    },
    removeFromCart: (state, action: PayloadAction<number>) => {
      state.items = state.items.filter(
        (item) => item.productId !== action.payload
      );
    },
    setQuantity: (
      state,
      action: PayloadAction<{ productId: number; quantity: number }>
    ) => {
      const item = state.items.find(
        (entry) => entry.productId === action.payload.productId
      );

      if (!item) {
        return;
      }

      if (action.payload.quantity <= 0) {
        state.items = state.items.filter(
          (entry) => entry.productId !== action.payload.productId
        );
        return;
      }

      item.quantity = action.payload.quantity;
    },
    clearCart: (state) => {
      state.items = [];
    },
  },
});

export const {
  openCart,
  closeCart,
  toggleCart,
  addToCart,
  removeFromCart,
  setQuantity,
  clearCart,
} = cartSlice.actions;

export default cartSlice.reducer;
