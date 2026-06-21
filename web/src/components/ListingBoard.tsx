import { mockListings } from "../data/mockListings";
import ListingCard from "./ListingCard";
import styles from "./ListingBoard.module.css";

export default function ListingBoard() {
  return (
    <section className={styles.board} id="board">
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <span className={styles.label}>§ The board</span>
          <h2 className={styles.title}>Today&apos;s index</h2>
        </div>
        <div className={styles.headerRight}>
          <span className={styles.count}>{mockListings.length} shown</span>
          <span className={styles.hint}>Sign in for search &amp; filters</span>
        </div>
      </header>

      <div className={styles.catalog}>
        <div className={styles.catalogHead} aria-hidden="true">
          <span>#</span>
          <span>Company / Role</span>
          <span>Location</span>
          <span>Type</span>
          <span>Posted</span>
        </div>
        {mockListings.map((listing, index) => (
          <ListingCard key={listing.id} listing={listing} index={index + 1} />
        ))}
      </div>
    </section>
  );
}
