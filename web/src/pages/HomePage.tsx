import Navbar from "../components/Navbar";
import StatusBar from "../components/StatusBar";
import Hero from "../components/Hero";
import FeatureStrip from "../components/FeatureStrip";
import ListingBoard from "../components/ListingBoard";
import Footer from "../components/Footer";
import PageShell from "../components/PageShell";
import styles from "./HomePage.module.css";

export default function HomePage() {
  return (
    <PageShell>
      <Navbar />
      <StatusBar />
      <main className={styles.main}>
        <Hero />
        <FeatureStrip />
        <ListingBoard />
      </main>
      <Footer />
    </PageShell>
  );
}
