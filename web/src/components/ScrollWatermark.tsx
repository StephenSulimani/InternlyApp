import { useEffect, useState } from "react";
import styles from "./ScrollWatermark.module.css";

export default function ScrollWatermark() {
  const [offset, setOffset] = useState(0);

  useEffect(() => {
    const onScroll = () => setOffset(window.scrollY * 0.4);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <div className={styles.wrap} aria-hidden="true">
      <span
        className={styles.mark}
        style={{ transform: `translate(-50%, ${offset}px)` }}
      >
        Internly
      </span>
    </div>
  );
}
