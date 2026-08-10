// Command checkout demonstrates the core payment-collection flow against the
// Bachs sandbox: it creates a product, creates a checkout session for it, and
// prints the hosted checkout URL.
//
// Run it with a sandbox key:
//
//	BACHS_API_KEY=sk_sandbox_... go run ./examples/checkout
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/HalxDocs/bachs-go"
)

func main() {
	apiKey := os.Getenv("BACHS_API_KEY")
	if apiKey == "" {
		log.Fatal("set BACHS_API_KEY to a sandbox key (sk_sandbox_...) before running")
	}

	ctx := context.Background()

	// NewClient defaults to the sandbox environment. Use bachs.WithProduction
	// only when you are ready to take real money.
	client, err := bachs.NewClient(apiKey)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	// 1. Create a product. A billing_cycle makes it recurring; omit it for a
	// one-time product.
	product, _, err := client.Products.Create(ctx, bachs.CreateProductRequest{
		Name: "Pro plan",
		Price: bachs.ProductPrice{
			Currency:  "USD",
			PriceType: "fixed",
			Amount:    "29.00",
		},
		BillingCycle: &bachs.Cadence{Interval: "month", Frequency: 1},
	})
	if err != nil {
		log.Fatalf("create product: %v", err)
	}
	fmt.Printf("created product %s (%s)\n", product.ID, product.Name)

	// 2. Create a checkout session that sells one unit of the product. Bachs
	// returns a hosted URL; send the customer there to pay.
	session, _, err := client.Checkouts.Create(ctx, bachs.CreateCheckoutSessionRequest{
		Customer: bachs.CheckoutCustomer{
			Email: "customer@example.com",
			Name:  "John Doe",
		},
		ProductCart: []bachs.ProductItemRequest{
			{ProductID: product.ID, Quantity: 1},
		},
		SuccessURL: "https://shop.example.com/success",
	}, bachs.WithIdempotencyKey("checkout_demo_attempt_1"))
	if err != nil {
		log.Fatalf("create checkout session: %v", err)
	}
	fmt.Printf("checkout URL: %s\n", session.CheckoutURL)
	fmt.Printf("checkout id: %s (status %s)\n", session.CheckoutID, session.Status)
}
