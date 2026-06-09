package vnpay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	vnpaysim "github.com/tikiclone/tiki/services/payment/internal/infrastructure/vnpay"
)

func TestSimulator_CreatePaymentRequest(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "TEST123",
		OrderInfo: "Test payment",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, paymentURL, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	if txnID == "" {
		t.Error("Expected transaction ID to be non-empty")
	}

	if paymentURL == "" {
		t.Error("Expected payment URL to be non-empty")
	}
}

func TestSimulator_ProcessPayment_Success(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "TEST-SUCCESS",
		OrderInfo: "Test payment success",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, _, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	resp, err := sim.ProcessPayment(context.Background(), txnID, "TEST-SUCCESS", true, "00")
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if resp.ResponseCode != "00" {
		t.Errorf("Expected ResponseCode 00, got %s", resp.ResponseCode)
	}

	if resp.Message != "Success" {
		t.Errorf("Expected Message Success, got %s", resp.Message)
	}
}

func TestSimulator_ProcessPayment_Failed(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "TEST-FAIL",
		OrderInfo: "Test payment failed",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, _, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	resp, err := sim.ProcessPayment(context.Background(), txnID, "TEST-FAIL", false, "05")
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if resp.ResponseCode != "05" {
		t.Errorf("Expected ResponseCode 05, got %s", resp.ResponseCode)
	}

	if resp.Message != "Insufficient balance" {
		t.Errorf("Expected Message 'Insufficient balance', got %s", resp.Message)
	}
}

func TestSimulator_QueryTransaction(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "TEST-QUERY",
		OrderInfo: "Test query",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, _, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	_, err = sim.ProcessPayment(context.Background(), txnID, "TEST-QUERY", true, "00")
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	queryReq := &domain.VNPayQueryRequest{
		RequestId:   "QUERY-123",
		Version:     "2.1.0",
		Command:     "querydr",
		TmnCode:     "TESTMERCHANT",
		TxnRef:      "TEST-QUERY",
		CreateDate:  domain.FormatVNPayDate(time.Now()),
		IpAddr:      "127.0.0.1",
		OrderInfo:   "Query transaction",
	}

	queryResp, err := sim.QueryTransaction(context.Background(), queryReq)
	if err != nil {
		t.Fatalf("QueryTransaction failed: %v", err)
	}

	if queryResp.ResponseCode != "00" {
		t.Errorf("Expected ResponseCode 00, got %s", queryResp.ResponseCode)
	}

	if queryResp.TransactionStatus != "00" {
		t.Errorf("Expected TransactionStatus 00, got %s", queryResp.TransactionStatus)
	}

	if queryResp.SecureHash == "" {
		t.Error("Expected SecureHash to be non-empty")
	}
}

func TestSimulator_RefundTransaction(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "TEST-REFUND",
		OrderInfo: "Test refund",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	txnID, _, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	_, err = sim.ProcessPayment(context.Background(), txnID, "TEST-REFUND", true, "00")
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	refundReq := &domain.VNPayRefundRequest{
		RequestId:       "REFUND-123",
		Version:         "2.1.0",
		Command:         "refund",
		TmnCode:         "TESTMERCHANT",
		TransactionType: "02",
		TxnRef:          "TEST-REFUND",
		Amount:          1000000,
		OrderInfo:       "Full refund",
		CreateBy:        "admin@test.com",
		CreateDate:      domain.FormatVNPayDate(time.Now()),
		IpAddr:          "127.0.0.1",
	}

	refundResp, err := sim.RefundTransaction(context.Background(), refundReq)
	if err != nil {
		t.Fatalf("RefundTransaction failed: %v", err)
	}

	if refundResp.ResponseCode != "00" {
		t.Errorf("Expected ResponseCode 00, got %s", refundResp.ResponseCode)
	}

	if refundResp.Message != "Refund success" {
		t.Errorf("Expected Message 'Refund success', got %s", refundResp.Message)
	}
}

func TestSimulator_CreateToken(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	tokenReq := &domain.VNPayTokenRequest{
		Version:   "2.1.0",
		Command:   "token_create",
		TmnCode:   "TESTMERCHANT",
		AppUserID: "user-123",
		BankCode:  "VIB",
		Locale:    "vn",
		CardType:  "01",
		TxnRef:    "TOKEN-123",
		TxnDesc:   "Create token for user",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	tokenResp, err := sim.CreateToken(context.Background(), tokenReq)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if tokenResp.ResponseCode != "00" {
		t.Errorf("Expected ResponseCode 00, got %s", tokenResp.ResponseCode)
	}

	if tokenResp.Token == "" {
		t.Error("Expected Token to be non-empty")
	}

	if tokenResp.SecureHash == "" {
		t.Error("Expected SecureHash to be non-empty")
	}
}

func TestSimulator_PayWithToken(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	tokenReq := &domain.VNPayTokenRequest{
		Version:   "2.1.0",
		Command:   "token_create",
		TmnCode:   "TESTMERCHANT",
		AppUserID: "user-456",
		BankCode:  "VIB",
		Locale:    "vn",
		CardType:  "01",
		TxnRef:    "TOKEN-PAY-456",
		TxnDesc:   "Create token",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	_, err := sim.CreateToken(context.Background(), tokenReq)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	payReq := &domain.VNPayTokenRequest{
		Version:   "2.1.0",
		Command:   "token_pay",
		TmnCode:   "TESTMERCHANT",
		AppUserID: "user-456",
		Amount:    500000,
		CurrCode:  "VND",
		TxnRef:    "PAY-456",
		TxnDesc:   "Pay with token",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	payResp, err := sim.PayWithToken(context.Background(), payReq)
	if err != nil {
		t.Fatalf("PayWithToken failed: %v", err)
	}

	if payResp.ResponseCode != "00" {
		t.Errorf("Expected ResponseCode 00, got %s", payResp.ResponseCode)
	}
}

func TestSimulator_HTTPHandler(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			VNPayTmnCode:    "TESTMERCHANT",
			VNPayHashSecret: "secret-key-123",
		},
	}

	sim := vnpaysim.NewSimulator(cfg, nil, nil, nil)

	req := &domain.VNPayPaymentRequest{
		Version:   "2.1.0",
		Command:   "pay",
		TmnCode:   "TESTMERCHANT",
		Amount:    1000000,
		CurrCode:  "VND",
		TxnRef:    "HTTP-TEST",
		OrderInfo: "HTTP handler test",
		ReturnUrl: "http://localhost/return",
		IpAddr:    "127.0.0.1",
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	_, _, err := sim.CreatePaymentRequest(req)
	if err != nil {
		t.Fatalf("CreatePaymentRequest failed: %v", err)
	}

	// Test HTTP handler
	httpReq := httptest.NewRequest("GET", "/simulator/vnpay?vnp_TxnRef=HTTP-TEST&vnp_Amount=1000000", nil)
	httpReq = httpReq.WithContext(context.Background())
	recorder := httptest.NewRecorder()

	sim.ServeHTTP(recorder, httpReq)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", recorder.Header().Get("Content-Type"))
	}
}

func TestGenerateVNPaySecureHash(t *testing.T) {
	params := map[string]string{
		"vnp_Version":  "2.1.0",
		"vnp_Command": "pay",
		"vnp_TmnCode": "TESTMERCHANT",
	}

	hash := domain.GenerateVNPaySecureHash(params, "secret-key")

	if hash == "" {
		t.Error("Expected hash to be non-empty")
	}

	if !domain.ValidateVNPaySecureHash(params, "secret-key", hash) {
		t.Error("Hash validation failed")
	}

	if domain.ValidateVNPaySecureHash(params, "wrong-secret", hash) {
		t.Error("Expected validation to fail with wrong secret")
	}
}

func TestVNPayResponseCodes(t *testing.T) {
	codes := domain.VNPayResponseCodes

	if codes["00"] == "" {
		t.Error("Expected code 00 to exist")
	}

	if codes["00"] != "Success" {
		t.Errorf("Expected code 00 to be 'Success', got %s", codes["00"])
	}

	if codes["91"] == "" {
		t.Error("Expected code 91 to exist")
	}
}

func TestVNPayFormatDate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	formatted := domain.FormatVNPayDate(now)

	parsed, err := domain.ParseVNPayDateTime(formatted)
	if err != nil {
		t.Fatalf("Failed to parse formatted date: %v", err)
	}

	if !now.Equal(parsed) {
		t.Errorf("Expected %v, got %v", now, parsed)
	}
}