import { useJobStats } from "../hooks/useJobStats";
import { useBoardPreview } from "../hooks/useBoardPreview";
import styles from "./StatusBar.module.css";

function formatCount(value: number | undefined, loading: boolean): string {
  if (loading || value === undefined) {
    return "—";
  }
  return value.toLocaleString("en-US");
}

function formatUpdatedLabel(iso: string | undefined): string {
  if (!iso) {
    return "Summer 2026 · Updated daily";
  }

  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) {
    return "Summer 2026 · Updated daily";
  }

  const diffMs = Date.now() - then;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) {
    return "Summer 2026 · Updated just now";
  }
  if (minutes < 60) {
    return `Summer 2026 · Updated ${minutes}m ago`;
  }

  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `Summer 2026 · Updated ${hours}h ago`;
  }

  const days = Math.floor(hours / 24);
  if (days === 1) {
    return "Summer 2026 · Updated yesterday";
  }
  return `Summer 2026 · Updated ${days}d ago`;
}

export default function StatusBar() {
  const { data, isPending, isError } = useJobStats();
  const loading = isPending && !data;

  const stats = [
    { label: "Indexed", value: formatCount(data?.total_jobs, loading) },
    { label: "This week", value: formatCount(data?.added_this_week, loading) },
    { label: "Companies", value: formatCount(data?.total_companies, loading) },
  ];

  return (
    <div className={styles.bar} role="status" aria-label="Board statistics">
      <div className={styles.live}>
        <span className={styles.pulse} aria-hidden="true" />
        <span className={styles.liveLabel}>
          {isError ? "Board offline" : "Board live"}
        </span>
      </div>
      <div className={styles.stats}>
        {stats.map((s) => (
          <div key={s.label} className={styles.stat}>
            <span className={styles.statValue}>{s.value}</span>
            <span className={styles.statLabel}>{s.label}</span>
          </div>
        ))}
      </div>
      <p className={styles.season}>
        {isError
          ? "Summer 2026 · Stats unavailable"
          : formatUpdatedLabel(data?.last_updated)}
      </p>
    </div>
  );
}

export function HeroStack() {
  const { data: listings = [], isPending } = useBoardPreview();
  const preview = listings.slice(0, 3);
  const offsets = [styles.cardA, styles.cardB, styles.cardC];

  if (isPending && preview.length === 0) {
    return (
      <div className={styles.stack} aria-hidden="true">
        {[0, 1, 2].map((i) => (
          <div key={i} className={`${styles.miniCard} ${offsets[i]}`}>
            <span className={styles.miniType}>—</span>
            <p className={styles.miniCompany}>Loading</p>
            <p className={styles.miniRole}>Fetching listings…</p>
            <span className={styles.miniLoc}>—</span>
          </div>
        ))}
      </div>
    );
  }

  if (preview.length === 0) {
    return (
      <div className={styles.stack} aria-hidden="true">
        <div className={`${styles.miniCard} ${styles.cardA}`}>
          <span className={styles.miniType}>Board</span>
          <p className={styles.miniCompany}>Internly</p>
          <p className={styles.miniRole}>Listings appear here as roles are indexed</p>
          <span className={styles.miniLoc}>Coming soon</span>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.stack} aria-hidden="true">
      {preview.map((job, i) => (
        <div key={job.id} className={`${styles.miniCard} ${offsets[i]}`}>
          <span className={styles.miniType}>{job.type}</span>
          <p className={styles.miniCompany}>{job.company}</p>
          <p className={styles.miniRole}>{job.role}</p>
          <span className={styles.miniLoc}>{job.location}</span>
        </div>
      ))}
    </div>
  );
}
