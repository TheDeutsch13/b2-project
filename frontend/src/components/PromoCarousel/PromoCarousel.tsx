import { ChevronLeft, ChevronRight } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import styles from "./PromoCarousel.module.css";

export interface PromoSlide {
  id: string;
  image: string;
  title: string;
  description: string;
  link: string;
}

interface PromoCarouselProps {
  slides: readonly PromoSlide[];
  autoPlayMs?: number;
}

export function PromoCarousel({ slides, autoPlayMs = 6000 }: PromoCarouselProps) {
  const [index, setIndex] = useState(0);
  const [paused, setPaused] = useState(false);

  const goTo = useCallback(
    (next: number) => {
      const total = slides.length;
      setIndex(((next % total) + total) % total);
    },
    [slides.length]
  );

  const goPrev = () => goTo(index - 1);
  const goNext = () => goTo(index + 1);

  useEffect(() => {
    if (paused || slides.length <= 1) {
      return;
    }

    const timer = window.setInterval(() => {
      setIndex((current) => (current + 1) % slides.length);
    }, autoPlayMs);

    return () => window.clearInterval(timer);
  }, [autoPlayMs, paused, slides.length]);

  if (slides.length === 0) {
    return null;
  }

  const isExternalLink = (href: string) =>
    href.startsWith("http://") || href.startsWith("https://");

  return (
    <section
      className={styles.carousel}
      aria-roledescription="carousel"
      aria-label="Новинки"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <div className={styles.viewport}>
        {slides.map((item, slideIndex) => (
          <article
            key={item.id}
            className={`${styles.slide} ${slideIndex === index ? styles.slideActive : ""}`}
            aria-hidden={slideIndex !== index}
          >
            {isExternalLink(item.link) ? (
              <a
                href={item.link}
                className={styles.slideLink}
                aria-label={`${item.title}. ${item.description}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                <img src={item.image} alt="" className={styles.slideImage} />
                <div className={styles.overlay} />
                <div className={styles.content}>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                </div>
              </a>
            ) : (
              <Link
                to={item.link}
                className={styles.slideLink}
                aria-label={`${item.title}. ${item.description}`}
              >
                <img src={item.image} alt="" className={styles.slideImage} />
                <div className={styles.overlay} />
                <div className={styles.content}>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                </div>
              </Link>
            )}
          </article>
        ))}

        {slides.length > 1 && (
          <>
            <button
              type="button"
              className={`${styles.navBtn} ${styles.navPrev}`}
              onClick={goPrev}
              aria-label="Предыдущий слайд"
            >
              <ChevronLeft size={22} />
            </button>
            <button
              type="button"
              className={`${styles.navBtn} ${styles.navNext}`}
              onClick={goNext}
              aria-label="Следующий слайд"
            >
              <ChevronRight size={22} />
            </button>
          </>
        )}
      </div>

      {slides.length > 1 && (
        <div className={styles.dots} role="tablist" aria-label="Слайды новинок">
          {slides.map((item, dotIndex) => (
            <button
              key={item.id}
              type="button"
              role="tab"
              aria-selected={dotIndex === index}
              aria-label={`Слайд ${dotIndex + 1}`}
              className={`${styles.dot} ${dotIndex === index ? styles.dotActive : ""}`}
              onClick={() => goTo(dotIndex)}
            />
          ))}
        </div>
      )}
    </section>
  );
}
