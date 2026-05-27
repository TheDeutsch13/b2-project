import type { AuthUser } from "../features/auth/authSlice";

export interface ProfileNameFields {
  firstName?: string;
  lastName?: string;
  nickname?: string;
}

export function getDefaultDisplayName(user: AuthUser): string {
  const localPart = user.email.split("@")[0]?.trim();
  if (!localPart) {
    return "Пользователь";
  }

  return localPart
    .replace(/[._-]+/g, " ")
    .split(" ")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function getProfileFullName(profile: ProfileNameFields): string {
  const fullName = [profile.firstName?.trim(), profile.lastName?.trim()]
    .filter(Boolean)
    .join(" ");

  return fullName;
}

export function getProfileDisplayName(
  user: AuthUser,
  profile: ProfileNameFields
): string {
  const fullName = getProfileFullName(profile);
  if (fullName) {
    return fullName;
  }

  const nickname = profile.nickname?.trim();
  if (nickname) {
    return nickname;
  }

  return getDefaultDisplayName(user);
}

export function getAvatarInitial(name: string): string {
  const firstChar = name.trim().charAt(0);
  return firstChar ? firstChar.toUpperCase() : "?";
}

export function userAvatarUrl(url?: string): string | undefined {
  if (!url) {
    return undefined;
  }

  if (
    url.startsWith("http://") ||
    url.startsWith("https://") ||
    url.startsWith("data:")
  ) {
    return url;
  }

  return url.startsWith("/") ? url : `/${url}`;
}

export function formatBirthDateForInput(iso?: string): string {
  if (!iso) {
    return "";
  }

  return iso.slice(0, 10);
}

export function formatBirthDateForDisplay(iso?: string): string {
  if (!iso) {
    return "";
  }

  const [year, month, day] = iso.slice(0, 10).split("-");
  if (!year || !month || !day) {
    return iso;
  }

  return `${day}.${month}.${year}`;
}
