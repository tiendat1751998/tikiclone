import Link from "next/link";

function SearchIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M15.7832 14.1911L19.2708 17.6787C19.5725 17.9804 19.5725 18.4679 19.2708 18.7696C18.9692 19.0712 18.4816 19.0712 18.18 18.7696L14.6923 15.2819C13.3012 17.0504 11.2383 18.0362 8.98366 17.9881C5.01229 17.9008 1.83319 14.6286 1.83319 10.6558C1.83319 6.5924 5.12652 3.29907 9.18994 3.29907C13.1628 3.29907 16.4335 6.47818 16.5223 10.4495C16.5704 12.7042 15.5846 14.7671 13.8161 16.1582L15.7832 14.1911ZM9.18994 4.82073C6.00087 4.82073 3.35485 7.46525 3.35485 10.6558C3.35485 13.8449 6.00087 16.4909 9.18994 16.4909C12.3805 16.4909 15.0265 13.8449 15.0265 10.6558C15.0265 7.46525 12.3805 4.82073 9.18994 4.82073Z" fill="#0A68FF"/>
    </svg>
  );
}

function CheckBadgeIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 20 20" fill="none">
      <path d="M16.7 5.3a1 1 0 0 0-1.4 0L7 13.6 4.7 11.3a1 1 0 0 0-1.4 1.4l3 3a1 1 0 0 0 1.4 0l10-10a1 1 0 0 0 0-1.4z" fill="#0A68FF"/>
    </svg>
  );
}

function TruckIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
      <path d="M5 16V7a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v9M5 16h14a2 2 0 0 1 2 2v1H3v-1a2 2 0 0 1 2-2h2zM8 19a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v1H8v-1z" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

function RefreshIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
      <path d="M4 12a8 8 0 0 1 14.9-3.8l1.4 1.4M4 12a8 8 0 0 0 14.9 3.8l1.4-1.4" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

function BoxIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
      <path d="M21 16V8a2 2 0 0 0-1-1.73l-8-5a2 2 0 0 0-2 0l-8 5a2 2 0 0 0-1 1.73v8a2 2 0 0 0 1 1.73l8 5a2 2 0 0 0 2 0l8-5a2 2 0 0 0 1-1.73z" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

function SpeedIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="9" stroke="#808089" strokeWidth="1.5"/>
      <path d="M12 7v5l3 3" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

function TagIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
      <path d="M12 2l3 3 4 4a5 5 0 0 1-7 7L8 14l-4-4a5 5 0 0 1 7-7l2-2 2-2z" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

function HomeIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
      <path d="M3 9.5l9-7 9 7V20a2 2 0 0 1-2 2h-5a2 2 0 0 1-2-2V13H9v7a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9.5z" stroke="#808089" strokeWidth="1.5"/>
    </svg>
  );
}

export function Header() {
  return (
    <header>
      {/* Layer A: Top Utility Banner */}
      <div className="bg-[#EBFBFA] text-[#00998D] text-xs py-1.5 text-center font-medium">
        <div className="max-w-tiki mx-auto px-6">
          Freeship đơn từ 45k, giảm nhiều hơn cùng FREESHIP XTRA
        </div>
      </div>

      {/* Layer B: Main Interaction Header Block */}
      <div className="bg-white border-b border-gray-100">
        <div className="max-w-tiki mx-auto px-6 py-3">
          <div className="flex items-start justify-between gap-6">
            {/* Left Column: Logo */}
            <Link href="/" className="flex-shrink-0">
              <div className="flex flex-col items-center">
                <svg width="64" height="22" viewBox="0 0 64 22" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <rect width="64" height="22" rx="4" fill="#0A68FF"/>
                  <text x="10" y="16" fill="white" fontSize="12" fontWeight="700" fontFamily="Inter, sans-serif">Tiki</text>
                </svg>
                <span className="text-[8px] text-[#003EA1] font-semibold mt-0.5 tracking-tight">Tốt & Nhanh</span>
              </div>
            </Link>

            {/* Center Column: Search Core */}
            <div className="flex-1 max-w-[560px]">
              {/* Top Row: Search Input */}
              <div className="flex items-center border border-[#DDDDE3] rounded-lg overflow-hidden focus-within:border-[#0A68FF]">
                <div className="pl-4 pr-2">
                  <SearchIcon />
                </div>
                <input
                  type="text"
                  placeholder="Freeship đơn từ 45k"
                  className="flex-1 py-2.5 text-sm outline-none text-gray-800 placeholder:text-gray-500"
                />
                <button
                  type="button"
                  className="px-4 h-[38px] text-[#0A68FF] text-sm font-medium border-l border-[#DDDDE3] hover:bg-blue-50 transition"
                >
                  Tìm kiếm
                </button>
              </div>

              {/* Bottom Row: Quick Search Keywords */}
              <div className="flex items-center gap-x-4 mt-2 text-xs">
                <div className="flex items-center gap-1">
                  <span className="bg-[#003EA1] text-white px-1.5 py-0.5 rounded text-[10px] font-medium">điện</span>
                  <span className="text-gray-800">gia dụng</span>
                </div>
                <span className="text-gray-500">xe cộ</span>
                <span className="text-gray-500">mẹ & bé</span>
                <span className="text-gray-500">khỏe đẹp</span>
                <span className="text-gray-500">nhà cửa</span>
                <span className="text-gray-500">sách</span>
                <span className="text-gray-500">thể thao</span>
              </div>
            </div>

            {/* Right Column: Utilities + Location */}
            <div className="flex-shrink-0 flex flex-col items-end">
              <div className="flex items-center gap-1 text-xs">
                <Link href="/" className="flex items-center gap-1 px-3 py-2 rounded-lg hover:bg-gray-100">
                  <HomeIcon />
                  <span className="text-gray-600">Trang chủ</span>
                </Link>
                <div className="flex items-center gap-1 px-3 py-2 rounded-lg hover:bg-gray-100 cursor-pointer">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <path d="M12 2l3 3 4 4a5 5 0 0 1-7 7L8 14l-4-4a5 5 0 0 1 7-7l2-2 2-2z" fill="#808089"/>
                  </svg>
                  <span className="text-gray-600">Tiki VIP</span>
                </div>
                <div className="flex items-center gap-1 px-3 py-2 rounded-lg hover:bg-gray-100 cursor-pointer">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" fill="#808089"/>
                  </svg>
                  <span className="text-sm text-gray-600">Tài khoản</span>
                </div>
                <div className="relative flex items-center gap-1 px-3 py-2 rounded-lg hover:bg-gray-100 cursor-pointer">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <path d="M7 18c-1.1 0-1.99.9-1.99 2S5.9 22 7 22s2-.9 2-2-.9-2-2-2zM1 2v2h2l3.6 7.59-1.35 2.45c-.16.28-.25.61-.25.96 0 1.1.9 2 2 2h12v-2H7.42c-.14 0-.25-.11-.25-.25l.03-.12.9-1.63h7.45c.75 0 1.41-.41 1.75-1.03l3.58-6.49c.08-.14.12-.31.12-.48 0-.55-.45-1-1-1H5.21l-.94-2H1zm16 16c-1.1 0-1.99.9-1.99 2s.89 2 1.99 2 2-.9 2-2-.9-2-2-2z" fill="#808089"/>
                  </svg>
                  <span className="absolute -top-0.5 -right-0.5 bg-red-500 text-white text-[10px] font-bold rounded-full min-w-[16px] h-4 flex items-center justify-center px-1">0</span>
                </div>
              </div>

              {/* Geolocation */}
              <div className="hidden lg:flex items-center gap-1 text-xs text-gray-600 mt-2">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M8 1C5.24 1 3 3.24 3 6c0 3.5 5 9 5 9s5-5.5 5-9c0-2.76-2.24-5-5-5zm0 7.5C6.62 8.5 5.5 7.38 5.5 6S6.62 3.5 8 3.5 10.5 4.62 10.5 6 9.38 8.5 8 8.5z" fill="#808089"/>
                </svg>
                <span className="text-gray-800 font-medium">Giao đến:</span>
                <span className="underline text-gray-800 truncate max-w-[180px]">Q. Hoàn Kiếm, P. Hàng Trống, Hà Nội</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Layer C: Bottom Trust Ribbon */}
      <div className="bg-white border-b border-gray-100">
        <div className="max-w-tiki mx-auto px-6">
          <div className="flex items-center h-10 text-xs text-gray-600">
            <span className="text-gray-800 font-medium">Cam kết</span>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <CheckBadgeIcon />
              <span>100% hàng thật</span>
            </div>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <TruckIcon />
              <span>Freeship mọi đơn</span>
            </div>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <RefreshIcon />
              <span>Hoàn 200% nếu hàng giả</span>
            </div>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <BoxIcon />
              <span>30 ngày đổi trả</span>
            </div>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <SpeedIcon />
              <span>Giao nhanh 2h</span>
            </div>
            <div className="w-px h-3 bg-gray-300 mx-3"></div>
            <div className="flex items-center gap-1">
              <TagIcon />
              <span>Giá siêu rẻ</span>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}