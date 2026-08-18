import { Link } from "react-router-dom";
import { HeroStack } from "./StatusBar";
import { useAuth } from "../providers/AuthProvider";
import styles from "./Hero.module.css";

export default function Hero() {
  const { isAuthenticated } = useAuth();

  return (
    <section className={styles.hero}>
      <div className={styles.layout}>
        <div className={styles.copy}>
          <h1 className={styles.headline}>
            The classifieds,
            <br />
            <em>reclassified.</em>
          </h1>

          <p className={styles.subhead}>
            Internships, new-grad roles, and early-career openings — one board,
            deduplicated, indexed, and kept current.
          </p>

          <div className={styles.actions}>
            <Link to="/board" className={styles.primary}>
              Open the board
            </Link>
            {!isAuthenticated && (
              <Link to="/about#access" className={styles.secondary}>
                Request access
              </Link>
            )}
          </div>

          <p className={styles.marginalia}>Est. 2026</p>
        </div>

        <div className={styles.visual}>
          <HeroStack />
        </div>
      </div>
    </section>
  );
}
