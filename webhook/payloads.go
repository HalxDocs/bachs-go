package webhook

import (
	"encoding/json"
	"time"
)

// DataAs unmarshals the event's Data payload into v, which should be a
// pointer to one of the typed payload structs (for example
// *CollectionSucceededData for a "collection.succeeded" event) or any other
// shape the caller prefers.
func (e Event) DataAs(v any) error {
	return json.Unmarshal(e.Data, v)
}

// CollectionCartItem is one product purchased in a one-time checkout session.
// Amount is present only for custom-priced products.
type CollectionCartItem struct {
	// ProductID identifies the product.
	ProductID string `json:"product_id"`

	// Quantity purchased.
	Quantity int `json:"quantity"`

	// Amount chosen for a custom-priced product. Absent otherwise.
	Amount *string `json:"amount"`
}

// CollectionCustomer is the buyer on a collection event. ID is always
// present but can be null for a guest or a payment received without a
// checkout, such as a virtual account deposit; Email and Name appear only
// when a checkout collected them.
type CollectionCustomer struct {
	// ID of the customer record. Null when no record backs the payment.
	ID *string `json:"id"`

	// Email collected at checkout, when available.
	Email *string `json:"email"`

	// Name collected at checkout, when available.
	Name *string `json:"name"`
}

// CollectionVirtualAccountRef identifies the virtual account that received a
// bank transfer.
type CollectionVirtualAccountRef struct {
	// ID of the virtual account. Null for a one-time account number.
	ID *string `json:"id"`

	// AccountNumber that received the money.
	AccountNumber string `json:"account_number"`

	// BankName holding the number.
	BankName string `json:"bank_name"`

	// Type of the virtual account (for example "permanent" or "one_time").
	Type *string `json:"type"`

	// ExpiresAt is when a one-time number stops working. Null for a
	// permanent (fixed) virtual account.
	ExpiresAt *time.Time `json:"expires_at"`
}

// CollectionBankTransferDetails names how a bank transfer moved: the sender,
// the interbank session, their narration, and the receiving virtual account.
// Treat every nested field as optional — the sending bank may omit it.
type CollectionBankTransferDetails struct {
	// SenderName of the payer, as reported by the sending bank.
	SenderName *string `json:"sender_name"`

	// SenderBank of the payer, as reported by the sending bank.
	SenderBank *string `json:"sender_bank"`

	// SenderBankCode of the payer's bank.
	SenderBankCode *string `json:"sender_bank_code"`

	// SenderAccountNumber of the payer, as reported by the sending bank.
	SenderAccountNumber *string `json:"sender_account_number"`

	// SessionID is the interbank session ID of the transfer.
	SessionID *string `json:"session_id"`

	// Narration the sender attached, when provided.
	Narration *string `json:"narration"`

	// VirtualAccount that received the money, when it landed in one.
	VirtualAccount *CollectionVirtualAccountRef `json:"virtual_account"`
}

// CollectionPaymentMethodDetails describes how the money moved, keyed by
// Type. Null on methods that provide no payer details.
type CollectionPaymentMethodDetails struct {
	// Type is the method the details describe (for example
	// "bank_transfer").
	Type string `json:"type"`

	// BankTransfer holds the sender details for a bank transfer. Null for
	// other methods.
	BankTransfer *CollectionBankTransferDetails `json:"bank_transfer"`
}

// CollectionSucceededData is the typed payload of a "collection.succeeded"
// event: a charge reaching a successful state, including a virtual account
// deposit. ChargeID may be null on test deliveries, legacy payments, and
// manual reconciliations — always guard before using it for lookups. Source:
// https://docs.bachs.io/guides/webhooks/events/collection-succeeded
type CollectionSucceededData struct {
	// ChargeID for reconciliation and retrieval calls. Null in the cases
	// above.
	ChargeID *string `json:"charge_id"`

	// CheckoutID that originated the charge. Null for a fixed virtual
	// account deposit and other payments without a checkout.
	CheckoutID *string `json:"checkout_id"`

	// Reference is the checkout reference supplied at creation, when
	// available.
	Reference *string `json:"reference"`

	// Status is the uppercase charge state: SUCCEEDED, ACCEPTED, or
	// OVERPAID. The payment object returns these lowercase — do not compare
	// the values directly.
	Status string `json:"status"`

	// Amount originally charged, in Currency.
	Amount string `json:"amount"`

	// Currency is the customer payment currency code.
	Currency string `json:"currency"`

	// SettlementAmount credited in SettlementCurrency.
	SettlementAmount *string `json:"settlement_amount"`

	// SettlementCurrency used for the settlement credit.
	SettlementCurrency *string `json:"settlement_currency"`

	// ProcessingFee in ProcessingFeeCurrency. Null when the final
	// settlement value is not yet determined.
	ProcessingFee *string `json:"processing_fee"`

	// ProcessingFeeCurrency, typically the settlement currency.
	ProcessingFeeCurrency *string `json:"processing_fee_currency"`

	// FeeBearer is who absorbed the fee: "customer" or "merchant".
	FeeBearer *string `json:"fee_bearer"`

	// ProductCart lists the products purchased. Null when no checkout
	// created the charge. Each item carries product_id, quantity, and
	// amount only for custom-priced products.
	ProductCart []CollectionCartItem `json:"product_cart"`

	// Customer who made the payment.
	Customer *CollectionCustomer `json:"customer"`

	// PaymentMethodDetails describes how the money moved. Null when the
	// method provides no payer details.
	PaymentMethodDetails *CollectionPaymentMethodDetails `json:"payment_method_details"`

	// Metadata is the public metadata stored with the charge.
	Metadata map[string]any `json:"metadata"`

	// Reason explains why the charge reached this state, when Bachs has
	// something to say. Absent on an ordinary success.
	Reason *string `json:"reason"`

	// ExpectedAmount was due; ReceivedAmount arrived; OverpaidAmount is the
	// difference. Present only when Status is OVERPAID.
	ExpectedAmount *string `json:"expected_amount"`
	ReceivedAmount *string `json:"received_amount"`
	OverpaidAmount *string `json:"overpaid_amount"`
}

// SubscriptionEventBillingAddress is the billing address embedded in
// subscription webhook payloads.
type SubscriptionEventBillingAddress struct {
	// Line1 is the street address.
	Line1 *string `json:"line1"`

	// Line2 is the apartment, suite, unit, etc.
	Line2 *string `json:"line2"`

	// City is the city, district, or suburb.
	City *string `json:"city"`

	// State is the state, province, or region.
	State *string `json:"state"`

	// PostalCode is the ZIP or postal code.
	PostalCode *string `json:"postal_code"`

	// Country is the two-letter ISO 3166-1 alpha-2 country code.
	Country *string `json:"country"`
}

// SubscriptionEventCustomer is the full customer embedded in subscription
// webhook payloads.
type SubscriptionEventCustomer struct {
	// CustomerID is the unique identifier, prefixed with "cust_".
	CustomerID string `json:"customer_id"`

	// Email of the customer.
	Email string `json:"email"`

	// Name is the full name. Null when not set.
	Name *string `json:"name"`

	// PhoneNumber in E.164 format. Null when not set.
	PhoneNumber *string `json:"phone_number"`

	// Metadata attached to the customer.
	Metadata map[string]any `json:"metadata"`

	// BillingAddress, or null when none is set.
	BillingAddress *SubscriptionEventBillingAddress `json:"billing_address"`

	// CreatedAt is when the customer was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the customer was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionEventCadence is the billing cadence in subscription payloads.
type SubscriptionEventCadence struct {
	// Interval is "day", "week", "month", or "year".
	Interval string `json:"interval"`

	// Frequency is the number of intervals per cycle.
	Frequency int `json:"frequency"`
}

// CustomerSubscriptionData is the typed payload of the
// customer.subscription.* events (created, updated, deleted): the
// subscription and the customer behind it. Use it to provision, adjust, or
// remove access to the recurring product. Source:
// https://docs.bachs.io/guides/webhooks/events/customer-subscription-created
type CustomerSubscriptionData struct {
	// SubscriptionID uniquely identifies the subscription.
	SubscriptionID string `json:"subscription_id"`

	// Customer the subscription bills.
	Customer SubscriptionEventCustomer `json:"customer"`

	// ProductID the subscription bills for.
	ProductID string `json:"product_id"`

	// Status is "trialing", "active", "past_due", "unpaid", "canceled", or
	// "paused".
	Status string `json:"status"`

	// CollectionMethod is how renewals are collected (for example
	// "charge_automatically").
	CollectionMethod string `json:"collection_method"`

	// Currency the subscription is billed in.
	Currency string `json:"currency"`

	// Amount is the recurring amount as a decimal string.
	Amount string `json:"amount"`

	// BillingCycle is the recurring cadence.
	BillingCycle SubscriptionEventCadence `json:"billing_cycle"`

	// Quantity billed.
	Quantity int `json:"quantity"`

	// CurrentPeriodStart is the start of the period being billed for, UTC.
	CurrentPeriodStart time.Time `json:"current_period_start"`

	// CurrentPeriodEnd is the end of the period being billed for, UTC.
	CurrentPeriodEnd time.Time `json:"current_period_end"`

	// NextBilledAt is the next scheduled charge date. Null when none is
	// scheduled.
	NextBilledAt *time.Time `json:"next_billed_at"`

	// TrialEnd is when the free trial ends. Null when not trialing.
	TrialEnd *time.Time `json:"trial_end"`

	// CancelAtPeriodEnd is true when the subscription stays active until
	// CurrentPeriodEnd and is not renewed.
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`

	// CanceledAt is when the subscription was canceled. Null otherwise.
	CanceledAt *time.Time `json:"canceled_at"`

	// CreatedAt is when the subscription was created, UTC.
	CreatedAt time.Time `json:"created_at"`

	// Items are the line items being billed. Each entry's shape follows the
	// subscription line items; decode individually when needed.
	Items []map[string]any `json:"items"`

	// Metadata attached to the subscription.
	Metadata map[string]any `json:"metadata"`
}
