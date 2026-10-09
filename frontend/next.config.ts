import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  // This repo also has a root-level package lock; the Next workspace is frontend/.
  turbopack: {
    root: process.cwd(),
  },
};

export default nextConfig;
