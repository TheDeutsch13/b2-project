import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.join(__dirname, "..");
const srcDir = path.join(root, "Carousel");
const publicDir = path.join(root, "frontend", "public", "carousel");
const generatedFile = path.join(root, "frontend", "src", "data", "carousel.generated.ts");
const link = process.argv[2] ?? "/catalog";

const imageExts = new Set([".png", ".jpg", ".jpeg", ".webp", ".gif"]);

const productLineRe =
  /^(?:product|product_id|товар|id)\s*:\s*(\d+)$/i;
const linkLineRe = /^link\s*:\s*(.+)$/i;
const productPathRe = /^\/product\/(\d+)\s*$/i;

function resolveSlideLink(line, defaultLink) {
  const productMatch =
    line.match(productLineRe) ?? line.match(productPathRe);
  if (productMatch) {
    return `/product/${productMatch[1]}`;
  }

  const linkMatch = line.match(linkLineRe);
  if (linkMatch) {
    const value = linkMatch[1].trim();
    if (value.startsWith("http://") || value.startsWith("https://")) {
      return value;
    }
    return value.startsWith("/") ? value : `/${value}`;
  }

  return null;
}

function readSlideText(txtPath, defaultLink) {
  if (!fs.existsSync(txtPath)) {
    return { title: "", description: "", link: defaultLink };
  }

  const rawLines = fs
    .readFileSync(txtPath, "utf8")
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);

  let link = defaultLink;
  const contentLines = [];

  for (const line of rawLines) {
    const resolved = resolveSlideLink(line, defaultLink);
    if (resolved) {
      link = resolved;
      continue;
    }

    contentLines.push(line);
  }

  const title = contentLines[0] ?? "";
  const description = contentLines.slice(1).join(" ").trim();

  return { title, description, link };
}

function collectSlideIds() {
  if (!fs.existsSync(srcDir)) {
    fs.mkdirSync(srcDir, { recursive: true });
    return [];
  }

  const ids = new Set();

  for (const file of fs.readdirSync(srcDir)) {
    const ext = path.extname(file).toLowerCase();
    const base = path.basename(file, ext);

    if (!/^\d+$/.test(base)) {
      continue;
    }

    if (imageExts.has(ext) || ext === ".txt") {
      ids.add(base);
    }
  }

  return [...ids].sort((a, b) => Number(a) - Number(b));
}

function findImageFile(id) {
  for (const ext of imageExts) {
    const candidate = path.join(srcDir, `${id}${ext}`);
    if (fs.existsSync(candidate)) {
      return { path: candidate, ext };
    }
  }

  return null;
}

fs.mkdirSync(publicDir, { recursive: true });

const slides = [];
const missingImages = [];

for (const id of collectSlideIds()) {
  const image = findImageFile(id);
  const { title, description, link: slideLink } = readSlideText(
    path.join(srcDir, `${id}.txt`),
    link
  );

  if (!image) {
    missingImages.push(id);
    continue;
  }

  const destName = `${id}${image.ext}`;
  fs.copyFileSync(image.path, path.join(publicDir, destName));

  slides.push({
    id,
    image: `/carousel/${destName}`,
    title: title || `Слайд ${id}`,
    description,
    link: slideLink,
  });
}

const tsContent = `/** Автогенерация: scripts/sync-carousel.mjs — не редактировать вручную */
import type { PromoSlide } from "../components/PromoCarousel/PromoCarousel";

export const carouselSlides: PromoSlide[] = ${JSON.stringify(slides, null, 2)};
`;

fs.writeFileSync(generatedFile, tsContent, "utf8");

console.log(`Carousel: ${slides.length} slide(s) synced.`);

if (missingImages.length > 0) {
  console.warn(
    `Пропущены слайды без картинки: ${missingImages.join(", ")} (нужен файл N.png рядом с N.txt)`
  );
}
