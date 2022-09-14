/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  images: {
    loader: 'imgix',
    path: '/',
    domains: ['cdn.akamai.steamstatic.com'],
  },
};

module.exports = nextConfig;
