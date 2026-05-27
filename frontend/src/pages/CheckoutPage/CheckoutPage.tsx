import { yupResolver } from "@hookform/resolvers/yup";
import {
  CreditCard,
  MapPin,
  Package,
  Truck,
  User,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import * as yup from "yup";
import { productApi } from "../../api/productApi";
import type { CdekPickupPoint } from "../../api/cdekApi";
import { useAppDispatch, useAppSelector } from "../../app/hooks";
import { CdekPvzModal } from "../../components/CdekPvzModal/CdekPvzModal";
import { Footer } from "../../components/Footer/Footer";
import { Header } from "../../components/Header/Header";
import { clearCart } from "../../features/cart/cartSlice";
import {
  DELIVERY_COSTS,
  DELIVERY_LABELS,
  type DeliveryType,
} from "../../data/deliveryConfig";
import styles from "./CheckoutPage.module.css";

interface ContactFormValues {
  contactName: string;
  contactPhone: string;
  contactEmail: string;
}

const contactSchema: yup.ObjectSchema<ContactFormValues> = yup.object({
  contactName: yup.string().required("Укажите имя"),
  contactPhone: yup.string().required("Укажите телефон"),
  contactEmail: yup.string().email("Некорректный email").required("Укажите email"),
});

const steps = [
  { key: "data", label: "Данные", icon: User },
  { key: "delivery", label: "Доставка", icon: Truck },
  { key: "payment", label: "Оплата", icon: CreditCard },
] as const;

export function CheckoutPage() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  const cartItems = useAppSelector((state) => state.cart.items);

  const [step, setStep] = useState(0);
  const [deliveryType, setDeliveryType] = useState<DeliveryType>("cdek");
  const [deliveryCity, setDeliveryCity] = useState("Саратов");
  const [selectedPvz, setSelectedPvz] = useState<CdekPickupPoint | null>(null);
  const [customAddress, setCustomAddress] = useState("");
  const [deliveryPayment, setDeliveryPayment] = useState<"online" | "on_receipt">(
    "on_receipt"
  );
  const [comment, setComment] = useState("");
  const [pvzModalOpen, setPvzModalOpen] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deliveryError, setDeliveryError] = useState("");

  const {
    register,
    handleSubmit,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<ContactFormValues>({
    resolver: yupResolver(contactSchema),
    defaultValues: {
      contactName: "",
      contactPhone: "",
      contactEmail: user?.email ?? "",
    },
  });

  const itemsSubtotal = useMemo(
    () => cartItems.reduce((sum, item) => sum + item.price * item.quantity, 0),
    [cartItems]
  );
  const deliveryCost = DELIVERY_COSTS[deliveryType];
  const payNow =
    itemsSubtotal + (deliveryPayment === "online" ? deliveryCost : 0);
  const payOnReceipt =
    deliveryPayment === "on_receipt" ? deliveryCost : 0;
  const orderTotal = itemsSubtotal + deliveryCost;
  const itemsCount = cartItems.reduce((sum, item) => sum + item.quantity, 0);

  useEffect(() => {
    if (cartItems.length === 0) {
      navigate("/", { replace: true });
    }
  }, [cartItems.length, navigate]);

  const goNext = async () => {
    if (step === 0) {
      const valid = await trigger();
      if (!valid) {
        return;
      }
      setStep(1);
      return;
    }

    if (step === 1) {
      setDeliveryError("");

      if (deliveryType === "cdek") {
        if (!deliveryCity.trim()) {
          setDeliveryError("Укажите город");
          return;
        }
        if (!selectedPvz) {
          setDeliveryError("Выберите пункт выдачи СДЭК");
          return;
        }
      } else if (!customAddress.trim()) {
        setDeliveryError("Укажите адрес доставки");
        return;
      }

      setStep(2);
    }
  };

  const buildDeliveryAddress = (): string => {
    if (deliveryType === "cdek" && selectedPvz) {
      return `СДЭК ${selectedPvz.code}: ${selectedPvz.address}`;
    }

    return `${deliveryCity.trim()}, ${customAddress.trim()}`;
  };

  const onSubmit = async () => {
    setIsSubmitting(true);
    setSubmitError("");

    const contacts = getValues();

    try {
      await productApi.createOrder({
        contact_name: contacts.contactName,
        contact_phone: contacts.contactPhone,
        contact_email: contacts.contactEmail,
        delivery_address: buildDeliveryAddress(),
        delivery_type: deliveryType,
        delivery_city: deliveryCity.trim(),
        cdek_pvz_code:
          deliveryType === "cdek" ? selectedPvz?.code ?? "" : undefined,
        delivery_payment: deliveryPayment,
        payment_method: "card",
        comment,
        items: cartItems.map((item) => ({
          product_id: item.productId,
          quantity: item.quantity,
        })),
      });

      dispatch(clearCart());
      navigate("/profile/orders", { replace: true });
    } catch {
      setSubmitError("Не удалось оформить заказ. Проверьте авторизацию и product-service.");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (cartItems.length === 0) {
    return null;
  }

  return (
    <div className="page">
      <Header />
      <main className={`container ${styles.main}`}>
        <nav className={styles.breadcrumbs}>
          <Link to="/">Главная</Link>
          <span>/</span>
          <span>Оформление заказа</span>
        </nav>

        <h1 className={styles.title}>Оформление заказа</h1>

        <div className={styles.layout}>
          <div className={styles.content}>
            <div className={styles.steps}>
              {steps.map(({ label, icon: Icon }, index) => (
                <div
                  key={label}
                  className={`${styles.step} ${
                    index <= step ? styles.stepActive : ""
                  }`}
                >
                  <Icon size={18} />
                  <span>{label}</span>
                </div>
              ))}
            </div>

            <div className={styles.card}>
              {step === 0 && (
                <section>
                  <h2>Личные данные</h2>
                  <label>
                    Имя *
                    <input {...register("contactName")} placeholder="Иван" />
                    {errors.contactName && (
                      <small>{errors.contactName.message}</small>
                    )}
                  </label>
                  <label>
                    Телефон *
                    <input
                      {...register("contactPhone")}
                      placeholder="+7 (999) 000-00-00"
                    />
                    {errors.contactPhone && (
                      <small>{errors.contactPhone.message}</small>
                    )}
                  </label>
                  <label>
                    Email *
                    <input {...register("contactEmail")} type="email" />
                    {errors.contactEmail && (
                      <small>{errors.contactEmail.message}</small>
                    )}
                  </label>
                </section>
              )}

              {step === 1 && (
                <section>
                  <h2>Способ доставки</h2>

                  <div className={styles.deliveryOptions}>
                    <button
                      type="button"
                      className={`${styles.deliveryOption} ${
                        deliveryType === "cdek" ? styles.deliveryOptionActive : ""
                      }`}
                      onClick={() => setDeliveryType("cdek")}
                    >
                      <strong>{DELIVERY_LABELS.cdek}</strong>
                      <span>от {DELIVERY_COSTS.cdek.toLocaleString("ru-RU")} ₽</span>
                      <p>Выбор пункта выдачи на карте или в списке</p>
                    </button>
                    <button
                      type="button"
                      className={`${styles.deliveryOption} ${
                        deliveryType === "custom"
                          ? styles.deliveryOptionActive
                          : ""
                      }`}
                      onClick={() => setDeliveryType("custom")}
                    >
                      <strong>{DELIVERY_LABELS.custom}</strong>
                      <span>
                        +{DELIVERY_COSTS.custom.toLocaleString("ru-RU")} ₽
                      </span>
                      <p>Курьер до указанного адреса</p>
                    </button>
                  </div>

                  <label>
                    Город *
                    <input
                      value={deliveryCity}
                      onChange={(e) => {
                        setDeliveryCity(e.target.value);
                        if (deliveryType === "cdek") {
                          setSelectedPvz(null);
                        }
                      }}
                    />
                  </label>

                  {deliveryType === "cdek" ? (
                    <div className={styles.pvzBlock}>
                      <div className={styles.pvzHead}>
                        <span>Выберите пункт выдачи СДЭК</span>
                        <button
                          type="button"
                          className={styles.linkBtn}
                          onClick={() => setPvzModalOpen(true)}
                        >
                          Выбрать пункт выдачи →
                        </button>
                      </div>
                      {selectedPvz ? (
                        <div className={styles.pvzSelected}>
                          <MapPin size={18} />
                          <div>
                            <strong>
                              {selectedPvz.code}, {selectedPvz.name}
                            </strong>
                            <p>{selectedPvz.address}</p>
                          </div>
                        </div>
                      ) : (
                        <div className={styles.pvzEmpty}>
                          <MapPin size={28} />
                          <p>Пункт выдачи не выбран</p>
                          <button
                            type="button"
                            className={styles.linkBtn}
                            onClick={() => setPvzModalOpen(true)}
                          >
                            Выбрать пункт выдачи →
                          </button>
                        </div>
                      )}
                    </div>
                  ) : (
                    <label>
                      Адрес доставки *
                      <textarea
                        rows={3}
                        value={customAddress}
                        placeholder="Улица, дом, квартира, подъезд, домофон"
                        onChange={(e) => setCustomAddress(e.target.value)}
                      />
                    </label>
                  )}

                  {deliveryError && (
                    <p className={styles.fieldError}>{deliveryError}</p>
                  )}
                </section>
              )}

              {step === 2 && (
                <section>
                  <h2>Оплата</h2>

                  <div className={styles.paymentBlock}>
                    <h3>Товары</h3>
                    <div className={styles.paymentInfo}>
                      <CreditCard size={18} />
                      <div>
                        <strong>Оплата на сайте</strong>
                        <p>
                          {itemsSubtotal.toLocaleString("ru-RU")} ₽ — банковской
                          картой при оформлении
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className={styles.paymentBlock}>
                    <h3>Доставка</h3>
                    <div className={styles.paymentChoices}>
                      <button
                        type="button"
                        className={`${styles.paymentChoice} ${
                          deliveryPayment === "online"
                            ? styles.paymentChoiceActive
                            : ""
                        }`}
                        onClick={() => setDeliveryPayment("online")}
                      >
                        <span className={styles.paymentChoiceRadio} />
                        <span>
                          <strong>На сайте</strong>
                          <small>
                            +{deliveryCost.toLocaleString("ru-RU")} ₽ к оплате
                            сейчас
                          </small>
                        </span>
                      </button>
                      <button
                        type="button"
                        className={`${styles.paymentChoice} ${
                          deliveryPayment === "on_receipt"
                            ? styles.paymentChoiceActive
                            : ""
                        }`}
                        onClick={() => setDeliveryPayment("on_receipt")}
                      >
                        <span className={styles.paymentChoiceRadio} />
                        <span>
                          <strong>При получении</strong>
                          <small>
                            {deliveryCost.toLocaleString("ru-RU")} ₽ наличными
                            или картой курьеру / в ПВЗ
                          </small>
                        </span>
                      </button>
                    </div>
                  </div>

                  <label className={styles.commentField}>
                    Комментарий к заказу
                    <textarea
                      rows={3}
                      value={comment}
                      onChange={(e) => setComment(e.target.value)}
                    />
                  </label>
                </section>
              )}

              {submitError && <div className={styles.error}>{submitError}</div>}

              <div className={styles.actions}>
                {step === 0 ? (
                  <Link to="/" className={styles.backLink}>
                    ← Вернуться в каталог
                  </Link>
                ) : (
                  <button
                    type="button"
                    className={styles.backLink}
                    onClick={() => setStep((current) => current - 1)}
                  >
                    ← Назад
                  </button>
                )}
                {step < 2 ? (
                  <button type="button" className={styles.primaryBtn} onClick={goNext}>
                    Продолжить →
                  </button>
                ) : (
                  <button
                    type="button"
                    className={styles.primaryBtn}
                    disabled={isSubmitting}
                    onClick={() => void handleSubmit(onSubmit)()}
                  >
                    {isSubmitting
                      ? "Оформление…"
                      : `Оплатить ${payNow.toLocaleString("ru-RU")} ₽`}
                  </button>
                )}
              </div>
            </div>
          </div>

          <aside className={styles.summary}>
            <h2>
              <Package size={18} /> Ваш заказ
            </h2>
            <ul className={styles.summaryList}>
              {cartItems.map((item) => (
                <li key={item.productId}>
                  <div className={styles.summaryThumb}>
                    {item.imageSrc ? (
                      <img src={item.imageSrc} alt="" />
                    ) : (
                      <div className={styles.summaryPlaceholder} />
                    )}
                  </div>
                  <div>
                    <p>{item.title}</p>
                    <span>
                      {item.quantity} × {item.price.toLocaleString("ru-RU")} ₽
                    </span>
                  </div>
                  <strong>
                    {(item.price * item.quantity).toLocaleString("ru-RU")} ₽
                  </strong>
                </li>
              ))}
            </ul>
            <div className={styles.summaryRows}>
              <div>
                <span>Товары ({itemsCount})</span>
                <strong>{itemsSubtotal.toLocaleString("ru-RU")} ₽</strong>
              </div>
              <div>
                <span>Доставка</span>
                <strong>{deliveryCost.toLocaleString("ru-RU")} ₽</strong>
              </div>
              <div className={styles.summaryHighlight}>
                <span>К оплате сейчас</span>
                <strong>{payNow.toLocaleString("ru-RU")} ₽</strong>
              </div>
              {payOnReceipt > 0 && (
                <div className={styles.summarySecondary}>
                  <span>При получении</span>
                  <strong>{payOnReceipt.toLocaleString("ru-RU")} ₽</strong>
                </div>
              )}
              <div className={styles.summaryTotal}>
                <span>Итого за заказ</span>
                <strong>{orderTotal.toLocaleString("ru-RU")} ₽</strong>
              </div>
            </div>
          </aside>
        </div>
      </main>

      <CdekPvzModal
        open={pvzModalOpen}
        city={deliveryCity}
        selectedCode={selectedPvz?.code}
        onClose={() => setPvzModalOpen(false)}
        onSelect={setSelectedPvz}
      />

      <Footer />
    </div>
  );
}
