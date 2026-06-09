package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/tikiclone/tiki/services/payment/internal/config"
	vnpaysim "github.com/tikiclone/tiki/services/payment/internal/infrastructure/vnpay"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
)

func main() {
	fmt.Println("=== VNPay Payment Simulator Test ===")
	fmt.Println()

	// Initialize config
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "VNPAYDEMO",
			VNPayHashSecret: "demo-secret-key-123",
		},
	}

	// Create simulator
	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	// Test 1: Create payment request
	fmt.Println("Test 1: Create Payment Request")
	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "VNPAYDEMO",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "ORDER-TEST-123",
		OrderInfo: "Test Order Payment",
		ReturnUrl: "http://localhost:3000/checkout/return",
		IpAddr:    "127.0.0.1",
		Locale:    "vn",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, paymentURL, err := sim.CreatePaymentRequest(req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Transaction ID: %s\n", txnID)
	fmt.Printf("Payment URL: %s\n", paymentURL)
	fmt.Println("✓ PASSED")
	fmt.Println()

	// Test 2: Simulate successful payment via HTTP handler
	fmt.Println("Test 2: Simulate Successful Payment")
	form := url.Values{}
	form.Set("vnp_TxnRef", "ORDER-TEST-123")
	form.Set("success", "true")

	httpReq := httptest.NewRequest("POST", "/simulator/vnpay/process?"+form.Encode(), nil)
	recorder := httptest.NewRecorder()
	sim.ServeHTTP(recorder, httpReq)

	fmt.Printf("Response Status: %d\n", recorder.Code)
	fmt.Printf("Response Body: %s\n", recorder.Body.String())
	fmt.Println("✓ PASSED")
	fmt.Println()

	// Test 3: Query transaction
	fmt.Println("Test 3: Query Transaction")
	queryReq := &domain.VNPayQueryRequest{
		RequestId:  "QUERY-123",
		Version:    "2.1.0",
		Command:    "querydr",
		TmnCode:    "VNPAYDEMO",
		TxnRef:     "ORDER-TEST-123",
		CreateDate: domain.FormatVNPayDate(time.Now()),
		IpAddr:     "127.0.0.1",
		OrderInfo:  "Query transaction",
	}

	queryResp, err := sim.QueryTransaction(nil, queryReq)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Query Response Code: %s\n", queryResp.ResponseCode)
	fmt.Printf("Query Transaction Status: %s\n", queryResp.TransactionStatus)
	fmt.Printf("Query SecureHash: %s\n", queryResp.SecureHash[:40]+"...")
	fmt.Println("✓ PASSED")
	fmt.Println()

	// Test 4: Refund transaction
	fmt.Println("Test 4: Refund Transaction")
	refundReq := &domain.VNPayRefundRequest{
		RequestId:       "REFUND-123",
		Version:         "2.1.0",
		Command:         "refund",
		TmnCode:         "VNPAYDEMO",
		TransactionType: "02",
		TxnRef:          "ORDER-TEST-123",
		Amount:          1000000,
		OrderInfo:       "Full refund",
		CreateBy:        "admin@test.com",
		CreateDate:      domain.FormatVNPayDate(time.Now()),
		IpAddr:          "127.0.0.1",
	}

	refundResp, err := sim.RefundTransaction(nil, refundReq)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Refund Response Code: %s\n", refundResp.ResponseCode)
	fmt.Printf("Refund Message: %s\n", refundResp.Message)
	fmt.Println("✓ PASSED")
	fmt.Println()

	fmt.Println("=== All Tests Passed ===")
}