import type { Metadata } from "next";
import { Darker_Grotesque, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Header } from "@/components/Header";

const sans = Darker_Grotesque({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700", "800", "900"],
  variable: "--font-sans",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  variable: "--font-mono",
});

export const metadata: Metadata = {
  title: "Reunite — campus lost & found",
  description: "Match lost and found items on campus using descriptions, photos, places, and dates.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${sans.variable} ${mono.variable} bg-surface font-sans text-ink antialiased`}>
        <Header />
        <main className="mx-auto w-full max-w-5xl px-4 pb-20 pt-8 sm:px-6">{children}</main>
      </body>
    </html>
  );
}
