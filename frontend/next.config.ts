import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  experimental: {
    // Оптимизации для продакшена
  },
};

export default nextConfig;

