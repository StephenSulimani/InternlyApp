import { useEffect, useId, useRef } from "react";
import type { Listing } from "../lib/listings";
import SaveButton from "./SaveButton";
import styles from "./JobDetailModal.module.css";

type Props = {
  listing: Listing | null;
  onClose: () => void;
  onToggleSave?: (listing: Listing) => void;
  saveDisabled?: boolean;
};

export default function JobDetailModal({
  listing,
  onClose,
  onToggleSave,
  saveDisabled = false,
}: Props) {
  const titleId = useId();
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!listing) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    closeRef.current?.focus();

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [listing, onClose]);

  if (!listing) {
    return null;
  }

  const applyHref = listing.applicationLink?.trim();

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div
        className={styles.dialog}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onClick={(event) => event.stopPropagation()}
      >
        <button
          ref={closeRef}
          type="button"
          className={styles.close}
          onClick={onClose}
          aria-label="Close job details"
        >
          Close
        </button>

        <p className={styles.eyebrow}>
          {listing.source || "Board"}
          <span aria-hidden="true"> · </span>
          {listing.posted}
          {listing.isNew ? <span className={styles.badge}>New</span> : null}
        </p>

        <h2 id={titleId} className={styles.company}>
          {listing.company}
        </h2>
        <p className={styles.role}>{listing.role}</p>

        <dl className={styles.meta}>
          <div>
            <dt>Location</dt>
            <dd>{listing.location}</dd>
          </div>
          <div>
            <dt>Type</dt>
            <dd>{listing.type}</dd>
          </div>
          <div>
            <dt>Posted</dt>
            <dd>{listing.posted}</dd>
          </div>
        </dl>

        <p className={styles.description}>
          {listing.description ||
            "No description on file yet. Open the application page for the full posting."}
        </p>

        <div className={styles.actions}>
          {applyHref ? (
            <a
              className={styles.apply}
              href={applyHref}
              target="_blank"
              rel="noopener noreferrer"
            >
              Apply
            </a>
          ) : (
            <span className={styles.applyDisabled}>Link unavailable</span>
          )}
          {onToggleSave ? (
            <SaveButton
              saved={Boolean(listing.saved)}
              disabled={saveDisabled}
              onToggle={() => onToggleSave(listing)}
            />
          ) : null}
          <button type="button" className={styles.dismiss} onClick={onClose}>
            Keep browsing
          </button>
        </div>
      </div>
    </div>
  );
}
