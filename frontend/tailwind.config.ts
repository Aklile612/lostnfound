import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: "#DD614C",
        secondary: "#DAA144",
        ink: "#111827",
        surface: "#FFFFFF",
        success: "#16A34A",
        warning: "#D97706",
        danger: "#DC2626",
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        display: ["var(--font-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "monospace"],
      },
      borderRadius: {
        sm: "4px",
        md: "8px",
      },
      boxShadow: {
        brutal: "6px 6px 0 0 #111827",
        "brutal-sm": "4px 4px 0 0 #111827",
        "brutal-primary": "6px 6px 0 0 #DD614C",
      },
      spacing: {
        3: "12px",
        4: "16px",
        6: "24px",
        8: "32px",
      },
    },
  },
  plugins: [],
};

export default config;
