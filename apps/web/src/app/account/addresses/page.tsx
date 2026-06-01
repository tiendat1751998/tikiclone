"use client";

import { useState } from "react";
import { Button } from "@/components/ui";
import type { Address } from "@/types";

const EMPTY_ADDRESS: Address = {
  name: "",
  phone: "",
  address_line1: "",
  city: "Hà Nội",
  state: "",
  postal_code: "",
  country: "Việt Nam",
};

export default function AddressesPage() {
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [editing, setEditing] = useState<Address | null>(null);
  const [isAdding, setIsAdding] = useState(false);

  function handleAdd() {
    setEditing({ ...EMPTY_ADDRESS, id: crypto.randomUUID() });
    setIsAdding(true);
  }

  function handleSaveAddress(e?: React.FormEvent) {
    e?.preventDefault();
    if (!editing) return;
    if (isAdding) {
      setAddresses((prev) => [...prev, editing]);
    } else {
      setAddresses((prev) => prev.map((a) => (a.id === editing.id ? editing : a)));
    }
    setEditing(null);
    setIsAdding(false);
  }

  function handleDelete(id: string) {
    setAddresses((prev) => prev.filter((a) => a.id !== id));
  }

  function handleEdit(addr: Address) {
    setEditing({ ...addr });
    setIsAdding(false);
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-lg font-semibold text-tiki-text">Sổ địa chỉ</h1>
        <Button onClick={handleAdd} size="sm">+ Thêm địa chỉ mới</Button>
      </div>

      {addresses.length === 0 && !editing ? (
        <div className="bg-white rounded-lg border border-tiki-border py-16 text-center">
          <p className="text-4xl mb-3">📍</p>
          <p className="text-sm text-tiki-text-secondary">Chưa có địa chỉ nào</p>
          <p className="text-xs text-tiki-text-secondary mt-1">Thêm địa chỉ để thanh toán nhanh hơn</p>
        </div>
      ) : (
        <div className="space-y-3">
          {addresses.map((addr) => (
            <div key={addr.id} className="bg-white rounded-lg border border-tiki-border p-4">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-semibold text-tiki-text">{addr.name}</span>
                    {addr.is_default && (
                      <span className="text-[10px] bg-green-100 text-green-700 px-1.5 py-0.5 rounded font-medium">Mặc định</span>
                    )}
                  </div>
                  <p className="text-sm text-tiki-text-secondary">{addr.phone}</p>
                  <p className="text-sm text-tiki-text-secondary">
                    {[addr.address_line1, addr.city, addr.state].filter(Boolean).join(", ")}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button onClick={() => handleEdit(addr)} className="text-xs text-tiki-blue hover:underline">Sửa</button>
                  <button onClick={() => handleDelete(addr.id!)} className="text-xs text-red-500 hover:underline">Xóa</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {editing && (
        <div className="bg-white rounded-lg border border-tiki-border p-4 mt-4">
          <h3 className="text-sm font-semibold text-tiki-text mb-4">
            {isAdding ? "Thêm địa chỉ mới" : "Sửa địa chỉ"}
          </h3>
          <form onSubmit={handleSaveAddress} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Họ và tên</label>
                <input
                  required
                  type="text"
                  value={editing.name}
                  onChange={(e) => setEditing({ ...editing, name: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Số điện thoại</label>
                <input
                  required
                  type="tel"
                  value={editing.phone}
                  onChange={(e) => setEditing({ ...editing, phone: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Địa chỉ</label>
              <input
                required
                type="text"
                value={editing.address_line1}
                onChange={(e) => setEditing({ ...editing, address_line1: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                placeholder="Số nhà, đường, phường/xã"
              />
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Tỉnh/Thành phố</label>
                <select
                  value={editing.city}
                  onChange={(e) => setEditing({ ...editing, city: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                >
                  <option value="Hà Nội">Hà Nội</option>
                  <option value="Hồ Chí Minh">Hồ Chí Minh</option>
                  <option value="Đà Nẵng">Đà Nẵng</option>
                  <option value="Hải Phòng">Hải Phòng</option>
                  <option value="Cần Thơ">Cần Thơ</option>
                  <option value="Khác">Khác</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Quận/Huyện</label>
                <input
                  type="text"
                  value={editing.state || ""}
                  onChange={(e) => setEditing({ ...editing, state: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-tiki-text-secondary mb-1">Mã bưu điện</label>
                <input
                  type="text"
                  value={editing.postal_code || ""}
                  onChange={(e) => setEditing({ ...editing, postal_code: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue outline-none"
                />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="isDefault"
                checked={Boolean(editing.is_default)}
                onChange={(e) => setEditing({ ...editing, is_default: e.target.checked })}
                className="rounded border-gray-300"
              />
              <label htmlFor="isDefault" className="text-sm text-tiki-text-secondary">Đặt làm địa chỉ mặc định</label>
            </div>
            <div className="flex gap-3 pt-2">
              <Button type="submit" size="sm">Lưu địa chỉ</Button>
              <Button type="button" variant="secondary" size="sm" onClick={() => { setEditing(null); setIsAdding(false); }}>
                Hủy
              </Button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}