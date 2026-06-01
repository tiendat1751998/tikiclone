"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Header } from "@/components/layout/header/Header";
import { Footer } from "@/components/layout/footer/Footer";
import { useLogin } from "@/hooks/useApi";

export default function LoginPage() {
  const router = useRouter();
  const loginMutation = useLogin();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    loginMutation.mutate(
      { email, password },
      {
        onSuccess: () => router.push("/"),
        onError: (err: Error) => setError(err.message || "Đăng nhập thất bại"),
      }
    );
  }

  return (
    <>
      <Header />
      <main className="py-12">
        <div className="max-w-[400px] mx-auto px-6">
          <div className="bg-white rounded-lg shadow-sm border border-tiki-border p-8">
            <h1 className="text-lg font-semibold text-tiki-text text-center mb-6">Đăng nhập</h1>
            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
                {error}
              </div>
            )}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-tiki-text mb-1">Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                  placeholder="Email của bạn"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-tiki-text mb-1">Mật khẩu</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-tiki-blue"
                  placeholder="Mật khẩu"
                />
              </div>
              <div className="flex items-center justify-between mt-3 mb-4">
              <label className="flex items-center gap-2 text-xs text-tiki-text-secondary cursor-pointer">
                <input type="checkbox" className="rounded border-gray-300" />
                Ghi nhớ đăng nhập
              </label>
              <Link href="/forgot-password" className="text-xs text-tiki-blue hover:underline">
                Quên mật khẩu?
              </Link>
            </div>

            <button
                type="submit"
                disabled={loginMutation.isPending}
                className="w-full py-2.5 bg-tiki-blue text-white rounded-lg font-semibold text-sm hover:bg-tiki-blue-dark transition disabled:opacity-50"
              >
                {loginMutation.isPending ? "Đang xử lý..." : "Đăng nhập"}
              </button>
            </form>
            <div className="mt-6 text-center text-xs text-tiki-text-secondary">
              Chưa có tài khoản?{" "}
              <Link href="/register" className="text-tiki-blue font-medium hover:underline">Đăng ký</Link>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
