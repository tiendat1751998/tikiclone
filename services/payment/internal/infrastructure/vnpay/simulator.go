package vnpay

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/kafka"
	mysqlinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
	"go.uber.org/zap"
)

type Simulator struct {
	cfg          *config.Config
	paymentRepo  *mysqlinfra.PaymentRepository
	redisStore   *redisinfra.Store
	kafkaProducer *kafka.Producer
	mu           sync.RWMutex
	transactions map[string]*Transaction
	tokens       map[string]*Token
}

type Transaction struct {
	ID            string
	TxnRef        string
	Amount        int64
	Currency      string
	Status        string
	BankCode      string
	TransactionNo string
	CardType      string
	PayDate       time.Time
	ResponseCode  string
	ResponseMsg   string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

type Token struct {
	UserID     string
	Token      string
	CardNumber string
	BankCode   string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type ProcessPaymentRequest struct {
	TxnRef        string `json:"vnp_TxnRef"`
	Success       bool   `json:"success"`
	ResponseCode  string `json:"response_code,omitempty"`
}

type ProcessPaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	TxnRef        string `json:"vnp_TxnRef"`
	Amount        int64  `json:"vnp_Amount"`
	ResponseCode  string `json:"vnp_ResponseCode"`
	TransactionNo string `json:"vnp_TransactionNo"`
	PayDate       string `json:"vnp_PayDate"`
	Message       string `json:"vnp_Message"`
}

func NewSimulator(cfg *config.Config, paymentRepo *mysqlinfra.PaymentRepository, redisStore *redisinfra.Store, kafkaProducer *kafka.Producer) *Simulator {
	if cfg.Payment.VNPayTmnCode == "" {
		cfg.Payment.VNPayTmnCode = "VNPAYDEMO"
	}
	if cfg.Payment.VNPayHashSecret == "" {
		cfg.Payment.VNPayHashSecret = "demo-secret-key"
	}

	return &Simulator{
		cfg:          cfg,
		paymentRepo:  paymentRepo,
		redisStore:   redisStore,
		kafkaProducer: kafkaProducer,
		transactions: make(map[string]*Transaction),
		tokens:       make(map[string]*Token),
	}
}

func (s *Simulator) CreatePaymentRequest(req *domain.VNPayPaymentRequest) (string, string, error) {
	if req.TxnRef == "" {
		req.TxnRef = fmt.Sprintf("VNPAY-%d", time.Now().UnixNano())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	txn := &Transaction{
		ID:            fmt.Sprintf("txn-%d", time.Now().UnixNano()),
		TxnRef:        req.TxnRef,
		Amount:        req.Amount,
		Currency:      req.CurrCode,
		Status:        "pending",
		BankCode:      req.BankCode,
		TransactionNo: domain.NewVNPayTransactionNo(),
		ResponseCode:  "00",
		ResponseMsg:   "Confirm at ATM/InternetBanking",
		CreatedAt:     time.Now().UTC(),
	}

	s.transactions[txn.ID] = txn

	return txn.ID, s.buildPaymentURL(txn, req), nil
}

func (s *Simulator) ProcessPayment(ctx context.Context, txnID, orderID string, success bool, responseCode string) (*ProcessPaymentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	txn, exists := s.transactions[txnID]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txnID)
	}

	now := time.Now().UTC()
	txn.ProcessedAt = &now
	txn.PayDate = now

	if success {
		txn.Status = "completed"
		txn.ResponseCode = "00"
		txn.ResponseMsg = "Success"
	} else {
		txn.Status = "failed"
		txn.ResponseCode = responseCode
		if msg, ok := domain.VNPayResponseCodes[responseCode]; ok {
			txn.ResponseMsg = msg
		} else {
			txn.ResponseMsg = "Unknown error"
		}
	}

	// Update payment record in database (if repo exists)
	if s.paymentRepo != nil {
		payment, err := s.paymentRepo.FindByOrderID(ctx, orderID)
		if err != nil {
			return nil, fmt.Errorf("payment not found for order: %s", orderID)
		}

		newStatus := domain.PaymentStatusFailed
		if success {
			newStatus = domain.PaymentStatusCaptured
			payment.PSPTransactionID = txn.TransactionNo
		} else {
			payment.FailureReason = txn.ResponseMsg
		}

		if err := payment.TransitionTo(newStatus); err != nil {
			return nil, fmt.Errorf("failed to transition payment: %w", err)
		}

		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return nil, fmt.Errorf("failed to update payment: %w", err)
		}

		s.publishVNPayWebhook(ctx, payment, success)
	}

	return &ProcessPaymentResponse{
		PaymentID:     txnID,
		TxnRef:        txn.TxnRef,
		Amount:        txn.Amount,
		ResponseCode:  txn.ResponseCode,
		TransactionNo: txn.TransactionNo,
		PayDate:       domain.FormatVNPayDate(txn.PayDate),
		Message:       txn.ResponseMsg,
	}, nil
}

func (s *Simulator) QueryTransaction(ctx context.Context, req *domain.VNPayQueryRequest) (*domain.VNPayQueryResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var txn *Transaction
	if req.TransactionNo != "" {
		for _, t := range s.transactions {
			if t.TransactionNo == req.TransactionNo && t.TxnRef == req.TxnRef {
				txn = t
				break
			}
		}
	} else {
		for _, t := range s.transactions {
			if t.TxnRef == req.TxnRef {
				txn = t
				break
			}
		}
	}

	if txn == nil {
		return &domain.VNPayQueryResponse{
			ResponseCode: "91",
			Message:      "Transaction not found",
		}, nil
	}

	status := "02"
	if txn.Status == "completed" {
		status = "00"
	}

	response := &domain.VNPayQueryResponse{
		ResponseId:        fmt.Sprintf("qr-%s", txn.ID),
		Command:           "querydr",
		TmnCode:           s.cfg.Payment.VNPayTmnCode,
		TxnRef:            txn.TxnRef,
		Amount:            txn.Amount,
		OrderInfo:         "Query transaction",
		ResponseCode:      "00",
		Message:           "Query Success",
		TransactionNo:     txn.TransactionNo,
		TransactionStatus: status,
	}

	response.SecureHash = domain.GenerateVNPaySecureHash(toVNPayMap(response), s.cfg.Payment.VNPayHashSecret)
	return response, nil
}

func (s *Simulator) RefundTransaction(ctx context.Context, req *domain.VNPayRefundRequest) (*domain.VNPayRefundResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var txn *Transaction
	for _, t := range s.transactions {
		if t.TxnRef == req.TxnRef || t.TransactionNo == req.TransactionNo {
			txn = t
			break
		}
	}

	if txn == nil {
		return &domain.VNPayRefundResponse{
			ResponseCode: "91",
			Message:      "Transaction not found",
		}, nil
	}

	now := time.Now().UTC()
	txn.Status = "refunded"

	response := &domain.VNPayRefundResponse{
		ResponseId:        fmt.Sprintf("ref-%s", txn.ID),
		Command:           "refund",
		TmnCode:           s.cfg.Payment.VNPayTmnCode,
		TxnRef:            txn.TxnRef,
		Amount:            req.Amount,
		OrderInfo:         req.OrderInfo,
		ResponseCode:      "00",
		Message:           "Refund success",
		BankCode:          txn.BankCode,
		PayDate:           domain.FormatVNPayDate(now),
		TransactionNo:     txn.TransactionNo,
		TransactionType:   req.TransactionType,
		TransactionStatus: "00",
	}

	response.SecureHash = domain.GenerateVNPaySecureHash(toVNPayRefundMap(response), s.cfg.Payment.VNPayHashSecret)
	return response, nil
}

func (s *Simulator) CreateToken(ctx context.Context, req *domain.VNPayTokenRequest) (*domain.VNPayTokenResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := fmt.Sprintf("TOKEN-%s-%d", req.TmnCode, time.Now().UnixNano())
	cardNumber := fmt.Sprintf("411111xxxxxx%04d", rand.Intn(10000))

	s.tokens[req.AppUserID] = &Token{
		UserID:     req.AppUserID,
		Token:      token,
		CardNumber: cardNumber,
		BankCode:   req.BankCode,
		CreatedAt:  time.Now().UTC(),
	}

	response := &domain.VNPayTokenResponse{
		AppUserID:    req.AppUserID,
		Token:        token,
		Command:      "token_create",
		TmnCode:      req.TmnCode,
		ResponseCode: "00",
		TxnRef:       req.TxnRef,
		CardNumber:   cardNumber,
	}

	response.SecureHash = domain.GenerateVNPaySecureHash(toVNPayTokenMap(response), s.cfg.Payment.VNPayHashSecret)
	return response, nil
}

func (s *Simulator) PayWithToken(ctx context.Context, req *domain.VNPayTokenRequest) (*domain.VNPayTokenResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tk, exists := s.tokens[req.AppUserID]
	if !exists {
		return &domain.VNPayTokenResponse{
			AppUserID:    req.AppUserID,
			ResponseCode: "91",
		}, nil
	}

	now := time.Now().UTC()
	txn := &Transaction{
		ID:            fmt.Sprintf("txn-%d", time.Now().UnixNano()),
		TxnRef:        req.TxnRef,
		Amount:        0,
		Currency:      "VND",
		Status:        "completed",
		BankCode:      tk.BankCode,
		TransactionNo: domain.NewVNPayTransactionNo(),
		PayDate:       now,
		ResponseCode:  "00",
		ResponseMsg:   "Success",
		CreatedAt:     now,
	}
	s.transactions[txn.ID] = txn

	tk.LastUsedAt = &now

	response := &domain.VNPayTokenResponse{
		AppUserID:        req.AppUserID,
		Token:            tk.Token,
		Command:          "token_pay",
		TmnCode:          req.TmnCode,
		ResponseCode:     "00",
		TxnRef:           req.TxnRef,
		TransactionNo:    txn.TransactionNo,
		CardType:         "QRCODE",
		BankCode:         tk.BankCode,
		TransactionStatus: "00",
		PayDate:          domain.FormatVNPayDate(now),
	}

	response.SecureHash = domain.GenerateVNPaySecureHash(toVNPayTokenMap(response), s.cfg.Payment.VNPayHashSecret)
	return response, nil
}

func (s *Simulator) RemoveToken(ctx context.Context, req *domain.VNPayTokenRequest) (*domain.VNPayTokenResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, req.AppUserID)

	response := &domain.VNPayTokenResponse{
		AppUserID:    req.AppUserID,
		Command:      "token_remove",
		TmnCode:      req.TmnCode,
		ResponseCode: "00",
		TxnRef:       req.TxnRef,
	}

	response.SecureHash = domain.GenerateVNPaySecureHash(toVNPayTokenMap(response), s.cfg.Payment.VNPayHashSecret)
	return response, nil
}

func (s *Simulator) buildPaymentURL(txn *Transaction, req *domain.VNPayPaymentRequest) string {
	params := url.Values{}
	params.Set("vnp_Version", req.Version)
	params.Set("vnp_Command", req.Command)
	params.Set("vnp_TmnCode", req.TmnCode)
	params.Set("vnp_Amount", fmt.Sprintf("%d", req.Amount))
	params.Set("vnp_CurrCode", req.CurrCode)
	params.Set("vnp_TxnRef", req.TxnRef)
	params.Set("vnp_OrderInfo", req.OrderInfo)
	params.Set("vnp_ReturnUrl", req.ReturnUrl)
	params.Set("vnp_IpAddr", req.IpAddr)
	params.Set("vnp_CreateDate", req.CreateDate)
	if req.ExpireDate != "" {
		params.Set("vnp_ExpireDate", req.ExpireDate)
	}
	if req.BankCode != "" {
		params.Set("vnp_BankCode", req.BankCode)
	}

	signature := domain.GenerateVNPaySecureHash(toMap(params), s.cfg.Payment.VNPayHashSecret)
	params.Set("vnp_SecureHash", signature)

	return "/simulator/vnpay?" + params.Encode()
}

func (s *Simulator) publishVNPayWebhook(ctx context.Context, payment *domain.Payment, success bool) {
	eventType := domain.EventPaymentCaptured
	if !success {
		eventType = domain.EventPaymentFailed
	}

	event := domain.NewPaymentEvent(payment, eventType, nil)
	payload, err := json.Marshal(event)
	if err != nil {
		zap.L().Error("failed to marshal vnpay webhook event", zap.Error(err))
		return
	}

	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.PublishEvent(ctx, event); err != nil {
			zap.L().Error("failed to publish vnpay webhook to Kafka", zap.Error(err))
		}
	}

	if s.paymentRepo != nil {
		if err := s.paymentRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("vnpay", payment.ID, string(eventType), payload)); err != nil {
			zap.L().Error("failed to save vnpay outbox event", zap.Error(err))
		}
	}
}

func toMap(v url.Values) map[string]string {
	m := make(map[string]string)
	for k, vs := range v {
		if len(vs) > 0 {
			m[k] = vs[0]
		}
	}
	return m
}

func toVNPayMap(resp *domain.VNPayQueryResponse) map[string]string {
	return map[string]string{
		"vnp_ResponseId":       resp.ResponseId,
		"vnp_Command":          resp.Command,
		"vnp_TmnCode":          resp.TmnCode,
		"vnp_TxnRef":           resp.TxnRef,
		"vnp_Amount":           fmt.Sprintf("%d", resp.Amount),
		"vnp_OrderInfo":        resp.OrderInfo,
		"vnp_ResponseCode":     resp.ResponseCode,
		"vnp_Message":          resp.Message,
		"vnp_BankCode":         resp.BankCode,
		"vnp_CardType":         resp.CardType,
		"vnp_PayDate":          resp.PayDate,
		"vnp_TransactionNo":    resp.TransactionNo,
		"vnp_TransactionType":  resp.TransactionType,
		"vnp_TransactionStatus": resp.TransactionStatus,
	}
}

func toVNPayRefundMap(resp *domain.VNPayRefundResponse) map[string]string {
	return map[string]string{
		"vnp_ResponseId":        resp.ResponseId,
		"vnp_Command":           resp.Command,
		"vnp_TmnCode":           resp.TmnCode,
		"vnp_TxnRef":            resp.TxnRef,
		"vnp_Amount":            fmt.Sprintf("%d", resp.Amount),
		"vnp_OrderInfo":         resp.OrderInfo,
		"vnp_ResponseCode":      resp.ResponseCode,
		"vnp_Message":           resp.Message,
		"vnp_BankCode":          resp.BankCode,
		"vnp_PayDate":           resp.PayDate,
		"vnp_TransactionNo":     resp.TransactionNo,
		"vnp_TransactionType":   resp.TransactionType,
		"vnp_TransactionStatus": resp.TransactionStatus,
	}
}

func toVNPayTokenMap(resp *domain.VNPayTokenResponse) map[string]string {
	m := map[string]string{
		"vnp_AppUserId":        resp.AppUserID,
		"vnp_Command":          resp.Command,
		"vnp_TmnCode":          resp.TmnCode,
		"vnp_ResponseCode":     resp.ResponseCode,
		"vnp_TxnRef":           resp.TxnRef,
		"vnp_TxnDesc":          resp.TxnDesc,
	}
	if resp.Token != "" {
		m["vnp_Token"] = resp.Token
	}
	if resp.CardNumber != "" {
		m["vnp_CardNumber"] = resp.CardNumber
	}
	if resp.TransactionNo != "" {
		m["vnp_TransactionNo"] = resp.TransactionNo
	}
	if resp.CardType != "" {
		m["vnp_CardType"] = resp.CardType
	}
	if resp.BankCode != "" {
		m["vnp_BankCode"] = resp.BankCode
	}
	if resp.PayDate != "" {
		m["vnp_PayDate"] = resp.PayDate
	}
	if resp.TransactionStatus != "" {
		m["vnp_TransactionStatus"] = resp.TransactionStatus
	}
	if resp.Message != "" {
		m["vnp_Message"] = resp.Message
	}
	return m
}

func (s *Simulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/simulator/vnpay":
		s.handlePaymentPage(w, r)
	case "/simulator/vnpay/process":
		s.handleProcessPayment(w, r)
	case "/simulator/vnpay/query":
		s.handleQueryTransaction(w, r)
	case "/simulator/vnpay/refund":
		s.handleRefundTransaction(w, r)
	case "/simulator/vnpay/token/create":
		s.handleCreateToken(w, r)
	case "/simulator/vnpay/token/pay":
		s.handlePayWithToken(w, r)
	case "/simulator/vnpay/token/remove":
		s.handleRemoveToken(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Simulator) handlePaymentPage(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := `<!DOCTYPE html>
<html>
<head><title>VNPay Payment Simulator</title></head>
<body>
<h2>VNPay Payment Simulator</h2>
<form action="/simulator/vnpay/process" method="POST">
<input type="hidden" name="vnp_TxnRef" value="` + params.Get("vnp_TxnRef") + `">
<button type="submit" name="success" value="true">Simulate Success</button>
<button type="submit" name="success" value="false">Simulate Failed</button>
</form>
</body>
</html>`
	w.Write([]byte(html))
}

func (s *Simulator) handleProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txnRef := r.FormValue("vnp_TxnRef")
	success := r.FormValue("success") == "true"

	// If payment repo exists, use order lookup; otherwise fall back to in-memory
	if s.paymentRepo != nil {
		payment, err := s.paymentRepo.FindByOrderID(r.Context(), txnRef)
		if err != nil {
			zap.L().Error("payment not found in DB", zap.String("txn_ref", txnRef), zap.Error(err))
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}

		resp := &ProcessPaymentResponse{
			PaymentID:     payment.ID,
			TxnRef:        txnRef,
			Amount:        payment.Amount,
			ResponseCode:  "00",
			TransactionNo: domain.NewVNPayTransactionNo(),
			PayDate:       domain.FormatVNPayDate(time.Now()),
			Message:       "Success",
		}
		if !success {
			resp.ResponseCode = "24"
			resp.Message = "User cancelled"
		}

		newStatus := domain.PaymentStatusCaptured
		if !success {
			newStatus = domain.PaymentStatusFailed
			payment.FailureReason = resp.Message
		} else {
			payment.PSPTransactionID = resp.TransactionNo
		}

		if err := payment.TransitionTo(newStatus); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := s.paymentRepo.Update(r.Context(), payment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.publishVNPayWebhook(r.Context(), payment, success)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Fallback to in-memory transaction lookup
	var txnID string
	s.mu.RLock()
	for id, txn := range s.transactions {
		if txn.TxnRef == txnRef {
			txnID = id
			break
		}
	}
	s.mu.RUnlock()

	if txnID == "" {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	resp, err := s.ProcessPayment(r.Context(), txnID, txnRef, success, "00")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Simulator) handleQueryTransaction(w http.ResponseWriter, r *http.Request) {
	var req domain.VNPayQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.QueryTransaction(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Simulator) handleRefundTransaction(w http.ResponseWriter, r *http.Request) {
	var req domain.VNPayRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.RefundTransaction(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Simulator) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req domain.VNPayTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.CreateToken(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Simulator) handlePayWithToken(w http.ResponseWriter, r *http.Request) {
	var req domain.VNPayTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.PayWithToken(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Simulator) handleRemoveToken(w http.ResponseWriter, r *http.Request) {
	var req domain.VNPayTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.RemoveToken(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}