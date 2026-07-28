// SVG sprite virtual module (vite-plugin-svg-icons does not expose ./client via package exports)
declare module "virtual:svg-icons-register" {
  const component: unknown;
  export default component;
}
