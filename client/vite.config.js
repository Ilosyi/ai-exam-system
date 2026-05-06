import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import fs from "node:fs";
import path from "node:path";
function loadRootConfig() {
    var configPath = path.resolve(__dirname, "../config.json");
    try {
        var raw = fs.readFileSync(configPath, "utf8");
        return JSON.parse(raw);
    }
    catch (_a) {
        return {};
    }
}
var rootConfig = loadRootConfig();
var serverPort = rootConfig.serverPort && rootConfig.serverPort > 0 ? rootConfig.serverPort : 8080;
var clientPort = rootConfig.clientPort && rootConfig.clientPort > 0 ? rootConfig.clientPort : 3000;
export default defineConfig({
    plugins: [react()],
    server: {
        port: clientPort,
        proxy: {
            "/api": {
                target: "http://localhost:".concat(serverPort),
                changeOrigin: true
            }
        }
    },
    build: {
        outDir: "dist",
        emptyOutDir: true
    }
});
