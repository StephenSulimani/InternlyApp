import { type FormEvent, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";
import PageShell from "../components/PageShell";
import { useLogin } from "../hooks/useLogin";
import { useAuth } from "../providers/AuthProvider";
import styles from "./LoginPage.module.css";

export default function LoginPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const login = useLogin();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/", { replace: true });
    }
  }, [isAuthenticated, navigate]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);

    try {
      await login.mutateAsync({ email, password });
      navigate("/");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Unable to sign in";
      setError(message);
    }
  }

  return (
    <PageShell>
      <Navbar />
      <main className={styles.main}>
        <form className={styles.form} onSubmit={handleSubmit}>
          <p className={styles.eyebrow}>Sign in</p>
          <h1 className={styles.title}>Welcome back</h1>
          <p className={styles.subtitle}>
            Access the board, saved roles, and alerts.
          </p>

          {error && <p className={styles.error}>{error}</p>}

          <label className={styles.field}>
            <span>Email</span>
            <input
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>

          <label className={styles.field}>
            <span>Password</span>
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>

          <button
            type="submit"
            className={styles.submit}
            disabled={login.isPending}
          >
            {login.isPending ? "Signing in…" : "Sign in"}
          </button>

          <p className={styles.footer}>
            Need access?{" "}
            <Link to="/about#access">Request an account</Link>
          </p>
        </form>
      </main>
      <Footer />
    </PageShell>
  );
}
