// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import starlightThemeRapide from "starlight-theme-rapide";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      plugins: [starlightThemeRapide()],
      title: "Treenq",
      logo: { src: "./src/assets/logo.svg" },
      social: {
        github: "https://github.com/treenq/treenq",
      },
      customCss: ["./src/styles/custom.css"],
      sidebar: [
        {
          label: "Get Started",
          items: [{ label: "Install", slug: "started/get-started" }],
        },
      ],
    }),
  ],
});
