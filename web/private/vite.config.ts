import "vitest/config";
import type { UserConfig } from "vite";
import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import adapter from "@sveltejs/adapter-static";
import { sveltekit } from "@sveltejs/kit/vite";
import { createSvgIconsPlugin } from "vite-plugin-svg-icons";

type AppConfig = UserConfig & {
  test?: unknown;
};

const config: AppConfig = {
  server: {
    proxy: {
      // The whole backend API lives under /v1 (SPA routes never start with /v1),
      // so no method-based bypass hacks are needed.
      "/v1": {
        target: "http://localhost:8088/"
      },
      // Uploaded/generated files served by the backend.
      "/drive": {
        target: "http://localhost:8088/"
      },
      // Embedded signing iframe page (backend-rendered).
      "/embed": {
        target: "http://localhost:8088/"
      }
    }
  },
  plugins: [
    tailwindcss(),
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), "src/lib/assets/svg")],
      symbolId: "icon-[dir]-[name]"
    }),
    sveltekit({
      compilerOptions: {
        // Force runes mode for the project, except for libraries. Can be removed in svelte 6.
        runes: ({ filename }) => (filename.split(/[/\\]/).includes("node_modules") ? undefined : true)
      },
      adapter: adapter({
        pages: "dist",
        assets: "dist",
        fallback: "index.html"
      }),
      alias: {
        "@": "src/lib"
      }
    })
  ],
  test: {
    expect: { requireAssertions: true },
    projects: [
      {
        extends: "./vite.config.ts",
        test: {
          name: "server",
          environment: "node",
          include: ["src/**/*.{test,spec}.{js,ts}"],
          exclude: ["src/**/*.svelte.{test,spec}.{js,ts}"]
        }
      },
      {
        extends: "./vite.config.ts",
        test: {
          name: "client",
          environment: "jsdom",
          include: ["src/**/*.svelte.{test,spec}.{js,ts}"]
        },
        resolve: {
          conditions: ["browser"]
        }
      }
    ]
  }
};

export default config;
