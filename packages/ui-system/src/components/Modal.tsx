"use client";
import { useEffect, type ReactNode } from "react";
import { clsx } from "clsx";
import { zIndex, transition } from "@tiki/design-tokens";

interface ModalProps { isOpen: boolean; onClose: () => void; title?: string; children: ReactNode; size?: "sm" | "md" | "lg" | "xl"; }

export function Modal({ isOpen, onClose, title, children, size = "md" }: ModalProps) {
  useEffect(() => {
    if (isOpen) { document.body.style.overflow = "hidden"; }
    return () => { document.body.style.overflow = ""; };
  }, [isOpen]);

  if (!isOpen) return null;
  const sizes = { sm: "max-w-sm", md: "max-w-lg", lg: "max-w-2xl", xl: "max-w-4xl" };

  return (
    <div className="fixed inset-0 z-[1040] flex items-center justify-center p-4">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className={clsx("relative bg-white rounded-lg shadow-xl w-full max-h-[90vh] overflow-auto", sizes[size])} style={{ transition }}>
        {title && (
          <div className="flex items-center justify-between px-6 py-4 border-b border-[#e8e8e8]">
            <h3 className="text-lg font-semibold">{title}</h3>
            <button onClick={onClose} className="text-[#757575] hover:text-[#222] text-2xl leading-none">&times;</button>
          </div>
        )}
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}
