/**
 * Массовое создание тестовых товаров через API (нужен admin).
 *
 * Пример:
 *   set SEED_EMAIL=admin@test.com
 *   set SEED_PASSWORD=yourpassword
 *   node scripts/seed-products.mjs
 *
 * Опции (env):
 *   AUTH_URL=http://localhost:8081
 *   PRODUCT_URL=http://localhost:8082
 *   SEED_PER_CATEGORY=8   — товаров на каждую категорию
 */

const AUTH_URL = process.env.AUTH_URL ?? "http://localhost:8081";
const PRODUCT_URL = process.env.PRODUCT_URL ?? "http://localhost:8082";
const EMAIL = process.env.SEED_EMAIL ?? "";
const PASSWORD = process.env.SEED_PASSWORD ?? "";
const PER_CATEGORY = Number(process.env.SEED_PER_CATEGORY ?? 8);

const BRANDS = ["Logitech", "Razer", "Pulsar", "Finalmouse", "Vaxee", "Zowie"];
const SENSORS = ["HERO 2", "PAW3395", "PAW3311", "Focus Pro 30K"];
const WEIGHTS = ["52", "58", "63", "72", "49"];
const CONNECTIONS = ["Проводная", "Беспроводная"];
const POLL_RATES = ["125", "500", "1000", "8000"];
const MOUSE_MATERIALS = ["Пластик", "Магний", "Алюминий"];
const ENCODERS = ["TTC Gold", "Kailh", "Huano"];
const MOUSE_SWITCHES = ["Omron", "Kailh", "Huano"];
const PAD_SIZES_MM = ["400x450", "490x420", "900x400", "1200x600"];
const PAD_SURFACE = ["Speed", "Control", "Balance"];
const SWITCHES = ["Linear", "Tactile", "Magnetic (HE)"];
const FORM_FACTORS = ["60%", "65%", "70%", "75%", "TKL"];
const ACCESSORY_TYPES = ["Глайды", "Грипсы", "Рукава", "Кабель"];

function pick(list, index) {
  return list[index % list.length];
}

async function request(url, options = {}) {
  const response = await fetch(url, options);
  const text = await response.text();
  let data = null;

  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }

  if (!response.ok) {
    throw new Error(
      `${options.method ?? "GET"} ${url} → ${response.status}: ${JSON.stringify(data)}`
    );
  }

  return data;
}

async function login() {
  if (!EMAIL || !PASSWORD) {
    throw new Error("Задайте SEED_EMAIL и SEED_PASSWORD (аккаунт с ролью admin)");
  }

  const data = await request(`${AUTH_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: EMAIL, password: PASSWORD }),
  });

  return data.access_token;
}

function buildProduct(category, index) {
  const brand = pick(BRANDS, index);
  const basePrice = 3000 + index * 850 + category.id * 200;

  let specifications = [];
  let title = "";

  switch (category.name) {
    case "Мыши":
      title = `${brand} Mouse Pro ${index + 1}`;
      specifications = [
        { label: "Сенсор", value: pick(SENSORS, index) },
        { label: "Вес", value: pick(WEIGHTS, index) },
        { label: "Подключение", value: pick(CONNECTIONS, index) },
        { label: "Частота опроса", value: pick(POLL_RATES, index) },
        { label: "Материал", value: pick(MOUSE_MATERIALS, index) },
        { label: "Энкодер", value: pick(ENCODERS, index) },
        { label: "Переключатели", value: pick(MOUSE_SWITCHES, index) },
      ];
      break;
    case "Коврики":
      title = `${brand} Pad ${index + 1}`;
      specifications = [
        { label: "Размер", value: pick(PAD_SIZES_MM, index) },
        { label: "Толщина", value: pick(["3 мм", "4 мм", "5 мм"], index) },
        { label: "Поверхность", value: pick(PAD_SURFACE, index) },
        { label: "Материал", value: pick(["Ткань", "Стекло", "Гибрид"], index) },
      ];
      break;
    case "Клавиатуры":
      title = `${brand} Keyboard ${pick(FORM_FACTORS, index)}`;
      specifications = [
        { label: "Переключатели", value: pick(SWITCHES, index) },
        { label: "Формфактор", value: pick(FORM_FACTORS, index) },
        { label: "Подключение", value: pick(CONNECTIONS, index) },
        { label: "Раскладка", value: pick(["ANSI", "ISO"], index) },
      ];
      break;
    case "Аксессуары":
      title = `${brand} ${pick(ACCESSORY_TYPES, index)} Kit ${index + 1}`;
      specifications = [
        { label: "Тип", value: pick(ACCESSORY_TYPES, index) },
        { label: "Материал", value: pick(["PTFE", "Силикон", "Ткань"], index) },
      ];
      break;
    default:
      title = `${brand} Device ${index + 1}`;
  }

  return {
    title,
    description: `Тестовый товар для проверки каталога, поиска и фильтров (#${index + 1}).`,
    price: basePrice,
    category_id: category.id,
    brand,
    stock: 5 + (index % 20),
    images: [],
    variants: ["Стандарт"],
    specifications,
    reviews: [],
  };
}

async function main() {
  console.log("GAMEGEAR seed-products");
  console.log(`Auth: ${AUTH_URL}, Products: ${PRODUCT_URL}`);
  console.log(`По ${PER_CATEGORY} товаров на категорию\n`);

  const token = await login();
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };

  const categories = await request(`${PRODUCT_URL}/api/categories`);
  const targetCategories = categories.filter((c) =>
    ["Мыши", "Коврики", "Клавиатуры", "Аксессуары"].includes(c.name)
  );

  if (targetCategories.length === 0) {
    throw new Error("Категории не найдены. Запустите product-service и миграции.");
  }

  let created = 0;

  for (const category of targetCategories) {
    for (let i = 0; i < PER_CATEGORY; i += 1) {
      const payload = buildProduct(category, i);
      await request(`${PRODUCT_URL}/api/products`, {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
      });
      created += 1;
      process.stdout.write(`  + ${payload.title}\n`);
    }
  }

  console.log(`\nГотово: создано ${created} товаров.`);
  console.log("Откройте каталог и проверьте фильтры и поиск в шапке.");
}

main().catch((error) => {
  console.error("\nОшибка:", error.message);
  process.exit(1);
});
