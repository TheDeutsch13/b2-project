export const USER_ROLES = [
  "user",
  "admin",
  "moderator",
  "courier",
] as const;

export type UserRole = (typeof USER_ROLES)[number];

export const ROLE_LABELS: Record<UserRole, string> = {
  user: "Пользователь",
  admin: "Администратор",
  moderator: "Модератор",
  courier: "Курьер",
};

export const ASSIGNABLE_ROLES: UserRole[] = [
  "user",
  "admin",
  "moderator",
  "courier",
];

export function isUserRole(value: string): value is UserRole {
  return (USER_ROLES as readonly string[]).includes(value);
}

export function getRoleLabel(role: string): string {
  if (isUserRole(role)) {
    return ROLE_LABELS[role];
  }

  return role;
}

export function getDefaultPathForRole(role: UserRole): string {
  switch (role) {
    case "admin":
    case "moderator":
      return "/admin";
    case "courier":
      return "/courier";
    default:
      return "/";
  }
}

export function canAccessAdminPanel(role: UserRole): boolean {
  return role === "admin" || role === "moderator";
}

export function canAccessCourierPanel(role: UserRole): boolean {
  return role === "courier" || role === "admin";
}
