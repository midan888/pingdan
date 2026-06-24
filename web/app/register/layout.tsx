import type { Metadata } from "next";
import type { ReactNode } from "react";

const description =
  "Create a free pingdan account and start monitoring HTTP endpoints in a minute — deep assertions, response-time charts, and instant alerts. No credit card required.";

export const metadata: Metadata = {
  title: "Create your free account — pingdan",
  description,
  alternates: { canonical: "/register" },
  openGraph: {
    title: "Create your free account — pingdan",
    description,
    url: "/register",
  },
};

export default function RegisterLayout({ children }: { children: ReactNode }) {
  return children;
}
