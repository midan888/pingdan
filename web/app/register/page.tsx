"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { API_URL, setToken } from "@/lib/api";
import { AuthAside } from "@/components/AuthAside";

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/auth/email/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, name }),
      });
      if (!res.ok) {
        setError((await res.text()) || "Registration failed");
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
          <h1>Create your account</h1>
          <p className="sub">Start monitoring in under a minute. No credit card required.</p>

          <div className="oauth-grid">
            <a href={`${API_URL}/auth/google/start`}><button type="button" className="oauth-btn">Google</button></a>
          </div>

          <div className="divider">or sign up with email</div>

          <form onSubmit={onSubmit}>
            <div className="field">
              <label>Name <span className="faint">(optional)</span></label>
              <input type="text" placeholder="Ada Lovelace" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="field">
              <label>Email</label>
              <input type="email" placeholder="you@company.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
            </div>
            <div className="field">
              <label>Password</label>
              <input type="password" placeholder="At least 8 characters" value={password} onChange={(e) => setPassword(e.target.value)} minLength={8} required />
            </div>
            {error && <p className="error-text">{error}</p>}
            <button type="submit" className="primary" style={{ width: "100%", marginTop: "0.5rem" }} disabled={loading}>
              {loading ? "Creating account…" : "Create account"}
            </button>
          </form>

          <p className="auth-foot">
            Already have an account? <Link href="/login">Sign in</Link>
          </p>
        </div>
      </main>
    </div>
  );
}
