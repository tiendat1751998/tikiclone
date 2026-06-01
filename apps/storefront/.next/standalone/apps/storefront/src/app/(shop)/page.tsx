// apps/storefront/src/app/(shop)/page.tsx
// SERVER COMPONENT - Zero client JS for the shell
import { Suspense } from "react";
import { Header } from "@/components/layout/Header";
import { Footer } from "@/components/layout/Footer";
import { ProductGridSkeleton } from "@/components/ui/Skeleton";
import { HeroCarousel } from "@/components/ui/HeroCarousel";
import { ServiceHighlights } from "@/components/ui/ServiceHighlights";
import { TopCategoriesServer } from "./_components/TopCategoriesServer";
import { FeaturedProductsServer } from "./_components/FeaturedProductsServer";
import { FlashSaleServer } from "./_components/FlashSaleServer";
import { RecommendedProductsServer } from "./_components/RecommendedProductsServer";

// Force dynamic rendering for fresh data
export const dynamic = "force-dynamic";
export const revalidate = 30;

export default function HomePage() {
  return (
    <div className="min-h-screen bg-[#f5f5f5]">
      <Header />
      <main className="container py-4 md:py-6 space-y-4 md:space-y-6">
        {/* Hero - static, cached */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-4">
          <div className="lg:col-span-3">
            <HeroCarousel />
          </div>
          <div className="hidden lg:block lg:col-span-1">
            <CampaignSidebar />
          </div>
        </div>

        <ServiceHighlights />

        {/* Categories - streamed from server */}
        <section>
          <SectionHeader title="Categories" href="/categories" />
          <Suspense fallback={<CategorySkeleton />}>
            <TopCategoriesServer />
          </Suspense>
        </section>

        {/* Flash Sale - streamed with countdown */}
        <Suspense fallback={<ProductGridSkeleton count={10} />}>
          <FlashSaleServer />
        </Suspense>

        {/* Featured Products - streamed */}
        <section>
          <SectionHeader title="Featured Products" href="/products" />
          <Suspense fallback={<ProductGridSkeleton count={12} />}>
            <FeaturedProductsServer />
          </Suspense>
        </section>

        {/* Recommendations - streamed */}
        <Suspense fallback={<ProductGridSkeleton count={12} />}>
          <RecommendedProductsServer />
        </Suspense>
      </main>
      <Footer />
    </div>
  );
}

// Server Component for campaign sidebar (no client JS)
function CampaignSidebar() {
  const campaigns = [
    { title: "Summer Sale 50%", desc: "Limited time offers", bg: "linear-gradient(135deg, #ec4899, #f43f5e)" },
    { title: "Free Shipping", desc: "On orders over 500k", bg: "linear-gradient(135deg, #22c55e, #10b981)" },
    { title: "New Arrivals", desc: "Latest products", bg: "linear-gradient(135deg, #3b82f6, #6366f1)" },
  ];
  return (
    <div className="bg-white rounded-xl shadow-sm p-4 w-full flex flex-col gap-2 h-full">
      <h3 className="text-xs font-bold text-[#222] uppercase tracking-wider">Campaigns</h3>
      <div className="space-y-2 flex-1">
        {campaigns.map((c) => (
          <a key={c.title} href="/products?sort=deals" className="block p-3 rounded-lg text-white hover:opacity-90 transition-opacity" style={{ background: c.bg }}>
            <div className="text-sm font-bold">{c.title}</div>
            <div className="text-[10px] text-white/80">{c.desc}</div>
          </a>
        ))}
      </div>
    </div>
  );
}

function SectionHeader({ title, href }: { title: string; href: string }) {
  return (
    <div className="flex items-center justify-between mb-3">
      <h2 className="text-sm md:text-base font-bold text-[#222]">{title}</h2>
      <a href={href} className="text-xs text-[#189eff] hover:underline font-medium">View All</a>
    </div>
  );
}

function CategorySkeleton() {
  return (
    <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-8 gap-3 md:gap-4">
      {[...Array(8)].map((_, i) => (
        <div key={i} className="flex flex-col items-center gap-2 animate-pulse">
          <div className="w-14 h-14 md:w-16 md:h-16 rounded-full bg-gray-200" />
          <div className="h-3 w-12 bg-gray-200 rounded" />
        </div>
      ))}
    </div>
  );
}
