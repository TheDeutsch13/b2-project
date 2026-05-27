import { useEffect, useRef, useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import {
  fetchProfile,
  removeAvatar,
  saveProfile,
  uploadAvatar,
} from "../../features/profile/profileSlice";
import {
  getProfileDisplayName,
  getProfileFullName,
  userAvatarUrl,
} from "../../utils/userDisplay";
import { AccountAvatar } from "../../components/AccountAvatar/AccountAvatar";
import styles from "./ProfileSettingsPage.module.css";

const MAX_AVATAR_BYTES = 2 * 1024 * 1024;

export function ProfileSettingsPage() {
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const profile = useAppSelector((state) => state.profile);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [basicOpen, setBasicOpen] = useState(true);
  const [contactsOpen, setContactsOpen] = useState(true);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [nickname, setNickname] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [gender, setGender] = useState<"" | "male" | "female">("");
  const [phone, setPhone] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    void dispatch(fetchProfile());
  }, [dispatch]);

  useEffect(() => {
    setFirstName(profile.firstName);
    setLastName(profile.lastName);
    setNickname(profile.nickname);
    setBirthDate(profile.birthDate);
    setGender(profile.gender);
    setPhone(profile.phone);
  }, [
    profile.firstName,
    profile.lastName,
    profile.nickname,
    profile.birthDate,
    profile.gender,
    profile.phone,
  ]);

  if (!user) {
    return null;
  }

  const displayName = getProfileDisplayName(user, {
    firstName,
    lastName,
    nickname,
  });

  const buildPayload = () => ({
    first_name: firstName.trim(),
    last_name: lastName.trim(),
    nickname: nickname.trim(),
    birth_date: birthDate || undefined,
    gender,
    phone: phone.trim(),
  });

  const handleSaveBasic = async () => {
    if (!firstName.trim() || !lastName.trim()) {
      setError("Укажите имя и фамилию");
      return;
    }

    setError("");
    setMessage("");

    const result = await dispatch(saveProfile(buildPayload()));

    if (saveProfile.fulfilled.match(result)) {
      setMessage("Основная информация сохранена");
    } else {
      setError(profile.error ?? "Не удалось сохранить");
    }
  };

  const handleSaveContacts = async () => {
    setError("");
    setMessage("");

    const result = await dispatch(saveProfile(buildPayload()));

    if (saveProfile.fulfilled.match(result)) {
      setMessage("Контакты сохранены");
    } else {
      setError(profile.error ?? "Не удалось сохранить");
    }
  };

  const handleAvatarChange = async (file: File | undefined) => {
    if (!file) {
      return;
    }

    if (!file.type.startsWith("image/")) {
      setError("Выберите JPG, PNG или WEBP");
      return;
    }

    if (file.size > MAX_AVATAR_BYTES) {
      setError("Фото должно быть меньше 2 МБ");
      return;
    }

    setError("");
    const result = await dispatch(uploadAvatar(file));

    if (uploadAvatar.fulfilled.match(result)) {
      setMessage("Фото профиля обновлено");
    } else {
      setError(profile.error ?? "Не удалось загрузить фото");
    }
  };

  const handleRemoveAvatar = async () => {
    setError("");
    const result = await dispatch(removeAvatar(buildPayload()));

    if (removeAvatar.fulfilled.match(result)) {
      setMessage("Фото удалено");
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } else {
      setError(profile.error ?? "Не удалось удалить фото");
    }
  };

  return (
    <section className={styles.page}>
      <h1 className={styles.title}>Профиль</h1>

      {message && <p className={styles.message}>{message}</p>}
      {error && <p className={styles.error}>{error}</p>}

      <article className={styles.card}>
        <header className={styles.cardHead}>
          <h2>Основная информация</h2>
          <button
            type="button"
            className={styles.collapseBtn}
            onClick={() => setBasicOpen((value) => !value)}
            aria-expanded={basicOpen}
          >
            {basicOpen ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
          </button>
        </header>

        {basicOpen && (
          <div className={styles.cardBody}>
            <div className={styles.avatarRow}>
              <AccountAvatar
                name={displayName}
                avatarUrl={userAvatarUrl(profile.avatarUrl)}
                size="lg"
              />
              <div className={styles.avatarActions}>
                <button
                  type="button"
                  className={styles.primaryOutlineBtn}
                  disabled={profile.isSaving}
                  onClick={() => fileInputRef.current?.click()}
                >
                  Обновить фото
                </button>
                {profile.avatarUrl && (
                  <button
                    type="button"
                    className={styles.secondaryBtn}
                    disabled={profile.isSaving}
                    onClick={() => void handleRemoveAvatar()}
                  >
                    Удалить фото
                  </button>
                )}
                <p className={styles.hint}>
                  Минимальный размер аватара — 320×320 пикселей
                </p>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  className={styles.hiddenInput}
                  onChange={(event) =>
                    void handleAvatarChange(event.target.files?.[0])
                  }
                />
              </div>
            </div>

            <div className={styles.fieldGrid}>
              <label className={styles.field}>
                <span>*Имя</span>
                <input
                  value={firstName}
                  onChange={(event) => setFirstName(event.target.value)}
                  required
                />
              </label>
              <label className={styles.field}>
                <span>*Фамилия</span>
                <input
                  value={lastName}
                  onChange={(event) => setLastName(event.target.value)}
                  required
                />
              </label>
              <label className={`${styles.field} ${styles.fieldWide}`}>
                <span>Никнейм</span>
                <input
                  value={nickname}
                  onChange={(event) => setNickname(event.target.value)}
                  placeholder="Ваш никнейм"
                />
              </label>
              <label className={styles.field}>
                <span>Дата рождения</span>
                <input
                  type="date"
                  value={birthDate}
                  onChange={(event) => setBirthDate(event.target.value)}
                />
              </label>
            </div>

            <div className={styles.genderBlock}>
              <span className={styles.genderLabel}>Пол</span>
              <div className={styles.genderOptions}>
                <button
                  type="button"
                  className={`${styles.genderBtn} ${gender === "male" ? styles.genderBtnActive : ""}`}
                  onClick={() => setGender("male")}
                >
                  Мужчина
                </button>
                <button
                  type="button"
                  className={`${styles.genderBtn} ${gender === "female" ? styles.genderBtnActive : ""}`}
                  onClick={() => setGender("female")}
                >
                  Женщина
                </button>
              </div>
            </div>

            <button
              type="button"
              className={styles.saveBtn}
              disabled={profile.isSaving}
              onClick={() => void handleSaveBasic()}
            >
              {profile.isSaving ? "Сохранение…" : "Сохранить изменения"}
            </button>
          </div>
        )}
      </article>

      <article className={styles.card}>
        <header className={styles.cardHead}>
          <h2>Контакты</h2>
          <button
            type="button"
            className={styles.collapseBtn}
            onClick={() => setContactsOpen((value) => !value)}
            aria-expanded={contactsOpen}
          >
            {contactsOpen ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
          </button>
        </header>

        {contactsOpen && (
          <div className={styles.cardBody}>
            <div className={styles.fieldGrid}>
              <label className={styles.field}>
                <span>Номер телефона</span>
                <input
                  value={phone}
                  onChange={(event) => setPhone(event.target.value)}
                  placeholder="+7 (___) ___-__-__"
                />
              </label>
              <label className={styles.field}>
                <span>Почта</span>
                <input value={user.email} disabled />
              </label>
            </div>

            <div className={styles.actionsRow}>
              <button
                type="button"
                className={styles.saveBtn}
                disabled={profile.isSaving}
                onClick={() => void handleSaveContacts()}
              >
                {profile.isSaving ? "Сохранение…" : "Сохранить изменения"}
              </button>
              <button
                type="button"
                className={styles.secondaryBtn}
                onClick={() => {
                  setPhone(profile.phone);
                  setError("");
                  setMessage("");
                }}
              >
                Отмена
              </button>
            </div>
          </div>
        )}
      </article>

      {getProfileFullName({ firstName, lastName }) && (
        <p className={styles.previewName}>
          Отображаемое имя: <strong>{displayName}</strong>
        </p>
      )}
    </section>
  );
}
