import { Link } from "react-router-dom";
import {
  DiscordIcon,
  TelegramIcon,
  VkIcon,
  YoutubeIcon,
} from "./SocialIcons";
import styles from "./Footer.module.css";

const catalogLinks = ["Мыши", "Коврики", "Клавиатуры", "Аксессуары"];
const supportLinks = ["Контакты", "FAQ"];
const buyerLinks = ["Доставка и оплата", "Гарантия", "Возврат", "О компании"];

const socialLinks = [
  {
    label: "YouTube",
    href: "https://www.youtube.com",
    icon: YoutubeIcon,
  },
  {
    label: "ВКонтакте",
    href: "https://vk.com",
    icon: VkIcon,
  },
  {
    label: "Telegram",
    href: "https://t.me",
    icon: TelegramIcon,
  },
  {
    label: "Discord",
    href: "https://discord.com",
    icon: DiscordIcon,
  },
] as const;

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={`container ${styles.inner}`}>
        <div className={styles.brand}>
          <Link className="logo" to="/">
            GAME<span>GEAR</span>
          </Link>
          <p className={styles.description}>Премиальная игровая периферия</p>

          <div className={styles.social}>
            {socialLinks.map(({ label, href, icon: Icon }) => (
              <a
                key={label}
                href={href}
                className={styles.socialLink}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={label}
                title={label}
              >
                <Icon />
              </a>
            ))}
          </div>
        </div>

        <FooterColumn title="Каталог" links={catalogLinks} />
        <FooterColumn title="Поддержка" links={supportLinks} />
        <FooterColumn title="Покупателям" links={buyerLinks} />
      </div>
    </footer>
  );
}

interface FooterColumnProps {
  title: string;
  links: string[];
}

function FooterColumn({ title, links }: FooterColumnProps) {
  return (
    <div className={styles.column}>
      <h3>{title}</h3>

      <nav>
        {links.map((link) => (
          <a href="#" key={link}>
            {link}
          </a>
        ))}
      </nav>
    </div>
  );
}