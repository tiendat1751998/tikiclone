package domain

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	VNPayVersion = "2.1.0"
)

var VNPayResponseCodes = map[string]string{
	"00": "Success",
	"05": "Insufficient balance",
	"06": "Invalid OTP",
	"07": "Suspected fraud",
	"09": "Card not registered for InternetBanking",
	"10": "Incorrect card/account info (3 times)",
	"11": "Payment expired",
	"12": "Card/account locked",
	"24": "User cancelled transaction",
	"79": "Too many authentication attempts",
	"65": "Daily transaction limit exceeded",
	"97": "Invalid signature",
	"91": "Transaction not found",
	"99": "Other errors",
}

type VNPayPaymentMethod string

const (
	VNPayMethodATM      VNPayPaymentMethod = "ATM"
	VNPayMethodIB       VNPayPaymentMethod = "IB"
	VNPayMethodACC      VNPayPaymentMethod = "ACC"
	VNPayMethodQRCODE   VNPayPaymentMethod = "QRCODE"
	VNPayMethodCreditCard VNPayPaymentMethod = "credit_card"
	VNPayMethodQRCODEDirect VNPayPaymentMethod = "qr_code"
)

type VNPayPaymentRequest struct {
	Version      string             `json:"vnp_Version"`
	Command      string             `json:"vnp_Command"`
	TmnCode      string             `json:"vnp_TmnCode"`
	Amount       int64              `json:"vnp_Amount"`
	CurrCode     string             `json:"vnp_CurrCode"`
	BankCode     string             `json:"vnp_BankCode,omitempty"`
	Locale       string             `json:"vnp_Locale"`
	TxnRef       string             `json:"vnp_TxnRef"`
	OrderInfo    string             `json:"vnp_OrderInfo"`
	OrderType    string             `json:"vnp_OrderType"`
	ReturnUrl    string             `json:"vnp_ReturnUrl"`
	IpAddr       string             `json:"vnp_IpAddr"`
	CreateDate   string             `json:"vnp_CreateDate"`
	ExpireDate   string             `json:"vnp_ExpireDate,omitempty"`
	BillMobile   string             `json:"vnp_Bill_Mobile,omitempty"`
	BillEmail    string             `json:"vnp_Bill_Email,omitempty"`
	BillFirstName string            `json:"vnp_Bill_FirstName,omitempty"`
	BillLastName  string            `json:"vnp_Bill_LastName,omitempty"`
	BillAddress   string            `json:"vnp_Bill_Address,omitempty"`
	BillCity      string            `json:"vnp_Bill_City,omitempty"`
	BillCountry   string            `json:"vnp_Bill_Country,omitempty"`
	BillState     string            `json:"vnp_Bill_State,omitempty"`
	InvPhone      string            `json:"vnp_Inv_Phone,omitempty"`
	InvEmail      string            `json:"vnp_Inv_Email,omitempty"`
	InvCustomer   string            `json:"vnp_Inv_Customer,omitempty"`
	InvAddress    string            `json:"vnp_Inv_Address,omitempty"`
	InvCompany    string            `json:"vnp_Inv_Company,omitempty"`
	InvTaxcode    string            `json:"vnp_Inv_Taxcode,omitempty"`
	InvType       string            `json:"vnp_Inv_Type,omitempty"`
}

type VNPayPaymentResponse struct {
	TmnCode        string `json:"vnp_TmnCode"`
	TxnRef         string `json:"vnp_TxnRef"`
	Amount         int64  `json:"vnp_Amount"`
	OrderInfo      string `json:"vnp_OrderInfo"`
	ResponseCode   string `json:"vnp_ResponseCode"`
	BankCode       string `json:"vnp_BankCode,omitempty"`
	BankTranNo     string `json:"vnp_BankTranNo,omitempty"`
	CardType       string `json:"vnp_CardType,omitempty"`
	PayDate        string `json:"vnp_PayDate"`
	TransactionNo  string `json:"vnp_TransactionNo,omitempty"`
	TransactionStatus string `json:"vnp_TransactionStatus"`
	SecureHash     string `json:"vnp_SecureHash"`
}

type VNPayQueryRequest struct {
	RequestId       string `json:"vnp_RequestId"`
	Version         string `json:"vnp_Version"`
	Command         string `json:"vnp_Command"`
	TmnCode         string `json:"vnp_TmnCode"`
	TxnRef          string `json:"vnp_TxnRef"`
	TransactionNo   string `json:"vnp_TransactionNo,omitempty"`
	TransactionDate string `json:"vnp_TransactionDate,omitempty"`
	OrderInfo       string `json:"vnp_OrderInfo"`
	CreateDate      string `json:"vnp_CreateDate"`
	IpAddr          string `json:"vnp_IpAddr"`
}

type VNPayQueryResponse struct {
	ResponseId      string `json:"vnp_ResponseId"`
	Command         string `json:"vnp_Command"`
	TmnCode         string `json:"vnp_TmnCode"`
	TxnRef          string `json:"vnp_TxnRef"`
	Amount          int64  `json:"vnp_Amount"`
	OrderInfo       string `json:"vnp_OrderInfo"`
	ResponseCode    string `json:"vnp_ResponseCode"`
	Message         string `json:"vnp_Message"`
	BankCode        string `json:"vnp_BankCode,omitempty"`
	CardType        string `json:"vnp_CardType,omitempty"`
	PayDate         string `json:"vnp_PayDate,omitempty"`
	TransactionNo   string `json:"vnp_TransactionNo,omitempty"`
	TransactionType string `json:"vnp_TransactionType,omitempty"`
	TransactionStatus string `json:"vnp_TransactionStatus,omitempty"`
	SecureHash      string `json:"vnp_SecureHash"`
}

type VNPayRefundRequest struct {
	RequestId       string `json:"vnp_RequestId"`
	Version         string `json:"vnp_Version"`
	Command         string `json:"vnp_Command"`
	TmnCode         string `json:"vnp_TmnCode"`
	TransactionType string `json:"vnp_TransactionType"`
	TxnRef          string `json:"vnp_TxnRef"`
	Amount          int64  `json:"vnp_Amount"`
	OrderInfo       string `json:"vnp_OrderInfo"`
	TransactionNo   string `json:"vnp_TransactionNo,omitempty"`
	TransactionDate string `json:"vnp_TransactionDate,omitempty"`
	CreateBy        string `json:"vnp_CreateBy"`
	CreateDate      string `json:"vnp_CreateDate"`
	IpAddr          string `json:"vnp_IpAddr"`
}

type VNPayRefundResponse struct {
	ResponseId      string `json:"vnp_ResponseId"`
	Command         string `json:"vnp_Command"`
	TmnCode         string `json:"vnp_TmnCode"`
	TxnRef          string `json:"vnp_TxnRef"`
	Amount          int64  `json:"vnp_Amount"`
	OrderInfo       string `json:"vnp_OrderInfo"`
	ResponseCode    string `json:"vnp_ResponseCode"`
	Message         string `json:"vnp_Message"`
	BankCode        string `json:"vnp_BankCode"`
	PayDate         string `json:"vnp_PayDate,omitempty"`
	TransactionNo   string `json:"vnp_TransactionNo"`
	TransactionType string `json:"vnp_TransactionType"`
	TransactionStatus string `json:"vnp_TransactionStatus"`
	SecureHash      string `json:"vnp_SecureHash"`
}

type VNPayTokenRequest struct {
	Version     string `json:"vnp_Version"`
	Command     string `json:"vnp_Command"`
	TmnCode     string `json:"vnp_TmnCode"`
	AppUserID   string `json:"vnp_AppUserId"`
	BankCode    string `json:"vnp_BankCode,omitempty"`
	Locale      string `json:"vnp_Locale"`
	CardType    string `json:"vnp_CardType"`
	TxnRef      string `json:"vnp_TxnRef"`
	TxnDesc     string `json:"vnp_TxnDesc,omitempty"`
	Token       string `json:"vnp_Token,omitempty"`
	Amount      int64  `json:"vnp_Amount,omitempty"`
	CurrCode    string `json:"vnp_CurrCode,omitempty"`
	ReturnUrl   string `json:"vnp_ReturnUrl,omitempty"`
	CancelUrl   string `json:"vnp_CancelUrl,omitempty"`
	IpAddr      string `json:"vnp_IpAddr"`
	CreateDate  string `json:"vnp_CreateDate"`
	StoreToken  string `json:"vnp_StoreToken,omitempty"`
}

type VNPayTokenResponse struct {
	AppUserID        string `json:"vnp_AppUserId"`
	Token            string `json:"vnp_Token,omitempty"`
	CardNumber       string `json:"vnp_CardNumber,omitempty"`
	Command          string `json:"vnp_Command"`
	TmnCode          string `json:"vnp_TmnCode"`
	ResponseCode     string `json:"vnp_ResponseCode"`
	Message          string `json:"vnp_Message,omitempty"`
	TxnRef           string `json:"vnp_TxnRef"`
	TxnDesc          string `json:"vnp_TxnDesc,omitempty"`
	TransactionNo    string `json:"vnp_TransactionNo,omitempty"`
	CardType         string `json:"vnp_CardType,omitempty"`
	BankCode         string `json:"vnp_BankCode,omitempty"`
	BankTranNo       string `json:"vnp_BankTranNo,omitempty"`
	TransactionStatus string `json:"vnp_TransactionStatus,omitempty"`
	PayDate          string `json:"vnp_PayDate,omitempty"`
	SecureHash       string `json:"vnp_SecureHash"`
}

func GenerateVNPaySecureHash(params map[string]string, secretKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if v := params[k]; v != "" {
			if i > 0 { sb.WriteString("&") }
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
		}
	}

	data := sb.String() + secretKey
	h := hmac.New(sha512.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateVNPaySecureHashByPipe(dataParts []string, secretKey string) string {
	data := strings.Join(dataParts, "|") + secretKey
	h := hmac.New(sha512.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateVNPaySecureHash(params map[string]string, secretKey, signature string) bool {
	expected := GenerateVNPaySecureHash(params, secretKey)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func ParseVNPayParams(params map[string]string) (*VNPayPaymentRequest, error) {
	req := &VNPayPaymentRequest{
		Version:    params["vnp_Version"],
		Command:    params["vnp_Command"],
		TmnCode:    params["vnp_TmnCode"],
		Locale:     params["vnp_Locale"],
		TxnRef:     params["vnp_TxnRef"],
		OrderInfo:  params["vnp_OrderInfo"],
		ReturnUrl:  params["vnp_ReturnUrl"],
		IpAddr:     params["vnp_IpAddr"],
		CreateDate: params["vnp_CreateDate"],
		BillMobile: params["vnp_Bill_Mobile"],
		BillEmail:  params["vnp_Bill_Email"],
	}

	if amountStr := params["vnp_Amount"]; amountStr != "" {
		var amount int64
		if _, err := fmt.Sscanf(amountStr, "%d", &amount); err == nil {
			req.Amount = amount
		}
	}

	req.BankCode = params["vnp_BankCode"]
	req.CurrCode = params["vnp_CurrCode"]
	req.OrderType = params["vnp_OrderType"]
	req.ExpireDate = params["vnp_ExpireDate"]
	req.BillFirstName = params["vnp_Bill_FirstName"]
	req.BillLastName = params["vnp_Bill_LastName"]
	req.BillAddress = params["vnp_Bill_Address"]
	req.BillCity = params["vnp_Bill_City"]
	req.BillCountry = params["vnp_Bill_Country"]
	req.BillState = params["vnp_Bill_State"]
	req.InvPhone = params["vnp_Inv_Phone"]
	req.InvEmail = params["vnp_Inv_Email"]
	req.InvCustomer = params["vnp_Inv_Customer"]
	req.InvAddress = params["vnp_Inv_Address"]
	req.InvCompany = params["vnp_Inv_Company"]
	req.InvTaxcode = params["vnp_Inv_Taxcode"]
	req.InvType = params["vnp_Inv_Type"]

	if req.Version == "" {
		req.Version = VNPayVersion
	}
	if req.Locale == "" {
		req.Locale = "vn"
	}
	if req.CurrCode == "" {
		req.CurrCode = "VND"
	}

	return req, nil
}

func FormatVNPayDate(t time.Time) string {
	return t.Format("20060102150405")
}

func ParseVNPayDateTime(s string) (time.Time, error) {
	return time.Parse("20060102150405", s)
}

func NewVNPayTransactionNo() string {
	return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
}

func BuildVNPayQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(pairs, "&")
}