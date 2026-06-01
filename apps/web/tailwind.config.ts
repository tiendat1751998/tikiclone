import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        tiki: {
          blue: "#0A68FF",
          "blue-dark": "#0052CC",
          red: "#FF424E",
          "red-dark": "#D6303C",
          green: "#00AB56",
          yellow: "#FDD835",
          orange: "#FC820A",
          bg: "#F5F5FA",
          text: "#27272A",
          "text-secondary": "#787880",
          border: "#EBEBF0",
          "card-bg": "#FFFFFF",
        },
      },
      fontFamily: {
        sans: [
          "Inter",
          "-apple-system",
          "BlinkMacSystemFont",
          '"Segoe UI"',
          "Roboto",
          "Arial",
          "sans-serif",
        ],
      },
      fontSize: {
        xs: ["12px", { lineHeight: "18px" }],
        sm: ["13px", { lineHeight: "20px" }],
        base: ["14px", { lineHeight: "21px" }],
        lg: ["16px", { lineHeight: "24px" }],
        xl: ["18px", { lineHeight: "27px" }],
        "2xl": ["24px", { lineHeight: "36px" }],
      },
      borderRadius: {
        sm: "4px",
        md: "8px",
        lg: "12px",
      },
      boxShadow: {
        tiki: "0 1px 4px rgba(0,0,0,0.08)",
        "tiki-md": "0 2px 12px rgba(0,0,0,0.12)",
        "tiki-lg": "0 4px 24px rgba(0,0,0,0.16)",
      },
      maxWidth: {
        "tiki": "1280px",
        "tiki-wide": "1440px",
      },
    },
  },
  plugins: [],
};

export default config;
