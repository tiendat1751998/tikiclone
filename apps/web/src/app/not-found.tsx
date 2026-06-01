import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";

export default function NotFound() {
  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-16 min-h-[60vh]">
        <div className="max-w-md mx-auto px-3 text-center">
          <div className="bg-white rounded-lg border border-tiki-border p-8">
            <p className="text-6xl mb-4">🔍</p>
            <h1 className="text-xl font-bold text-tiki-text mb-2">404</h1>
            <p className="text-sm text-tiki-text-secondary mb-2">Không tìm thấy trang</p>
            <p className="text-xs text-tiki-text-secondary mb-6">
              Trang bạn đang tìm có thể đã bị xóa hoặc không tồn tại.
            </p>
            <div className="space-y-3">
              <Link href="/" className="block w-full py-2.5 bg-tiki-blue text-white rounded-lg text-sm font-medium hover:bg-tiki-blue-dark transition">
                Về trang chủ
              </Link>
              <Link href="/products" className="block w-full py-2.5 border border-tiki-border text-tiki-text-secondary rounded-lg text-sm font-medium hover:bg-gray-50 transition">
                Xem sản phẩm
              </Link>
            </div>
            <p className="mt-6 text-xs text-tiki-text-secondary">
              Cần hỗ trợ? Gọi <strong className="text-tiki-blue">1900-6035</strong>
            </p>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
