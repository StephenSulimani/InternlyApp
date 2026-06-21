import type { Listing } from "../data/mockListings";
import styles from "./ListingCard.module.css";

type Props = {
  listing: Listing;
};

export default function ListingCard({ listing }: Props) {
  return (
    <article className={styles.card}>
      {listing.isNew && <span className={styles.stamp}>New</span>}
      <header className={styles.header}>
        <h3 className={styles.company}>{listing.company}</h3>
        <span className={styles.posted}>{listing.posted}</span>
      </header>
      <p className={styles.role}>{listing.role}</p>
      <footer className={styles.footer}>
        <span>{listing.location}</span>
        <span className={styles.divider}>·</span>
        <span className={styles.type}>{listing.type}</span>
      </footer>
    </article>
  );
}
