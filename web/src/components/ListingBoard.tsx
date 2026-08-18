import { useCallback, useState } from "react";
import { Link } from "react-router-dom";
import ListingCard from "./ListingCard";
import JobDetailModal from "./JobDetailModal";
import { useBoardPreview } from "../hooks/useBoardPreview";
import type { Listing } from "../lib/listings";
import styles from "./ListingBoard.module.css";

export default function ListingBoard() {
  const { data: listings = [], isPending, isError } = useBoardPreview();
  const [selected, setSelected] = useState<Listing | null>(null);
  const closeModal = useCallback(() => setSelected(null), []);

  return (
    <section className={styles.board} id="board">
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <span className={styles.label}>§ The board</span>
          <h2 className={styles.title}>Today&apos;s index</h2>
        </div>
        <div className={styles.headerRight}>
          <span className={styles.count}>
            {isPending ? "Loading…" : `${listings.length} shown`}
          </span>
          <Link to="/board" className={styles.hint}>
            Open the full board
          </Link>
        </div>
      </header>

      <div className={styles.catalog}>
        <div className={styles.catalogHead} aria-hidden="true">
          <span>#</span>
          <span>Company / Role</span>
          <span>Location</span>
          <span>Type</span>
          <span>Posted</span>
          <span />
        </div>

        {isPending && (
          <p className={styles.empty}>Loading latest listings…</p>
        )}

        {isError && (
          <p className={styles.empty}>Couldn’t load the board preview.</p>
        )}

        {!isPending && !isError && listings.length === 0 && (
          <p className={styles.empty}>No listings yet — check back soon.</p>
        )}

        {listings.map((listing, index) => (
          <ListingCard
            key={listing.id}
            listing={listing}
            index={index + 1}
            selected={selected?.id === listing.id}
            onSelect={setSelected}
          />
        ))}
      </div>

      <JobDetailModal listing={selected} onClose={closeModal} />
    </section>
  );
}
