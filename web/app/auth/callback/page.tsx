"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { setToken } from "@/lib/api";

function Callback() {
  const router = useRouter();
  const params = useSearchParams();
  useEffect(() => {
    const token = params.get("token");
    if (token) {
      setToken(token);
      router.replace("/dashboard");
    } else {
      router.replace("/login");
    }
  }, [params, router]);
  return <div className="container">Signing you in…</div>;
}

export default function Page() {
  return (
    <Suspense fallback={<div className="container">Signing you in…</div>}>
      <Callback />
    </Suspense>
  );
}
