import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import fs from "node:fs";
import path from "node:path";

type RootConfig = {
  serverPort?: number;
  clientPort?: number;
};

function loadRootConfig(): RootConfig {
  const configPath = path.resolve(__dirname, "../config.json");
  try {
    const raw = fs.readFileSync(configPath, "utf8");
    return JSON.parse(raw) as RootConfig;
  } catch {
    return {};
  }
}

const rootConfig = loadRootConfig();
const serverPort = rootConfig.serverPort && rootConfig.serverPort > 0 ? rootConfig.serverPort : 8080;
const clientPort = rootConfig.clientPort && rootConfig.clientPort > 0 ? rootConfig.clientPort : 3000;

export default defineConfig({
  plugins: [react()],
  server: {
    port: clientPort,
    proxy: {
      "/api": {
        target: `http://localhost:${serverPort}`,
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: "dist",
    emptyOutDir: true
  }
});
