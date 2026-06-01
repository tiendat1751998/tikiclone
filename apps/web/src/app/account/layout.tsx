"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { useAuthStore } from "@/stores/auth";

export default function AccountLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, router]);

  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-6 min-h-[60vh]">
        <div className="max-w-7xl mx-auto px-4">{children}</div>
      </main>
      <Footer />
    </>
  );
}