import { yupResolver } from "@hookform/resolvers/yup";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useLocation, useNavigate } from "react-router-dom";
import * as yup from "yup";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { PasswordField } from "../../components/PasswordField/PasswordField";
import { getDefaultPathForRole, isUserRole } from "../../constants/roles";
import { clearAuthError, loginUser } from "../../features/auth/authSlice";
import authStyles from "../../styles/auth.module.css";

interface LoginFormValues {
  email: string;
  password: string;
}

const schema: yup.ObjectSchema<LoginFormValues> = yup.object({
  email: yup.string().email("Введите корректный email").required("Email обязателен"),
  password: yup.string().min(6, "Минимум 6 символов").required("Пароль обязателен"),
});

export function LoginPage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { isLoading, error } = useAppSelector((state) => state.auth);

  const redirectPath =
    (location.state as { from?: { pathname?: string } } | null)?.from
      ?.pathname ?? null;

  useEffect(() => {
    dispatch(clearAuthError());
  }, [dispatch]);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: yupResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = async (values: LoginFormValues) => {
    const result = await dispatch(loginUser(values));

    if (loginUser.fulfilled.match(result)) {
      const role = result.payload.user.role;

      if (role === "user" && redirectPath === "/checkout") {
        navigate("/checkout", { replace: true });
        return;
      }

      navigate(isUserRole(role) ? getDefaultPathForRole(role) : "/", {
        replace: true,
      });
    }
  };

  return (
    <div className="page">
      <Header />
      <main>
        <section className={authStyles.hero}>
          <form className={authStyles.card} onSubmit={handleSubmit(onSubmit)}>
            <div className={authStyles.heading}>
              <h1>Вход</h1>
              <p>
                Давайте войдем в аккаунт для того, чтобы вы смогли использовать
                полные возможности сайта
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
                autoComplete="current-password"
              />
              {errors.password && (
                <small className={authStyles.error}>{errors.password.message}</small>
              )}
            </label>

            <Link to="#" className={authStyles.forgotLink}>
              Забыли пароль?
            </Link>

            {error && <div className={authStyles.alertError}>{error}</div>}

            <button className={authStyles.submit} type="submit" disabled={isLoading}>
              {isLoading ? "Вход..." : "Войти"}
            </button>

            <p className={authStyles.footerText}>
              Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
            </p>
          </form>
        </section>
      </main>
      <Footer />
    </div>
  );
}
