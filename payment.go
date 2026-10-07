package bachs

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PaymentService provides methods for retrieving payments. A payment records
// money a customer paid you through a checkout, including the amount, status,
// fees, the customer, and any refunds against it. Source:
// https://docs.bachs.io/api-reference/payments/object
type PaymentService struct {
	service
}

// Payment is the full payment object returned by Payments.Get.
type Payment struct {
	// Reference is the checkout reference when available.
	Reference *string `json:"reference"`

	// PaymentID is the unique identifier for the payment.
	PaymentID string `json:"payment_id"`

	// BillingReason is why this payment exists: "purchase",
	// "subscription_create", "subscription_cycle", or "subscription_update".
	BillingReason string `json:"billing_reason"`

	// CheckoutID is the checkout that created the payment, when linked.
	CheckoutID *string `json:"checkout_id"`

	// Status is the payment lifecycle, for example "succeeded", "failed", or
	// "refunded".
	Status string `json:"status"`

	// IsRefundable reports whether the payment can currently be refunded.
	IsRefundable *bool `json:"is_refundable"`

	// Amount is the requested amount in Currency.
	Amount string `json:"amount"`

	// AmountPaid is the amount received so far.
	AmountPaid *string `json:"amount_paid"`

	// AmountRemaining is the amount still expected.
	AmountRemaining *string `json:"amount_remaining"`

	// Currency is the payment currency code.
	Currency string `json:"currency"`

	// FeeUSD is the processing fee converted to USD; null until the payment
	// settles.
	FeeUSD *string `json:"fee_usd"`

	// MerchantBearsCost reports whether the merchant bears the processing
	// cost.
	MerchantBearsCost *bool `json:"merchant_bears_cost"`

	// PaymentMethod used for this payment. For a virtual account deposit this
	// reports the rail used (for example "NGN_BANK_TRANSFER"); a card
	// payment reports "USD_CARD" or "NGN_CARD".
	PaymentMethod *string `json:"payment_method"`

	// PaymentMethodDetails carries payer details for methods that provide
	// them, for example the sender of a bank transfer. Null on payments
	// taken before this shipped, on methods that provide no payer details,
	// and on the payment list response (it is returned by Payments.Get and
	// sent on collection.succeeded only). Use it to distinguish deposits
	// into the same virtual account. Source: the Week of Sep 21, 2026 entry
	// in https://docs.bachs.io/changelog/api.
	PaymentMethodDetails *PaymentMethodDetails `json:"payment_method_details"`

	// Channel is the origin channel (for example "api").
	Channel *string `json:"channel"`

	// Narration is the payment description.
	Narration *string `json:"narration"`

	// Meta is the public metadata stored for this payment.
	Meta map[string]any `json:"meta"`

	// Message is a human-readable payment message derived from Status.
	Message *string `json:"message"`

	// Customer information when available.
	Customer *PaymentCustomer `json:"customer"`

	// LineItems are the products this payment covers.
	LineItems []PaymentProductItem `json:"line_items"`

	// SubscriptionID is the subscription this payment belongs to, or null for
	// a one-time purchase.
	SubscriptionID *string `json:"subscription_id"`

	// Invoice is present only for subscription payments; null for one-time
	// purchases.
	Invoice *PaymentInvoiceInfo `json:"invoice"`

	// Refunds are the IDs of any refunds issued for this payment. Null when
	// none exist.
	Refunds []string `json:"refunds"`

	// PlatformFee echoes the fee-first split sent at creation. Null when
	// the sale carries no split — or when it was share-first.
	PlatformFee *string `json:"platform_fee"`

	// DestinationAmount echoes the share-first split sent at creation. Null
	// when the sale carries no split — or when it was fee-first.
	DestinationAmount *string `json:"destination_amount"`

	// StatusHistory is the chronological list of status changes.
	StatusHistory []PaymentStatusHistory `json:"status_history"`

	// CreatedAt is when the payment was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the payment was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when the payment reached a successful terminal state.
	CompletedAt *time.Time `json:"completed_at"`
}

// PaymentMethodDetails carries payer details for the method a payment used.
// It is null when the method provides no payer details.
type PaymentMethodDetails struct {
	// Type is the method the details describe (for example
	// "bank_transfer").
	Type string `json:"type"`

	// BankTransfer holds the sender details for a bank transfer payment.
	// Null for other methods.
	BankTransfer *PaymentBankTransferDetails `json:"bank_transfer,omitempty"`
}

// PaymentBankTransferDetails names the sender of a bank transfer payment:
// the sender's name, bank, and account number, the interbank session ID, and
// the account number that received the money.
type PaymentBankTransferDetails struct {
	// SenderName of the payer, as reported by the sending bank.
	SenderName *string `json:"sender_name"`

	// SenderBank of the payer, as reported by the sending bank.
	SenderBank *string `json:"sender_bank"`

	// SenderAccountNumber of the payer, as reported by the sending bank.
	SenderAccountNumber *string `json:"sender_account_number"`

	// SessionID is the interbank session ID of the transfer.
	SessionID *string `json:"session_id"`

	// VirtualAccount is the account number that received the money, when the
	// deposit landed in a virtual account.
	VirtualAccount *PaymentVirtualAccountRef `json:"virtual_account"`
}

// PaymentVirtualAccountRef identifies the virtual account that received a
// bank transfer deposit.
type PaymentVirtualAccountRef struct {
	// ID of the virtual account (for example "va_...").
	ID string `json:"id"`

	// AccountNumber that received the money.
	AccountNumber string `json:"account_number"`

	// BankName holding the number.
	BankName string `json:"bank_name"`

	// Type of the virtual account (for example "permanent").
	Type *string `json:"type"`

	// ExpiresAt is when the number stops working. Null for a permanent
	// (fixed) virtual account.
	ExpiresAt *time.Time `json:"expires_at"`
}

// PaymentCustomer is the customer information attached to a payment.
type PaymentCustomer struct {
	// Name of the customer, when captured.
	Name *string `json:"name"`

	// Email of the customer, when captured.
	Email *string `json:"email"`
}

// PaymentProductItem is one line item a payment covers.
type PaymentProductItem struct {
	// ProductID identifies the product.
	ProductID string `json:"product_id"`

	// ProductName is the product's display name.
	ProductName string `json:"product_name"`

	// Quantity is the number of units purchased.
	Quantity int `json:"quantity"`

	// UnitAmount is the price per unit in Currency.
	UnitAmount string `json:"unit_amount"`

	// Currency is the currency code for this line item.
	Currency string `json:"currency"`

	// LineTotal is UnitAmount times Quantity.
	LineTotal string `json:"line_total"`
}

// PaymentInvoiceInfo is the invoice a subscription payment collected.
type PaymentInvoiceInfo struct {
	// InvoiceID identifies the invoice.
	InvoiceID string `json:"invoice_id"`

	// Number is the human-facing invoice number, if assigned.
	Number *string `json:"number"`

	// SubscriptionID is the subscription the invoice belongs to.
	SubscriptionID *string `json:"subscription_id"`

	// PeriodStart is the start of the billing period, UTC.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the billing period, UTC.
	PeriodEnd time.Time `json:"period_end"`

	// Kind is "cycle" (a regular period invoice) or "proration" (an off-cycle
	// mid-cycle change).
	Kind string `json:"kind"`
}

// PaymentStatusHistory is one status transition in a payment's history.
type PaymentStatusHistory struct {
	// Status at this point in time.
	Status string `json:"status"`

	// OccurredAt is when the status change occurred.
	OccurredAt time.Time `json:"occurred_at"`

	// ProviderReference is the provider-side reference for the transition, if
	// available.
	ProviderReference *string `json:"provider_reference"`

	// Reason is a human-readable reason for the change, if available.
	Reason *string `json:"reason"`
}

// PaymentListItem is the summary form of a payment returned by
// Payments.List.
type PaymentListItem struct {
	// Reference is the checkout reference when available.
	Reference *string `json:"reference"`

	// ID is the payment ID for retrieval and reconciliation.
	ID *string `json:"id"`

	// Status is the current payment status.
	Status string `json:"status"`

	// IsRefundable reports whether the payment is currently refundable.
	IsRefundable *bool `json:"is_refundable"`

	// Amount is the requested payment amount in Currency.
	Amount string `json:"amount"`

	// CustomerName from checkout data. May be empty when unavailable.
	CustomerName string `json:"customer_name"`

	// CustomerEmail from checkout data. May be empty when unavailable.
	CustomerEmail string `json:"customer_email"`

	// AmountPaid is the amount received so far.
	AmountPaid *string `json:"amount_paid"`

	// AmountRemaining is the amount still expected.
	AmountRemaining *string `json:"amount_remaining"`

	// SettlementAmount is the settlement-side amount captured.
	SettlementAmount *string `json:"settlement_amount"`

	// SettlementCurrency is the currency of SettlementAmount.
	SettlementCurrency *string `json:"settlement_currency"`

	// Fee is reserved for fees in list responses. May be null.
	Fee *string `json:"fee"`

	// Vat is reserved for VAT in list responses. May be null.
	Vat *string `json:"vat"`

	// Currency is the payment currency code.
	Currency string `json:"currency"`

	// Meta is the public metadata attached to the payment, when available.
	Meta map[string]any `json:"meta"`

	// TransactionDate is when the payment was created.
	TransactionDate *time.Time `json:"transaction_date"`

	// CompletedAt is when the payment reached a successful terminal state.
	CompletedAt *time.Time `json:"completed_at"`
}

// Get retrieves a single payment by its ID, with the full object: amount,
// status, the customer, fees, the products paid for, refunds, and status
// history.
func (s *PaymentService) Get(ctx context.Context, paymentID string) (*Payment, *ResponseMeta, error) {
	var out Payment
	meta, err := s.request(ctx, http.MethodGet, "/payments/"+url.PathEscape(paymentID), nil, &out)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// List returns a paginated list of payments, newest first. Set
// ListParams.StatusFilter (for example "succeeded" or "failed") to filter.
func (s *PaymentService) List(ctx context.Context, params ListParams) (*Page[PaymentListItem], *ResponseMeta, error) {
	var env pageEnvelope[PaymentListItem]
	meta, err := s.request(ctx, http.MethodGet, queryPath("/payments", params), nil, &env)
	if err != nil {
		return nil, meta, err
	}
	return env.page(), meta, nil
}
