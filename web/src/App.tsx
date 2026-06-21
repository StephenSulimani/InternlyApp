import { Routes, Route } from "react-router-dom";
import AppProviders from "./providers/AppProviders";
import HomePage from "./pages/HomePage";
import AboutPage from "./pages/AboutPage";

export default function App() {
  return (
    <AppProviders>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/about" element={<AboutPage />} />
      </Routes>
    </AppProviders>
  );
}
