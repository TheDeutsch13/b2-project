import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import type { UseFormRegisterReturn } from "react-hook-form";
import styles from "../../styles/auth.module.css";

interface PasswordFieldProps {
  registration: UseFormRegisterReturn;
  placeholder?: string;
  autoComplete?: "current-password" | "new-password" | "off";
}

export function PasswordField({
  registration,
  placeholder = "Не менее 6 символов",
  autoComplete = "current-password",
}: PasswordFieldProps) {
  const [visible, setVisible] = useState(false);

  return (
    <div className={styles.passwordInput}>
      <input
        type={visible ? "text" : "password"}
        placeholder={placeholder}
        autoComplete={autoComplete}
        {...registration}
      />
      <button
        type="button"
        className={styles.passwordToggle}
        onClick={() => setVisible((value) => !value)}
        aria-label={visible ? "Скрыть пароль" : "Показать пароль"}
      >
        {visible ? <Eye size={16} /> : <EyeOff size={16} />}
      </button>
    </div>
  );
}
