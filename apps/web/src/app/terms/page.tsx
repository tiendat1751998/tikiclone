import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";

export default function TermsPage() {
  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-16 min-h-[60vh]">
        <div className="max-w-3xl mx-auto px-3">
          <h1 className="text-2xl font-bold text-tiki-text mb-6">Điều khoản sử dụng</h1>
          <div className="bg-white rounded-lg border border-tiki-border p-6">
            <p className="text-sm text-tiki-text-secondary">Điều khoản sử dụng đang được phát triển.</p>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}