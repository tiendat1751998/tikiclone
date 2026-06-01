export function HeroCarousel() {
  const slides = [
    { title: "Summer Sale", subtitle: "Up to 50% off", bg: "linear-gradient(135deg, #FF6B00, #FF424E)" },
    { title: "Free Shipping", subtitle: "On orders over 500k", bg: "linear-gradient(135deg, #1A94FF, #5B8DEF)" },
    { title: "New Arrivals", subtitle: "Check out latest products", bg: "linear-gradient(135deg, #00AB56, #10B981)" },
  ];
  return (
    <div className="relative rounded-xl overflow-hidden h-48 md:h-64">
      {slides.map((s, i) => (
        <div key={i} className="absolute inset-0 flex items-center justify-center text-white" style={{ background: s.bg, opacity: i === 0 ? 1 : 0 }}>
          <div className="text-center">
            <h2 className="text-2xl md:text-4xl font-bold">{s.title}</h2>
            <p className="text-sm md:text-base mt-2 opacity-90">{s.subtitle}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
