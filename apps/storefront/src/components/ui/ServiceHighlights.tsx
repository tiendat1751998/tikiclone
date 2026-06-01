export function ServiceHighlights() {
  const services = [
    { icon: "🚚", title: "Free Shipping", desc: "On orders over 500k" },
    { icon: "🔄", title: "Free Returns", desc: "Within 30 days" },
    { icon: "🔒", title: "Secure Payment", desc: "256-bit SSL" },
    { icon: "📞", title: "24/7 Support", desc: "Call or chat" },
  ];
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
      {services.map((s) => (
        <div key={s.title} className="bg-white rounded-lg border border-gray-200 p-3 md:p-4 flex items-center gap-3">
          <span className="text-xl md:text-2xl">{s.icon}</span>
          <div>
            <div className="text-xs md:text-sm font-semibold text-gray-900">{s.title}</div>
            <div className="text-[10px] md:text-xs text-gray-500">{s.desc}</div>
          </div>
        </div>
      ))}
    </div>
  );
}
