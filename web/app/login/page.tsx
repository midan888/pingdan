"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { API_URL, setToken } from "@/lib/api";
import { AuthAside } from "@/components/AuthAside";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/auth/email/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) {
        setError((await res.text()) || "Login failed");
        return;
      }
      const { token } = await res.json();
      setToken(token);
      router.replace("/dashboard");
    } catch {
      setError("Network error — please try again");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-split">
      <AuthAside />
      <main className="auth-main">
        <div className="auth-card">
          <h1>Welcome back</h1>
          <p className="sub">Sign in to your monitoring dashboard.</p>

          <div className="oauth-grid">
            <a href={`${API_URL}/auth/google/start`}><button type="button" className="oauth-btn">Google</button></a>
          </div>

          <div className="divider">or continue with email</div>

          <form onSubmit={onSubmit}>
            <div className="field">
              <label>Email</label>
              <input type="email" placeholder="you@company.com" value={email} onChange={(e) => setEmail(e.target.value)} required autoFocus />
            </div>
            <div className="field">
              <label>Password</label>
              <input type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} required />
            </div>
            {error && <p className="error-text">{error}</p>}
            <button type="submit" className="primary" style={{ width: "100%", marginTop: "0.5rem" }} disabled={loading}>
              {loading ? "Signing in…" : "Sign in"}
            </button>
          </form>

          <p className="auth-foot">
            New to pingdan? <Link href="/register">Create an account</Link>
          </p>
        </div>
      </main>
    </div>
  );
}
