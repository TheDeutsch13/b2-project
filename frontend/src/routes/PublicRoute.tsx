import { Navigate, Outlet } from "react-router-dom";
import { useAppSelector } from "../app/hooks";
import { getDefaultPathForRole, isUserRole } from "../constants/roles";

export function PublicRoute() {
  const { token, user, isInitialized } = useAppSelector((state) => state.auth);

  if (!isInitialized) {
    return <div className="page-loading">Загрузка...</div>;
  }

  if (token && user) {
    const target = isUserRole(user.role)
      ? getDefaultPathForRole(user.role)
      : "/";
    return <Navigate to={target} replace />;
  }

  return <Outlet />;
}
