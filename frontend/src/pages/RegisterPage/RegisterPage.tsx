import { yupResolver } from "@hookform/resolvers/yup";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import * as yup from "yup";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { PasswordField } from "../../components/PasswordField/PasswordField";
import { clearAuthError, registerUser } from "../../features/auth/authSlice";
import authStyles from "../../styles/auth.module.css";

interface RegisterFormValues {
  email: string;
  password: string;
  repeatPassword: string;
  agreement: boolean;
}

const schema: yup.ObjectSchema<RegisterFormValues> = yup.object({
  email: yup
    .string()
    .email("Введите корректный email")
    .required("Email обязателен"),
  password: yup
    .string()
    .min(6, "Минимум 6 символов")
    .required("Пароль обязателен"),
  repeatPassword: yup
    .string()
    .oneOf([yup.ref("password")], "Пароли не совпадают")
    .required("Повторите пароль"),
  agreement: yup
    .boolean()
    .oneOf([true], "Необходимо согласие")
    .required(),
});

export function RegisterPage() {
  const dispatch = useAppDispatch();
  const { isLoading, error } = useAppSelector((state) => state.auth);

  const [successMessage, setSuccessMessage] = useState("");

  useEffect(() => {
    dispatch(clearAuthError());
  }, [dispatch]);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: yupResolver(schema),
    defaultValues: {
      email: "",
      password: "",
      repeatPassword: "",
      agreement: false,
    },
  });

  const onSubmit = async (values: RegisterFormValues) => {
    setSuccessMessage("");

    const result = await dispatch(
      registerUser({
        email: values.email,
        password: values.password,
      })
    );

    if (registerUser.fulfilled.match(result)) {
      setSuccessMessage("Аккаунт успешно создан");
    }
  };

  return (
    <div className="page">
      <Header />

      <main>
        <section className={authStyles.hero}>
          <form className={authStyles.card} onSubmit={handleSubmit(onSubmit)}>
            <div className={authStyles.heading}>
              <h1>Регистрация</h1>
              <p>
                Создайте аккаунт для удобства покупок, сохранения избранного и
                отслеживания заказов
              </p>
            </div>

            <label className={authStyles.field}>
              <span>Email</span>
              <input
                type="email"
                placeholder="example@mail.com"
                autoComplete="email"
                {...register("email")}
              />
              {errors.email && (
                <small className={authStyles.error}>{errors.email.message}</small>
              )}
            </label>

            <label className={authStyles.field}>
              <span>Пароль</span>
              <PasswordField
                registration={register("password")}
                autoComplete="new-password"
              />
              {errors.password && (
                <small className={authStyles.error}>
                  {errors.password.message}
                </small>
              )}
            </label>

            <label className={authStyles.field}>
              <span>Повторите пароль</span>
              <PasswordField
                registration={register("repeatPassword")}
                placeholder="Повторите пароль"
                autoComplete="off"
              />
              {errors.repeatPassword && (
                <small className={authStyles.error}>
                  {errors.repeatPassword.message}
                </small>
              )}
            </label>

            <label className={authStyles.checkbox}>
              <input type="checkbox" {...register("agreement")} />
              <span>
                Я согласен с{" "}
                <a href="#">политикой конфиденциальности</a>
              </span>
            </label>
            {errors.agreement && (
              <small className={authStyles.error}>{errors.agreement.message}</small>
            )}

            {error && <div className={authStyles.alertError}>{error}</div>}
            {successMessage && (
              <div className={authStyles.alertSuccess}>{successMessage}</div>
            )}

            <button className={authStyles.submit} type="submit" disabled={isLoading}>
              {isLoading ? "Регистрация..." : "Зарегистрироваться"}
            </button>

            <p className={authStyles.footerText}>
              Уже есть аккаунт? <Link to="/login">Войти</Link>
            </p>
          </form>
        </section>
      </main>

      <Footer />
    </div>
  );
}
