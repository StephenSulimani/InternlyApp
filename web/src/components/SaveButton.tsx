import styles from "./SaveButton.module.css";

type Props = {
  saved: boolean;
  onToggle: () => void;
  disabled?: boolean;
};

export default function SaveButton({ saved, onToggle, disabled = false }: Props) {
  return (
    <button
      type="button"
      className={saved ? `${styles.save} ${styles.saved}` : styles.save}
      aria-pressed={saved}
      aria-label={saved ? "Remove from saved jobs" : "Save job"}
      disabled={disabled}
      onClick={(event) => {
        event.stopPropagation();
        onToggle();
      }}
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          d="M12.1 21.35 10.55 19.94C5.4 15.27 2 12.2 2 8.5 2 5.5 4.42 3.1 7.4 3.1c1.74 0 3.41.81 4.5 2.09 1.09-1.28 2.76-2.09 4.5-2.09C19.58 3.1 22 5.5 22 8.5c0 3.7-3.4 6.77-8.55 11.44z"
          fill={saved ? "currentColor" : "none"}
          stroke="currentColor"
          strokeWidth="1.7"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  );
}
