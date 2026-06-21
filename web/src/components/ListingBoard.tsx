import { mockListings } from "../data/mockListings";
import ListingCard from "./ListingCard";
import styles from "./ListingBoard.module.css";

export default function ListingBoard() {
  return (
    <section className={styles.board} id="board">
      <div className={styles.header}>
        <div>
          <p className={styles.eyebrow}>Today on Internly</p>
          <h2 className={styles.title}>The board</h2>
        </div>
        <p className={styles.note}>
          A sample of what&apos;s live right now — internships, new-grad roles,
          and entry-level openings across tech and finance.
        </p>
      </div>

      <div className={styles.grid}>
        {mockListings.map((listing) => (
          <ListingCard key={listing.id} listing={listing} />
        ))}
      </div>

      <p className={styles.disclaimer}>
        <span className={styles.arrow}>→</span>
        Search, filters, and saved roles available when you sign in.
      </p>
    </section>
  );
}
