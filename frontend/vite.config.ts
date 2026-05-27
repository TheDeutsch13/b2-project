import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const frontendDir = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(frontendDir, "..");
const carouselDir = path.join(rootDir, "Carousel");
const syncScript = path.join(rootDir, "scripts", "sync-carousel.mjs");

function runCarouselSync(): void {
  execFileSync(process.execPath, [syncScript], {
    cwd: rootDir,
    stdio: "inherit",
  });
}

function carouselSyncPlugin(): Plugin {
  let watcher: fs.FSWatcher | undefined;

  return {
    name: "gamegear-carousel-sync",
    buildStart() {
      runCarouselSync();
    },
    configureServer(server) {
      runCarouselSync();

      if (!fs.existsSync(carouselDir)) {
        return;
      }

      watcher = fs.watch(carouselDir, { persistent: true }, () => {
        try {
          runCarouselSync();
          server.ws.send({ type: "full-reload", path: "*" });
        } catch (error) {
          console.error("[carousel-sync]", error);
        }
      });
    },
    buildEnd() {
      watcher?.close();
    },
  };
}

export default defineConfig({
  plugins: [carouselSyncPlugin(), react()],
  server: {
    proxy: {
      "/api/auth": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },
      "/auth-uploads": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },
      "^/api/(products|categories|orders|upload|cdek|reviews|support)": {
        target: "http://localhost:8082",
        changeOrigin: true,
      },
      "/uploads": {
        target: "http://localhost:8082",
        changeOrigin: true,
      },
      "/ws": {
        target: "http://localhost:8082",
        changeOrigin: true,
        ws: true,
      },
    },
  },
});
