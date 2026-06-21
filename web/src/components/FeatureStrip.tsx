import styles from "./FeatureStrip.module.css";

const features = [
  {
    num: "01",
    title: "Every early role, one place",
    body: "Internships, new-grad programs, and entry-level full-time roles — aggregated into a single board that stays current.",
  },
  {
    num: "02",
    title: "One board, fewer tabs",
    body: "Stop alt-tabbing between company portals, job boards, and the spreadsheet someone shared at midnight.",
  },
  {
    num: "03",
    title: "Built for the search",
    body: "Filters, saved roles, and alerts designed for people who take the early-career hunt seriously.",
  },
];

export default function FeatureStrip() {
  return (
    <section className={styles.strip}>
      <div className={styles.inner}>
        {features.map((f) => (
          <article key={f.num} className={styles.card}>
            <span className={styles.num}>{f.num}</span>
            <h2 className={styles.title}>{f.title}</h2>
            <p className={styles.body}>{f.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
