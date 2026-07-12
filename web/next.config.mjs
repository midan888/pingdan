/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
  async headers() {
    const noIndex = [{ key: "X-Robots-Tag", value: "noindex, nofollow" }];
    return [
      { source: "/admin/:path*", headers: noIndex },
      { source: "/auth/:path*", headers: noIndex },
      { source: "/channels/:path*", headers: noIndex },
      { source: "/dashboard/:path*", headers: noIndex },
      { source: "/endpoints/:path*", headers: noIndex },
      { source: "/groups/:path*", headers: noIndex },
      { source: "/status-pages/:path*", headers: noIndex },
    ];
  },
};
export default nextConfig;
