import Navbar from "../components/Navbar";
import Ticker from "../components/Ticker";
import Hero from "../components/Hero";
import FeatureStrip from "../components/FeatureStrip";
import ListingBoard from "../components/ListingBoard";
import Footer from "../components/Footer";
import styles from "./HomePage.module.css";

export default function HomePage() {
  return (
    <div className={styles.page}>
      <Navbar />
      <Ticker />
      <main>
        <Hero />
        <FeatureStrip />
        <ListingBoard />
      </main>
      <Footer />
    </div>
  );
}
