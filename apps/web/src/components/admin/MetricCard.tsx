import type { Alert, ChartDataPoint } from "@/types";

interface MetricCardProps {
  title: string;
  value: string;
  change?: number;
  icon: string;
  color: "blue" | "purple" | "green" | "orange" | "red";
}

const colorMap = {
  blue: "bg-blue-50 text-blue-600",
  purple: "bg-purple-50 text-purple-600",
  green: "bg-green-50 text-green-600",
  orange: "bg-orange-50 text-orange-600",
  red: "bg-red-50 text-red-600",
};

export default function MetricCard({ title, value, change, icon, color }: MetricCardProps) {
  return (
    <div className="bg-white rounded-lg border border-tiki-border p-4">
      <div className="flex items-center gap-3 mb-3">
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-lg ${colorMap[color]}`}>
          {icon}
        </div>
        <span className="text-xs text-tiki-text-secondary">{title}</span>
      </div>
      <div className="text-xl font-bold text-tiki-text">{value}</div>
      {change !== undefined && (
        <div className={`text-xs mt-1 ${change >= 0 ? "text-green-600" : "text-red-600"}`}>
          {change >= 0 ? "↑" : "↓"} {Math.abs(change).toFixed(1)}%
        </div>
      )}
    </div>
  );
}
