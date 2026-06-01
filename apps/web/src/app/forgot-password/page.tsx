"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";

export default function ForgotPasswordPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/v1/auth/forgot-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (!res.ok) throw new Error("Gửi email thất bại. Vui lòng thử lại.");
      setSubmitted(true);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Có lỗi xảy ra");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Header />
      <main className="bg-[#F5F5FA] py-10 min-h-[60vh]">
        <div className="max-w-md mx-auto px-3">
          <div className="bg-white rounded-lg border border-tiki-border p-6">
            {!submitted ? (
              <>
                <div className="text-center mb-6">
                  <div className="w-16 h-16 bg-blue-50 rounded-full flex items-center justify-center mx-auto mb-3">
                    <span className="text-2xl">🔑</span>
                  </div>
                  <h1 className="text-lg font-semibold text-tiki-text">Quên mật khẩu?</h1>
                  <p className="text-sm text-tiki-text-secondary mt-1">
                    Nhập email đăng nhập, chúng tôi sẽ gửi link đặt lại mật khẩu.
                  </p>
                </div>

                {error && (
                  <div className="text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2 mb-4">{error}</div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                  <div>
                    <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Email</label>
                    <input
                      type="email"
                      required
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                      placeholder="email@example.com"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={loading}
                    className="w-full py-2.5 bg-tiki-blue text-white rounded-lg text-sm font-medium hover:bg-tiki-blue-dark transition disabled:opacity-50"
                  >
                    {loading ? "Đang gửi..." : "Gửi link đặt lại"}
                  </button>
                </form>

                <div className="text-center mt-4">
                  <Link href="/login" className="text-sm text-tiki-blue hover:underline">← Quay lại đăng nhập</Link>
                </div>
              </>
            ) : (
              <div className="text-center py-6">
                <div className="w-16 h-16 bg-green-50 rounded-full flex items-center justify-center mx-auto mb-3">
                  <span className="text-2xl">✅</span>
                </div>
                <h1 className="text-lg font-semibold text-tiki-text mb-2">Email đã được gửi!</h1>
                <p className="text-sm text-tiki-text-secondary mb-4">
                  Chúng tôi đã gửi link đặt lại mật khẩu đến <strong>{email}</strong>.
                  Vui lòng kiểm tra hộp thư của bạn.
                </p>
                <Link href="/login" className="text-sm text-tiki-blue hover:underline">← Quay lại đăng nhập</Link>
              </div>
            )}
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
