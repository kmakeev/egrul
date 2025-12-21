import type { Metadata } from "next";
import { Onest } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";
import { MapsConfigStatus } from "@/components/dev/maps-config-status";

const onest = Onest({
  subsets: ["latin", "cyrillic"],
  variable: "--font-onest",
});

export const metadata: Metadata = {
  title: "ЕГРЮЛ/ЕГРИП - Поиск по реестрам",
  description: "Система поиска и анализа данных ЕГРЮЛ и ЕГРИП",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru" className={onest.variable}>
      <body className="antialiased bg-slate-950 text-slate-100 min-h-screen">
        <Providers>{children}</Providers>
        <MapsConfigStatus />
      </body>
    </html>
  );
}

