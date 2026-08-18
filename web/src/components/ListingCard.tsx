import type { Listing } from "../lib/listings";
import styles from "./ListingCard.module.css";

type Props = {
  listing: Listing;
  index: number;
  selected?: boolean;
  onSelect?: (listing: Listing) => void;
};

export default function ListingCard({
  listing,
  index,
  selected = false,
  onSelect,
}: Props) {
  const interactive = Boolean(onSelect);
  const className = `${styles.row} ${selected ? styles.selected : ""} ${
    interactive ? styles.interactive : ""
  }`;

  const content = (
    <>
      <span className={styles.index}>{String(index).padStart(2, "0")}</span>
      <div className={styles.main}>
        {listing.isNew && <span className={styles.badge}>New</span>}
        <h3 className={styles.company}>{listing.company}</h3>
        <p className={styles.role}>{listing.role}</p>
      </div>
      <span className={styles.location}>{listing.location}</span>
      <span className={styles.type}>{listing.type}</span>
      <span className={styles.posted}>{listing.posted}</span>
      {interactive && (
        <span className={styles.hint} aria-hidden="true">
          View
        </span>
      )}
    </>
  );

  if (!onSelect) {
    return <article className={className}>{content}</article>;
  }

  return (
    <article>
      <button
        type="button"
        className={className}
        onClick={() => onSelect(listing)}
        aria-pressed={selected}
        aria-haspopup="dialog"
      >
        {content}
      </button>
    </article>
  );
}
