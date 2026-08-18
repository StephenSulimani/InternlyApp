import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import styles from "./Navbar.module.css";

export default function Navbar() {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const { user, isAuthenticated, signOut } = useAuth();

  function handleSignOut() {
    signOut();
    navigate("/");
  }

  return (
    <header className={styles.header}>
      <nav className={styles.nav}>
        <Link to="/" className={styles.logo}>
          Internly
        </Link>
        <div className={styles.links}>
          <Link
            to="/board"
            className={`${styles.link} ${pathname === "/board" ? styles.linkActive : ""}`}
          >
            Board
          </Link>
          <Link
            to="/about"
            className={`${styles.link} ${pathname === "/about" ? styles.linkActive : ""}`}
          >
            About
          </Link>
          {isAuthenticated && user ? (
            <>
              <span className={styles.userLabel}>{user.first_name}</span>
              <button type="button" className={styles.cta} onClick={handleSignOut}>
                Sign out
              </button>
            </>
          ) : (
            <Link
              to="/login"
              className={`${styles.cta} ${pathname === "/login" ? styles.linkActive : ""}`}
            >
              Sign in
            </Link>
          )}
        </div>
      </nav>
    </header>
  );
}
