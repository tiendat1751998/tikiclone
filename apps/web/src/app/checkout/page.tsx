"use client";

import { useState, useMemo, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { useCartStore } from "@/stores/cart";
import { useAuthStore } from "@/stores/auth";
import { useUIStore } from "@/stores/ui";
import { ordersApi, paymentApi } from "@/lib/api/client";
import { deliveryApi } from "@/lib/api/client";
import { DeliveryLocationPicker } from "@/components/delivery/LocationPicker";
import { MapPin, Truck } from "lucide-react";

interface SelectedLocation {
  address: string;
  lat: number;
  lng: number;
}

export default function CheckoutPage() {
  const router = useRouter();
  const items = useCartStore((s) => s.items);
  const getSubtotal = useCartStore((s) => s.getSubtotal);
  const clearCart = useCartStore((s) => s.clearCart);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const addToast = useUIStore((s) => s.addToast);

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [city, setCity] = useState("Hà Nội");
  const [paymentMethod, setPaymentMethod] = useState("cod");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [voucherCode, setVoucherCode] = useState("");
  const [voucherApplied, setVoucherApplied] = useState<{
    code: string;
    discount: number;
    type: "percent" | "fixed";
  } | null>(null);
  const [voucherError, setVoucherError] = useState("");

  // Delivery location state
  const [deliveryLocation, setDeliveryLocation] = useState<SelectedLocation | null>(null);
  const [deliveryRoute, setDeliveryRoute] = useState<{
    distance: number;
    duration: number;
  } | null>(null);
  const [deliveryFee, setDeliveryFee] = useState(30000);

  const selectedItems = useMemo(() => items.filter((i) => i.is_selected), [items]);
  const subtotal = useMemo(() => getSubtotal(), [getSubtotal]);
  const voucherDiscount = voucherApplied
    ? voucherApplied.type === "percent"
      ? Math.round(subtotal * voucherApplied.discount / 100)
      : voucherApplied.discount
    : 0;
  const total = subtotal + deliveryFee - voucherDiscount;

  const handleLocationSelect = useCallback((loc: SelectedLocation) => {
    setDeliveryLocation(loc);
  }, []);

  const handleRouteCalculated = useCallback(
    (distance: number, duration: number) => {
      setDeliveryRoute({ distance, duration });
      // Calculate delivery fee based on distance
      if (distance < 2000) {
        setDeliveryFee(15000);
      } else if (distance < 5000) {
        setDeliveryFee(25000);
      } else if (distance < 10000) {
        setDeliveryFee(35000);
      } else {
        setDeliveryFee(50000);
      }
    },
    []
  );

  if (items.length === 0) {
    return (
      <>
        <Header />
        <main className="py-16 text-center">
          <div className="max-w-md mx-auto px-6">
            <div className="text-5xl mb-4">🛒</div>
            <h1 className="text-lg font-semibold text-tiki-text mb-2">Giỏ hàng trống</h1>
            <p className="text-sm text-tiki-text-secondary mb-6">
              Hãy thêm sản phẩm trước khi thanh toán nhé!
            </p>
            <Link
              href="/products"
              className="inline-block px-6 py-2.5 bg-tiki-blue text-white rounded-lg font-semibold text-sm hover:bg-tiki-blue-dark transition"
            >
              Mua sắm ngay
            </Link>
          </div>
        </main>
        <Footer />
      </>
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || !phone.trim() || !address.trim()) {
      addToast({
        type: "error",
        title: "Thiếu thông tin",
        message: "Vui lòng điền đầy đủ thông tin giao hàng",
      });
      return;
    }

    // If delivery location is set, use delivery API to create order with route
    if (deliveryLocation) {
      setIsSubmitting(true);
      try {
        // Create delivery order with geo data
        const deliveryOrder = await deliveryApi.createOrder({
          customer_id: "cust_" + Date.now(),
          pickup: {
            lat: 21.0285,
            lng: 105.8542,
            address: "Kho hàng Tiki, Q. Hoàn Kiếm, Hà Nội",
          },
          dropoff: {
            lat: deliveryLocation.lat,
            lng: deliveryLocation.lng,
            address: deliveryLocation.address,
          },
        });

        clearCart();
        addToast({
          type: "success",
          title: "Đặt hàng thành công!",
          message: "Đơn hàng của bạn đã được tạo. Đang tìm tài xế giao hàng...",
        });
        router.push(
          `/checkout/tracking/${(deliveryOrder as any).id}?order_id=${(deliveryOrder as any).id}`
        );
      } catch (err: any) {
        addToast({
          type: "error",
          title: "Có lỗi xảy ra",
          message: err?.message || "Không thể tạo đơn hàng, vui lòng thử lại",
        });
      } finally {
        setIsSubmitting(false);
      }
      return;
    }

    // Fallback: original order creation without geo
    setIsSubmitting(true);
    try {
      const order = await ordersApi.create({
        items: selectedItems.map((i) => ({
          product_id: i.product_id,
          sku_id: i.sku_id,
          quantity: i.quantity,
          price: i.price,
          name: i.name,
          image_url: i.image_url,
          shop_id: i.shop_id || "",
          shop_name: i.shop_name || "",
        })),
        shipping_address: {
          name: name.trim(),
          phone: phone.trim(),
          address_line1: address.trim(),
          city,
          state: city,
          country: "Vietnam",
          postal_code: "",
        },
        seller_id: selectedItems[0]?.shop_id || "usr-001",
        idempotency_key: `checkout_${Date.now()}_${Math.random().toString(36).slice(2)}`,
        currency: "VND",
        billing_address: {
          name: name.trim(),
          phone: phone.trim(),
          address_line1: address.trim(),
          city,
          state: city,
          country: "Vietnam",
          postal_code: "",
        },
        payment_method: paymentMethod,
        voucher_code: voucherApplied?.code,
      });
      clearCart();
      if (paymentMethod !== "cod") {
        try {
          await paymentApi.authorize({
            order_id: (order as any).id || (order as any).order_id || "",
            amount: total,
            currency: "VND",
            payment_method: paymentMethod,
            idempotency_key: `pay_${Date.now()}_${Math.random().toString(36).slice(2)}`,
          });
        } catch {
          // payment failed but order was created
        }
      }
      addToast({
        type: "success",
        title: "Đặt hàng thành công!",
        message: "Đơn hàng của bạn đã được tạo",
      });
      router.push(
        `/checkout/success?order_id=${(order as any).id || (order as any).order_id || ""}`
      );
    } catch (err: any) {
      addToast({
        type: "error",
        title: "Có lỗi xảy ra",
        message: err?.message || "Không thể tạo đơn hàng, vui lòng thử lại",
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <Header />
      <main className="py-4" style={{ backgroundColor: "#F5F5FA" }}>
        <div className="max-w-[1270px] mx-auto px-[15px]">
          <div className="flex items-center gap-2 text-xs text-tiki-text-secondary mb-4">
            <Link href="/" className="hover:text-tiki-blue">
              Trang chủ
            </Link>
            <span>/</span>
            <Link href="/cart" className="hover:text-tiki-blue">
              Giỏ hàng
            </Link>
            <span>/</span>
            <span className="text-tiki-text">Thanh toán</span>
          </div>

          {!isAuthenticated && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4 flex items-center justify-between">
              <p className="text-sm text-yellow-800">
                💡{" "}
                <Link
                  href="/login"
                  className="text-tiki-blue hover:underline font-medium"
                >
                  Đăng nhập
                </Link>{" "}
                để theo dõi đơn hàng và nhận ưu đãi
              </p>
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div className="flex gap-4">
              <div className="flex-1 min-w-0 space-y-4">
                {/* Delivery Address with Location Picker */}
                <div className="bg-white rounded-lg border border-tiki-border p-4">
                  <h2 className="flex items-center gap-2 text-base font-semibold text-tiki-text mb-4">
                    <MapPin className="w-5 h-5 text-tiki-red" />
                    Địa chỉ giao hàng
                  </h2>

                  {/* Nominatim-powered location search */}
                  <div className="mb-4">
                    <DeliveryLocationPicker
                      onLocationSelect={handleLocationSelect}
                      onRouteCalculated={handleRouteCalculated}
                      pickupLocation={{ lat: 21.0285, lng: 105.8542, address: "Hà Nội" }}
                      label="Giao đến"
                      placeholder="Tìm địa chỉ... (Vd: Q. Hoàn Kiếm, Hà Nội)"
                    />
                  </div>

                  {/* Manual address fields */}
                  <div className="space-y-3 border-t border-tiki-border pt-4">
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-tiki-text mb-1">
                          Họ và tên *
                        </label>
                        <input
                          type="text"
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                          placeholder="Nguyễn Văn A"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-tiki-text mb-1">
                          Số điện thoại *
                        </label>
                        <input
                          type="tel"
                          value={phone}
                          onChange={(e) => setPhone(e.target.value)}
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                          placeholder="0912345678"
                        />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-tiki-text mb-1">
                        Địa chỉ chi tiết *
                      </label>
                      <input
                        type="text"
                        value={address}
                        onChange={(e) => setAddress(e.target.value)}
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                        placeholder="Số nhà, đường, phường/xã"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-tiki-text mb-1">
                        Tỉnh/Thành phố
                      </label>
                      <select
                        value={city}
                        onChange={(e) => setCity(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                      >
                        <option>Hà Nội</option>
                        <option>TP. Hồ Chí Minh</option>
                        <option>Đà Nẵng</option>
                        <option>Hải Phòng</option>
                        <option>Cần Thơ</option>
                        <option>Khác</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Delivery Fee Info */}
                {deliveryRoute && (
                  <div className="bg-white rounded-lg border border-tiki-border p-4">
                    <h2 className="flex items-center gap-2 text-base font-semibold text-tiki-text mb-3">
                      <Truck className="w-5 h-5 text-tiki-blue" />
                      Thông tin giao hàng
                    </h2>
                    <div className="grid grid-cols-3 gap-4">
                      <div className="text-center p-3 bg-blue-50 rounded-lg">
                        <p className="text-xs text-tiki-text-secondary">Khoảng cách</p>
                        <p className="text-sm font-semibold text-tiki-text mt-1">
                          {deliveryRoute.distance >= 1000
                            ? `${(deliveryRoute.distance / 1000).toFixed(1)} km`
                            : `${deliveryRoute.distance} m`}
                        </p>
                      </div>
                      <div className="text-center p-3 bg-orange-50 rounded-lg">
                        <p className="text-xs text-tiki-text-secondary">Thời gian</p>
                        <p className="text-sm font-semibold text-tiki-text mt-1">
                          ~{Math.ceil(deliveryRoute.duration / 60)} phút
                        </p>
                      </div>
                      <div className="text-center p-3 bg-green-50 rounded-lg">
                        <p className="text-xs text-tiki-text-secondary">Phí giao hàng</p>
                        <p className="text-sm font-semibold text-tiki-green mt-1">
                          {deliveryFee === 0
                            ? "Miễn phí"
                            : `${deliveryFee.toLocaleString("vi-VN")} ₫`}
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {/* Payment method */}
                <div className="bg-white rounded-lg border border-tiki-border p-4">
                  <h2 className="text-base font-semibold text-tiki-text mb-4">
                    Phương thức thanh toán
                  </h2>
                  <div className="space-y-2">
                    {[
                      {
                        value: "cod",
                        label: "Thanh toán khi nhận hàng (COD)",
                        icon: "💵",
                      },
                      { value: "momo", label: "Ví MoMo", icon: "🟣" },
                      { value: "vnpay", label: "VNPay", icon: "🔵" },
                      {
                        value: "bank",
                        label: "Chuyển khoản ngân hàng",
                        icon: "🏦",
                      },
                    ].map((m) => (
                      <label
                        key={m.value}
                        className={`flex items-center gap-3 p-3 border rounded-lg cursor-pointer transition ${
                          paymentMethod === m.value
                            ? "border-tiki-blue bg-blue-50"
                            : "border-tiki-border hover:border-tiki-blue"
                        }`}
                      >
                        <input
                          type="radio"
                          name="payment"
                          value={m.value}
                          checked={paymentMethod === m.value}
                          onChange={() => setPaymentMethod(m.value)}
                          className="w-4 h-4"
                        />
                        <span className="text-lg">{m.icon}</span>
                        <span className="text-sm text-tiki-text">{m.label}</span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>

              {/* Order summary */}
              <div className="w-[380px] shrink-0">
                <div className="bg-white rounded-lg border border-tiki-border p-4 sticky top-4">
                  <h2 className="text-base font-semibold text-tiki-text mb-3">
                    Đơn hàng ({selectedItems.length} sản phẩm)
                  </h2>

                  <div className="max-h-[300px] overflow-y-auto space-y-3 mb-4">
                    {selectedItems.map((item) => (
                      <div key={item.id} className="flex items-center gap-3">
                        <img
                          src={item.image_url || "/images/placeholder.svg"}
                          alt={item.name}
                          className="w-12 h-12 object-cover rounded border border-tiki-border shrink-0"
                        />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-tiki-text truncate">{item.name}</p>
                          <p className="text-xs text-tiki-text-secondary">
                            x{item.quantity}
                          </p>
                        </div>
                        <span className="text-sm font-medium text-tiki-text">
                          {(item.price * item.quantity).toLocaleString("vi-VN")} ₫
                        </span>
                      </div>
                    ))}
                  </div>

                  <div className="border-t border-tiki-border pt-3 space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-tiki-text-secondary">Tạm tính</span>
                      <span className="text-tiki-text">
                        {subtotal.toLocaleString("vi-VN")} ₫
                      </span>
                    </div>

                    {/* Delivery fee */}
                    <div className="flex justify-between text-sm">
                      <span className="text-tiki-text-secondary">Phí vận chuyển</span>
                      {deliveryRoute ? (
                        <span className={deliveryFee === 0 ? "text-tiki-green" : "text-tiki-text"}>
                          {deliveryFee === 0
                            ? "Miễn phí"
                            : `${deliveryFee.toLocaleString("vi-VN")} ₫`}
                        </span>
                      ) : (
                        <span className="text-tiki-text-secondary text-xs">
                          Chọn địa chỉ giao hàng
                        </span>
                      )}
                    </div>

                    {/* Voucher */}
                    <div className="pt-2">
                      {voucherApplied ? (
                        <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg px-3 py-2">
                          <div className="flex items-center gap-2">
                            <span className="text-xs">🎟️</span>
                            <span className="text-xs font-medium text-green-700">
                              {voucherApplied.code}
                            </span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-medium text-green-700">
                              -{voucherDiscount.toLocaleString("vi-VN")} ₫
                            </span>
                            <button
                              onClick={() => {
                                setVoucherApplied(null);
                                setVoucherCode("");
                              }}
                              className="text-xs text-red-500 hover:underline"
                            >
                              Bỏ
                            </button>
                          </div>
                        </div>
) : (
                        <div>
                          <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Mã giảm giá</label>
                          <div className="flex gap-2">
                            <input
                              type="text"
                              value={voucherCode}
                              onChange={(e) => {
                                setVoucherCode(e.target.value.toUpperCase());
                                setVoucherError("");
                              }}
                              placeholder="Nhập mã giảm giá"
                              className="flex-1 rounded-lg border border-gray-300 px-3 py-1.5 text-xs focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                            />
                            <button
                              type="button"
                              onClick={() => {
                                if (!voucherCode.trim()) return;
                                if (voucherCode === "GIAM10") {
                                  setVoucherApplied({
                                    code: "GIAM10",
                                    discount: 10,
                                    type: "percent",
                                  });
                                  setVoucherError("");
                                } else if (voucherCode === "FREESHIP") {
                                  setVoucherApplied({
                                    code: "FREESHIP",
                                    discount: 30000,
                                    type: "fixed",
                                  });
                                  setVoucherError("");
                                } else {
                                  setVoucherError("Mã giảm giá không hợp lệ");
                                }
                              }}
                              className="px-3 py-1.5 bg-tiki-blue text-white rounded-lg text-xs font-medium hover:bg-tiki-blue-dark transition"
                            >
                              Áp dụng
                            </button>
                          </div>
                          {voucherError && (
                            <p className="text-[10px] text-red-500 mt-1">{voucherError}</p>
                          )}
                          <p className="text-[10px] text-tiki-text-secondary mt-1">
                            Thử: GIAM10 (giảm 10%) hoặc FREESHIP (giảm 30K)
                          </p>
                        </div>
                      )}
                    </div>

                    <div className="border-t border-tiki-border pt-2">
                      <div className="flex justify-between">
                        <span className="text-sm font-medium text-tiki-text">Tổng cộng</span>
                        <span className="text-lg font-semibold text-tiki-red">
                          {total.toLocaleString("vi-VN")} ₫
                        </span>
                      </div>
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={isSubmitting || selectedItems.length === 0}
                    className="w-full mt-4 py-3 bg-tiki-red text-white rounded-lg font-semibold text-sm hover:bg-tiki-red-dark transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isSubmitting
                      ? "Đang xử lý..."
                      : `Đặt hàng — ${total.toLocaleString("vi-VN")} ₫`}
                  </button>

                  <Link
                    href="/cart"
                    className="block text-center mt-2 text-sm text-tiki-blue hover:underline"
                  >
                    ← Quay lại giỏ hàng
                  </Link>
                </div>
              </div>
            </div>
          </form>
        </div>
      </main>
      <Footer />
    </>
  );
}
