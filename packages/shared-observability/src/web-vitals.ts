"use client";
import { onCLS, onFID, onLCP, onFCP, onTTFB, onINP, type Metric } from "web-vitals";

type VitalCallback = (metric: Metric) => void;

const reportMetric: VitalCallback = (metric) => {
  if (typeof window !== "undefined" && window.__tiki_observability) {
    window.__tiki_observability.report(metric);
  }
  if (process.env.NODE_ENV === "development") {
    console.log(`[WebVital] ${metric.name}: ${metric.value}`, metric);
  }
};

export function initWebVitals() {
  if (typeof window === "undefined") return;
  onCLS(reportMetric);
  onFID(reportMetric);
  onLCP(reportMetric);
  onFCP(reportMetric);
  onTTFB(reportMetric);
  onINP(reportMetric);
}

declare global {
  interface Window {
    __tiki_observability?: { report: VitalCallback };
  }
}
