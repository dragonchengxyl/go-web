const stripTrailingSlash = (value) => value.replace(/\/+$/, '')
const internalApiOrigin = stripTrailingSlash(process.env.INTERNAL_API_ORIGIN || 'http://localhost:8080')
const internalWsOrigin = stripTrailingSlash(process.env.INTERNAL_WS_ORIGIN || internalApiOrigin)

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  images: {
    domains: ['localhost'],
    formats: ['image/avif', 'image/webp'],
  },
  experimental: {
    optimizePackageImports: ['lucide-react'],
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${internalApiOrigin}/api/:path*`,
      },
      {
        source: '/ws/:path*',
        destination: `${internalWsOrigin}/ws/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
