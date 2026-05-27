export interface StoredProfile {
  displayName: string;
  avatarUrl: string;
}

const PREFIX = "gamegear_profile_";

export function profileStorageKey(email: string): string {
  return `${PREFIX}${email.toLowerCase()}`;
}

export function readProfile(email: string): StoredProfile {
  try {
    const raw = localStorage.getItem(profileStorageKey(email));
    if (!raw) {
      return { displayName: "", avatarUrl: "" };
    }

    const parsed = JSON.parse(raw) as StoredProfile;
    return {
      displayName: parsed.displayName ?? "",
      avatarUrl: parsed.avatarUrl ?? "",
    };
  } catch {
    return { displayName: "", avatarUrl: "" };
  }
}

export function writeProfile(email: string, profile: StoredProfile): void {
  localStorage.setItem(profileStorageKey(email), JSON.stringify(profile));
}
