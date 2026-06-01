import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";

export default function VIPPage() {
  return (
    <>
      <Header />
      <main style={{ backgroundColor: "#F5F5FA" }} className="py-12">
        <div className="max-w-tiki mx-auto px-6 text-center">
          <div className="text-5xl mb-4">👑</div>
          <h1 className="text-2xl font-bold text-tiki-text mb-2">Tiki VIP</h1>
          <p className="text-tiki-text-secondary mb-6">
            Chương trình khách hàng thân thiết Tiki - Ưu đãi đặc biệt dành cho bạn
          </p>
          <Link href="/products" className="inline-block px-6 py-2.5 bg-tiki-blue text-white rounded-lg font-semibold hover:bg-tiki-blue-dark transition">
            Khám phá ưu đãi
          </Link>
        </div>
      </main>
      <Footer />
    </>
  );
}