import { configDefaults, defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const proxyTarget =
  process.env.VITE_DEV_PROXY_TARGET ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": proxyTarget,
      "/healthz": proxyTarget,
    },
  },
  test: {
    environment: "jsdom",
    exclude: [...configDefaults.exclude, "e2e/**"],
    setupFiles: "./src/test/setup.ts",
  },
});
