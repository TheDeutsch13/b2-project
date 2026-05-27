import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAppSelector } from "../app/hooks";
import {
  getDefaultPathForRole,
  isUserRole,
  type UserRole,
} from "../constants/roles";

interface ProtectedRouteProps {
  roles?: UserRole[];
}

export function ProtectedRoute({ roles }: ProtectedRouteProps) {
  const location = useLocation();
  const { token, user, isInitialized } = useAppSelector((state) => state.auth);

  if (!isInitialized) {
    return <div className="page-loading">Загрузка...</div>;
  }

  if (!token || !user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (roles && !roles.includes(user.role)) {
    const fallback = isUserRole(user.role)
      ? getDefaultPathForRole(user.role)
      : "/";
    return <Navigate to={fallback} replace />;
  }

  return <Outlet />;
}
