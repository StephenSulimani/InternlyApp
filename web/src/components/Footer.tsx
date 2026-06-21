import { Link } from "react-router-dom";
import styles from "./Footer.module.css";

export default function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <p className={styles.wordmark}>Internly</p>
        <p className={styles.tagline}>
          Early-career roles · One board
        </p>
        <nav className={styles.nav}>
          <Link to="/about">About</Link>
          <a href="mailto:internly@suli.nyc">Contact</a>
        </nav>
        <p className={styles.credit}>© {new Date().getFullYear()} Internly</p>
      </div>
    </footer>
  );
}
