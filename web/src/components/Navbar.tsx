import { Link, useLocation } from "react-router-dom";
import styles from "./Navbar.module.css";

export default function Navbar() {
  const { pathname } = useLocation();

  return (
    <header className={styles.header}>
      <nav className={styles.nav}>
        <Link to="/" className={styles.logo}>
          Internly
        </Link>
        <div className={styles.links}>
          <Link
            to="/#board"
            className={`${styles.link} ${pathname === "/" ? "" : styles.linkMuted}`}
          >
            Board
          </Link>
          <Link
            to="/about"
            className={`${styles.link} ${pathname === "/about" ? styles.linkActive : ""}`}
          >
            About
          </Link>
          <button type="button" className={styles.cta}>
            Sign in
          </button>
        </div>
      </nav>
    </header>
  );
}
