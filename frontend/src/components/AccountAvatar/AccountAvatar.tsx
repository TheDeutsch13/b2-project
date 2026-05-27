import styles from "./AccountAvatar.module.css";

interface AccountAvatarProps {
  name: string;
  avatarUrl?: string;
  size?: "sm" | "md" | "lg";
  onClick?: () => void;
}

export function AccountAvatar({
  name,
  avatarUrl,
  size = "lg",
  onClick,
}: AccountAvatarProps) {
  const initial = name.trim().charAt(0).toUpperCase() || "?";
  const className = `${styles.avatar} ${styles[size]} ${
    onClick ? styles.clickable : ""
  }`;

  const Tag = onClick ? "button" : "div";

  return (
    <Tag
      type={onClick ? "button" : undefined}
      className={className}
      onClick={onClick}
      aria-label={onClick ? "Аватар" : undefined}
    >
      {avatarUrl ? <img src={avatarUrl} alt="" /> : <span>{initial}</span>}
    </Tag>
  );
}
