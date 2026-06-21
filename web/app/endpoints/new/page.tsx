"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { EndpointForm, EndpointFormValues, emptyEndpoint } from "@/components/EndpointForm";
import { api, getToken, type EndpointDetail } from "@/lib/api";

export default function NewEndpointPage() {
  const router = useRouter();

  useEffect(() => {
    if (!getToken()) router.replace("/login");
  }, [router]);

  async function create(v: EndpointFormValues) {
    const res = await api<EndpointDetail>("/endpoints", { method: "POST", body: JSON.stringify(v) });
    router.push(`/endpoints/${res.endpoint.id}`);
  }

  return (
    <>
      <Nav />
      <div className="container">
        <div style={{ marginBottom: "0.75rem" }}>
          <Link href="/endpoints" className="muted">← Endpoints</Link>
        </div>
        <div className="page-head">
          <div>
            <h1>New endpoint</h1>
            <div className="subtitle">Configure the request and what makes a check pass.</div>
          </div>
        </div>
        <EndpointForm initial={emptyEndpoint()} submitLabel="Create endpoint" onSubmit={create} onCancel={() => router.push("/endpoints")} />
      </div>
    </>
  );
}
