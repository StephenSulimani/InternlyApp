import type { Listing } from "../lib/listings";
import styles from "./ListingCard.module.css";

type Props = {
  listing: Listing;
  index: number;
};

export default function ListingCard({ listing, index }: Props) {
  return (
    <article className={styles.row}>
      <span className={styles.index}>{String(index).padStart(2, "0")}</span>
      <div className={styles.main}>
        {listing.isNew && <span className={styles.badge}>New</span>}
        <h3 className={styles.company}>{listing.company}</h3>
        <p className={styles.role}>{listing.role}</p>
      </div>
      <span className={styles.location}>{listing.location}</span>
      <span className={styles.type}>{listing.type}</span>
      <span className={styles.posted}>{listing.posted}</span>
    </article>
  );
}
