/** @type {import("prettier").Config} */
const config = {
  printWidth: 120,
  semi: true,
  tabWidth: 2,
  useTabs: false,
  trailingComma: "none",
  plugins: ["prettier-plugin-svelte", "prettier-plugin-tailwindcss"],
  overrides: [{ files: "*.svelte", options: { parser: "svelte" } }],
  tailwindStylesheet: "./src/lib/assets/app.css"
};

export default config;
