package domain

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/promotion/internal/domain"
)

func TestNewVoucher(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("SAVE10", "10% Off", domain.VoucherTypePercentage, 10, 5000, 20000, now, now.Add(24*time.Hour))
	if v.Code != "SAVE10" {
		t.Errorf("expected SAVE10, got %s", v.Code)
	}
	if v.Type != domain.VoucherTypePercentage {
		t.Errorf("expected percentage type, got %s", v.Type)
	}
	if v.DiscountValue != 10 {
		t.Errorf("expected discount 10, got %d", v.DiscountValue)
	}
	if v.Status != domain.VoucherStatusActive {
		t.Errorf("expected active status, got %s", v.Status)
	}
}

func TestVoucher_IsActive(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("TEST", "Test", domain.VoucherTypeFixed, 5000, 0, 0, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if !v.IsActive() {
		t.Error("voucher should be active")
	}
}

func TestVoucher_IsExpired(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("TEST", "Test", domain.VoucherTypeFixed, 5000, 0, 0, now.Add(-2*time.Hour), now.Add(-1*time.Hour))
	if !v.IsExpired() {
		t.Error("voucher should be expired")
	}
	if v.IsActive() {
		t.Error("expired voucher should not be active")
	}
}

func TestVoucher_CalculateDiscount_Percentage(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("PCT10", "10% Off", domain.VoucherTypePercentage, 10, 0, 0, now, now.Add(24*time.Hour))
	discount := v.CalculateDiscount(100000)
	if discount != 10000 {
		t.Errorf("expected discount 10000, got %d", discount)
	}
}

func TestVoucher_CalculateDiscount_PercentageWithCap(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("PCT50", "50% Off capped", domain.VoucherTypePercentage, 50, 0, 3000, now, now.Add(24*time.Hour))
	discount := v.CalculateDiscount(100000)
	if discount != 3000 {
		t.Errorf("expected capped discount 3000, got %d", discount)
	}
}

func TestVoucher_CalculateDiscount_Fixed(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("FIXED50", "50 Off", domain.VoucherTypeFixed, 5000, 0, 0, now, now.Add(24*time.Hour))
	discount := v.CalculateDiscount(100000)
	if discount != 5000 {
		t.Errorf("expected discount 5000, got %d", discount)
	}
}

func TestVoucher_CalculateDiscount_FixedExceedsSubtotal(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("FIXED50", "50 Off", domain.VoucherTypeFixed, 5000, 0, 0, now, now.Add(24*time.Hour))
	discount := v.CalculateDiscount(3000)
	if discount != 3000 {
		t.Errorf("expected discount capped at subtotal 3000, got %d", discount)
	}
}

func TestVoucher_CanRedeem(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("TEST", "Test", domain.VoucherTypeFixed, 5000, 10000, 0, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	err := v.CanRedeem("USER-001", 15000, "", "", "", "", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVoucher_CanRedeem_MinSpend(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("TEST", "Test", domain.VoucherTypeFixed, 5000, 10000, 0, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	err := v.CanRedeem("USER-001", 5000, "", "", "", "", "")
	if err == nil {
		t.Error("expected error for min spend not met")
	}
}

func TestVoucher_CanRedeem_Scope(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("SHOP", "Shop Voucher", domain.VoucherTypeFixed, 5000, 0, 0, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	v.Scope = domain.VoucherScopeShop
	v.ShopID = strPtr("SHOP-001")
	err := v.CanRedeem("USER-001", 10000, "SHOP-002", "", "", "", "")
	if err == nil {
		t.Error("expected error for wrong shop scope")
	}
}

func TestVoucher_CanRedeem_Region(t *testing.T) {
	now := time.Now()
	v := domain.NewVoucher("SG", "SG Only", domain.VoucherTypeFixed, 5000, 0, 0, now.Add(-1*time.Hour), now.Add(1*time.Hour))
	v.Region = strPtr("SG")
	err := v.CanRedeem("USER-001", 10000, "", "", "", "MY", "")
	if err == nil {
		t.Error("expected error for wrong region")
	}
}

func strPtr(s string) *string { return &s }
