/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // "standalone" emits a minimal self-contained server (node_modules traced)
  // into .next/standalone — that's what the Dockerfile ships, keeping the
  // production image small.
  output: "standalone",
};

export default nextConfig;
