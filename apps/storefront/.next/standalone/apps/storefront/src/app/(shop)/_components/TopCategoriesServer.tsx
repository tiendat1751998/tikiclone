// SERVER COMPONENT - Fetches categories on server, zero client JS
import Link from "next/link";
import { categoriesApi } from "@/lib/api/products";

interface Category {
  id: string; name: string; slug: string; image_url?: string;
}

export async function TopCategoriesServer() {
  const categories: Category[] = await categoriesApi.getTree();
  return (
    <div className="grid grid-cols-4 sm:grid-cols-5 md:grid-cols-8 gap-3 md:gap-4">
      {categories.slice(0, 8).map((cat) => (
        <Link key={cat.id} href={`/categories/${cat.slug}`} className="flex flex-col items-center gap-2 group">
          <div className="w-14 h-14 md:w-16 md:h-16 rounded-full bg-[#e5f4ff] flex items-center justify-center group-hover:bg-[#189eff] transition-all duration-200 shadow-sm group-hover:shadow-md">
            {cat.image_url ? (
              <img src={cat.image_url} alt={cat.name} className="w-full h-full object-cover rounded-full" loading="lazy" width={64} height={64} />
            ) : (
              <span className="text-xl md:text-2xl font-bold text-[#189eff] group-hover:text-white transition-colors">{cat.name.charAt(0)}</span>
            )}
          </div>
          <span className="text-[11px] md:text-xs text-center text-[#222] group-hover:text-[#189eff] transition-colors line-clamp-2 leading-tight">{cat.name}</span>
        </Link>
      ))}
    </div>
  );
}
