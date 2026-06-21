import { tickerMessages } from "../data/mockListings";
import styles from "./Ticker.module.css";

export default function Ticker() {
  const items = [...tickerMessages, ...tickerMessages];

  return (
    <div className={styles.wrap} aria-hidden="true">
      <div className={styles.track}>
        {items.map((msg, i) => (
          <span key={i} className={styles.item}>
            <span className={styles.dot} />
            {msg}
          </span>
        ))}
      </div>
    </div>
  );
}
