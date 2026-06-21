import { mockListings } from "../data/mockListings";
import styles from "./StatusBar.module.css";

const stats = [
  { label: "Indexed", value: "1,247" },
  { label: "Added today", value: "42" },
  { label: "Remote", value: "340" },
  { label: "Role types", value: "3" },
];

export default function StatusBar() {
  return (
    <div className={styles.bar} role="status" aria-label="Board statistics">
      <div className={styles.live}>
        <span className={styles.pulse} aria-hidden="true" />
        <span className={styles.liveLabel}>Board live</span>
      </div>
      <div className={styles.stats}>
        {stats.map((s) => (
          <div key={s.label} className={styles.stat}>
            <span className={styles.statValue}>{s.value}</span>
            <span className={styles.statLabel}>{s.label}</span>
          </div>
        ))}
      </div>
      <p className={styles.season}>Summer 2026 · Updated daily</p>
    </div>
  );
}

export function HeroStack() {
  const preview = mockListings.slice(0, 3);
  const offsets = [styles.cardA, styles.cardB, styles.cardC];

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
