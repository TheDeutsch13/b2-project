import { Navigate, Route, Routes } from "react-router-dom";
import {
  ProfileIndexRedirect,
  ProfileLayout,
} from "../components/ProfileLayout/ProfileLayout";
import { AdminPage } from "../pages/AdminPage/AdminPage";
import { CourierPage } from "../pages/CourierPage/CourierPage";
import { CatalogPage } from "../pages/CatalogPage/CatalogPage";
import { CartPage } from "../pages/CartPage/CartPage";
import { CheckoutPage } from "../pages/CheckoutPage/CheckoutPage";
import { HomePage } from "../pages/HomePage/HomePage";
import { LoginPage } from "../pages/LoginPage/LoginPage";
import { OrdersPage } from "../pages/OrdersPage/OrdersPage";
import { ProductPage } from "../pages/ProductPage/ProductPage";
import { ProfileFavoritesPage } from "../pages/ProfileFavoritesPage/ProfileFavoritesPage";
import { ProfileOrdersPage } from "../pages/ProfileOrdersPage/ProfileOrdersPage";
import { ProfileReviewsPage } from "../pages/ProfileReviewsPage/ProfileReviewsPage";
import { ProfileSettingsPage } from "../pages/ProfileSettingsPage/ProfileSettingsPage";
import { RegisterPage } from "../pages/RegisterPage/RegisterPage";
import { ProtectedRoute } from "./ProtectedRoute";
import { PublicRoute } from "./PublicRoute";

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/product/:id" element={<ProductPage />} />
      <Route path="/cart" element={<CartPage />} />

      <Route element={<PublicRoute />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
      </Route>

      <Route element={<ProtectedRoute />}>
        <Route path="/catalog" element={<CatalogPage />} />
        <Route path="/checkout" element={<CheckoutPage />} />
        <Route path="/profile" element={<ProfileLayout />}>
          <Route index element={<ProfileIndexRedirect />} />
          <Route path="orders" element={<ProfileOrdersPage />} />
          <Route path="reviews" element={<ProfileReviewsPage />} />
          <Route path="favorites" element={<ProfileFavoritesPage />} />
          <Route path="settings" element={<ProfileSettingsPage />} />
        </Route>
        <Route path="/orders" element={<OrdersPage />} />
      </Route>

      <Route element={<ProtectedRoute roles={["admin", "moderator"]} />}>
        <Route path="/admin" element={<AdminPage />} />
      </Route>

      <Route element={<ProtectedRoute roles={["admin", "courier"]} />}>
        <Route path="/courier" element={<CourierPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
