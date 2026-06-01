// apps/storefront/src/app/layout.tsx
import type { Metadata, Viewport } from "next";
import { SharedQueryProvider } from "@tiki/shared-query";
import { initWebVitals } from "@tiki/shared-observability";
import { Inter } from "next/font/google";
import "@/styles/globals.css";

const inter = Inter({ subsets: ["latin", "vietnamese"], display: "swap" });

export const metadata: Metadata = {
  title: "Tiki Clone - Mua Sắm Online",
  description: "Mua sắm trực tuyến hàng triệu sản phẩm chính hãng với giá tốt.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#189EFF",
};

// Initialize Web Vitals on client
if (typeof window !== "undefined") {
  initWebVitals();
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi" className={inter.className}>
      <body className="antialiased">
        <SharedQueryProvider>
          {children}
        </SharedQueryProvider>
      </body>
    </html>
  );
}
