import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";

export default async function CheckoutSuccessPage({ searchParams }: { searchParams: Promise<{ order_id?: string }> }) {
  const sp = await searchParams;
  const orderId = sp.order_id || "";

  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-10 min-h-[60vh]">
        <div className="max-w-lg mx-auto px-3">
          <div className="bg-white rounded-lg border border-tiki-border p-8 text-center">
            <div className="w-20 h-20 bg-green-50 rounded-full flex items-center justify-center mx-auto mb-4">
              <span className="text-4xl">✅</span>
            </div>
            <h1 className="text-xl font-bold text-tiki-text mb-2">Đặt hàng thành công!</h1>
            <p className="text-sm text-tiki-text-secondary mb-6">
              Cảm ơn bạn đã mua sắm tại Tiki. Chúng tôi sẽ xử lý đơn hàng trong thời gian sớm nhất.
            </p>

            {orderId && (
              <div className="bg-gray-50 rounded-lg p-4 mb-6">
                <p className="text-xs text-tiki-text-secondary">Mã đơn hàng</p>
                <p className="text-sm font-bold text-tiki-text font-mono">{orderId.slice(0, 12)}</p>
              </div>
            )}

            <div className="space-y-3">
              {orderId && (
                <Link
                  href={`/account/orders/${orderId}`}
                  className="block w-full py-2.5 bg-tiki-blue text-white rounded-lg text-sm font-medium hover:bg-tiki-blue-dark transition"
                >
                  Xem đơn hàng
                </Link>
              )}
              <Link
                href="/products"
                className="block w-full py-2.5 border border-tiki-border text-tiki-text-secondary rounded-lg text-sm font-medium hover:bg-gray-50 transition"
              >
                Tiếp tục mua sắm
              </Link>
            </div>

            <div className="mt-6 pt-6 border-t border-tiki-border text-xs text-tiki-text-secondary space-y-1">
              <p>Bạn sẽ nhận được email xác nhận đơn hàng trong vài phút tới.</p>
              <p>Cần hỗ trợ? Liên hotline: <strong className="text-tiki-blue">1900-6035</strong></p>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
