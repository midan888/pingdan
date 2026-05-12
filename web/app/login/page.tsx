"use client";

import { API_URL } from "@/lib/api";

export default function LoginPage() {
  return (
    <div className="container">
      <h1>pingdan</h1>
      <p>Sign in to monitor your HTTP endpoints.</p>
      <div className="row" style={{ marginTop: "1.5rem" }}>
        <a href={`${API_URL}/auth/google/start`}>
          <button>Continue with Google</button>
        </a>
        <a href={`${API_URL}/auth/github/start`}>
          <button>Continue with GitHub</button>
        </a>
      </div>
    </div>
  );
}
