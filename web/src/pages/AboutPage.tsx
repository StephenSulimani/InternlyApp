import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import PageShell from "../components/PageShell";
import styles from "./AboutPage.module.css";

const highlights = [
  { label: "Background", value: "Software engineer · CS graduate" },
  { label: "Built with", value: "Web scraping, aggregation, and full-stack development" },
  { label: "Focus", value: "Internships, new-grad, and early-career roles" },
  { label: "Origin", value: "A personal tool that grew into something worth sharing" },
];

export default function AboutPage() {
  return (
    <PageShell>
      <Navbar />
      <main className={styles.main}>
        <article className={styles.article}>
          <header className={styles.header}>
            <p className={styles.eyebrow}>About Internly</p>
            <h1 className={styles.title}>
              Built in the trenches,
              <em> for the trenches.</em>
            </h1>
          </header>

          <div className={styles.layout}>
            <div className={styles.prose}>
              <p className={styles.lead}>
                I&apos;m Stephen Sulimani — a software engineer and computer science
                graduate. I know the stress and anxiety that come with searching for
                internships and new-grad positions: the endless tabs, the refresh loop,
                the quiet fear that everyone else already found the listing you missed.
              </p>

              <p>
                I built Internly first for myself. I wanted one place to see
                early-career openings without switching between job sites, parsing
                group-chat links, or maintaining yet another spreadsheet. Scraping,
                aggregation, and product design are what I do — so I put those skills
                toward a problem I was living every day.
              </p>

              <p>
                Watching fellow students bounce between portals and hit refresh over
                and over made it clear this wasn&apos;t just my frustration. Internly
                is my answer: a single board for internships, new-grad programs, and
                entry-level roles — curated, deduplicated, and updated so you can
                spend less time hunting and more time applying.
              </p>

              <p>
                For more about me and my work, visit{" "}
                <a
                  href="https://suli.nyc"
                  target="_blank"
                  rel="noopener noreferrer"
                  className={styles.external}
                >
                  suli.nyc
                </a>
                .
              </p>
            </div>

            <aside className={styles.sheet} aria-label="At a glance">
              <p className={styles.sheetTitle}>At a glance</p>
              <dl className={styles.facts}>
                {highlights.map((item) => (
                  <div key={item.label} className={styles.fact}>
                    <dt>{item.label}</dt>
                    <dd>{item.value}</dd>
                  </div>
                ))}
              </dl>
            </aside>
          </div>

          <section className={styles.invite} id="access">
            <h2 className={styles.inviteTitle}>Request access</h2>
            <p className={styles.inviteBody}>
              Internly is opening up to students who would benefit from it. If that
              sounds like you, reach out from your student{" "}
              <strong>.edu</strong> email and I&apos;ll get you set up.
            </p>
            <a href="mailto:internly@suli.nyc" className={styles.email}>
              internly@suli.nyc
            </a>
          </section>
        </article>
      </main>
      <Footer />
    </PageShell>
  );
}
