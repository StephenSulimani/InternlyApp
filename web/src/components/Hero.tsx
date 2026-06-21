import { Link } from "react-router-dom";
import styles from "./Hero.module.css";

export default function Hero() {
  return (
    <section className={styles.hero}>
      <div className={styles.meta}>
        <span className={styles.issue}>Summer 2026 season</span>
        <span className={styles.date}>Updated daily</span>
      </div>

      <h1 className={styles.headline}>
        The classifieds,
        <em> reclassified.</em>
      </h1>

      <p className={styles.subhead}>
        Internly brings internships, new-grad roles, and early-career openings
        into one board — curated, deduplicated, and ready before your coffee
        cools.
      </p>

      <div className={styles.actions}>
        <a href="#board" className={styles.primary}>
          Browse the board
        </a>
        <Link to="/about#access" className={styles.secondary}>
          Request access
        </Link>
      </div>

      <blockquote className={styles.pullquote}>
        <p>
          &ldquo;I was juggling four tabs and a spreadsheet. Now I check one
          board.&rdquo;
        </p>
        <footer>— early-career candidate, Class of 2026</footer>
      </blockquote>
    </section>
  );
}
