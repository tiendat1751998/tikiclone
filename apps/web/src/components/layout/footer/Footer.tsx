import Link from "next/link";

export function Footer() {
  return (
    <footer className="bg-[#27272A] text-white mt-8">
      <div className="max-w-tiki mx-auto px-4 py-8">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-8">
          <div>
            <h4 className="text-xs font-semibold text-white/90 mb-3 uppercase tracking-wide">Hỗ trợ khách hàng</h4>
            <ul className="space-y-1.5 text-[11px] text-gray-400">
              <li>Hotline: <a href="tel:19006035" className="text-white/80 hover:text-white font-medium">1900-6035</a></li>
              <li className="text-[9px] text-gray-500">(1000 đ/phút, 8-21h kể cả T7, CN)</li>
              <li><Link href="/faq" className="hover:text-white transition-colors">Các câu hỏi thường gặp</Link></li>
              <li><Link href="/account/orders" className="hover:text-white transition-colors">Tra cứu đơn hàng</Link></li>
              <li><Link href="/products" className="hover:text-white transition-colors">Hướng dẫn đặt hàng</Link></li>
              <li><Link href="/return-policy" className="hover:text-white transition-colors">Chính sách đổi trả</Link></li>
              <li><Link href="/warranty" className="hover:text-white transition-colors">Chính sách bảo hành</Link></li>
            </ul>
          </div>

          <div>
            <h4 className="text-xs font-semibold text-white/90 mb-3 uppercase tracking-wide">Về Tiki</h4>
            <ul className="space-y-1.5 text-[11px] text-gray-400">
              <li><Link href="/about" className="hover:text-white transition-colors">Giới thiệu Tiki</Link></li>
              <li><Link href="/products" className="hover:text-white transition-colors">Tiki Blog</Link></li>
              <li><Link href="/register" className="hover:text-white transition-colors">Tuyển dụng</Link></li>
              <li><Link href="/privacy" className="hover:text-white transition-colors">Chính sách bảo mật</Link></li>
              <li><Link href="/terms" className="hover:text-white transition-colors">Điều khoản sử dụng</Link></li>
              <li><Link href="/shipping-policy" className="hover:text-white transition-colors">Chính sách vận chuyển</Link></li>
            </ul>
          </div>

          <div>
            <h4 className="text-xs font-semibold text-white/90 mb-3 uppercase tracking-wide">Hợp tác & Liên kết</h4>
            <ul className="space-y-1.5 text-[11px] text-gray-400">
              <li><Link href="/products" className="hover:text-white transition-colors">Quy chế hoạt động</Link></li>
              <li><Link href="/products" className="hover:text-white transition-colors">Bán hàng cùng Tiki</Link></li>
              <li><Link href="/products" className="hover:text-white transition-colors">Tiki Logistics</Link></li>
            </ul>
            <h4 className="text-xs font-semibold text-white/90 mb-3 mt-5 uppercase tracking-wide">Phương thức thanh toán</h4>
            <div className="flex flex-wrap gap-1.5">
              {["Visa", "MC", "JCB", "ZaloPay", "Momo", "VNPay", "COD"].map((m) => (
                <span key={m} className="px-2 py-0.5 bg-white/10 border border-white/10 rounded text-[9px] text-gray-300">{m}</span>
              ))}
            </div>
          </div>

          <div>
            <div className="flex flex-col items-start gap-1.5 mb-4">
              <svg width="72" height="24" viewBox="0 0 64 22" fill="none">
                <rect width="64" height="22" rx="4" fill="#1A94FF"/>
                <text x="10" y="16" fill="white" fontSize="12" fontWeight="700" fontFamily="Inter, sans-serif">Tiki</text>
              </svg>
              <span className="text-[9px] text-[#6CB4FF] font-semibold tracking-wider">TỐT &amp; NHANH</span>
            </div>
            <h4 className="text-xs font-semibold text-white/90 mb-3 uppercase tracking-wide">Tải ứng dụng</h4>
            <div className="flex flex-col gap-1.5 text-[11px] text-gray-400 mb-4">
              <a href="#" className="hover:text-white transition-colors">App Store</a>
              <a href="#" className="hover:text-white transition-colors">Google Play</a>
            </div>
            <h4 className="text-xs font-semibold text-white/90 mb-3 uppercase tracking-wide">Theo dõi chúng tôi</h4>
            <div className="flex gap-2 text-[11px] text-gray-400">
              <a href="#" className="hover:text-white transition-colors">Facebook</a>
              <a href="#" className="hover:text-white transition-colors">Instagram</a>
              <a href="#" className="hover:text-white transition-colors">Zalo</a>
            </div>
          </div>
        </div>

        <div className="border-t border-white/10 pt-5 text-[10px] text-gray-500 leading-5">
          <div className="font-medium text-gray-300 text-[11px] mb-1">Công ty TNHH TI KI</div>
          <div>Tòa nhà số 52, đường Út Tịch, P.4, Q. Tân Bình, TP. Hồ Chí Minh</div>
          <div>ĐKKD số: 0309532909 — cấp lần đầu: 06/01/2010. ĐT: (028) 38 321 232</div>
          <div className="mt-1">© 2010 - 2025 - Bản quyền của Công ty TNHH TI KI</div>
        </div>
      </div>
    </footer>
  );
}
