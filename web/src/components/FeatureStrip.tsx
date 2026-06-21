import styles from "./FeatureStrip.module.css";

const features = [
  {
    num: "01",
    title: "Every early role, one place",
    body: "Internships, new-grad programs, and entry-level full-time — aggregated into a single index that stays current.",
    layout: "tall",
  },
  {
    num: "02",
    title: "Vector similarity search",
    body: "Describe the role you want — skills, domain, vibe — and we match listings by meaning, not keywords. Find the needle in the haystack when the perfect posting never says the words you'd think to search.",
    layout: "default",
  },
  {
    num: "03",
    title: "Filters that narrow the board",
    body: "Cut by location, role type, company, remote policy, and more. Less scrolling through noise, more time on roles you'd actually take.",
    layout: "default",
  },
  {
    num: "04",
    title: "Email & Discord alerts",
    body: "Set what you're looking for and get notified when it lands — via email or Discord. Your dream job shouldn't sit on the board for three days before you see it.",
    layout: "half",
  },
  {
    num: "05",
    title: "One board, fewer tabs",
    body: "Company portals, job boards, and midnight spreadsheets — consolidated into one place you actually check.",
    layout: "half",
  },
];

export default function FeatureStrip() {
  return (
    <section className={styles.strip}>
      <div className={styles.header}>
        <span className={styles.label}>§ Features</span>
        <h2 className={styles.heading}>
          Find it, filter it, get notified
        </h2>
      </div>
      <div className={styles.bento}>
        {features.map((f) => (
          <article
            key={f.num}
            className={`${styles.card} ${styles[f.layout]}`}
          >
            <span className={styles.num}>{f.num}</span>
            <h3 className={styles.title}>{f.title}</h3>
            <p className={styles.body}>{f.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
