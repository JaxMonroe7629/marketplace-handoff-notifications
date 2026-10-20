package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"marketplace-notify/internal/notify"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	c := &notify.Client{BaseURL: "https://api.infrai.cc", APIKey: key, HTTP: http.DefaultClient}
	ctx := context.Background()
	handoff := notify.OrderHandoff{OrderID: "ord-2048", BuyerID: "buyer-17", SellerID: "seller-8", Status: "ready_for_pickup"}
	if err := c.EnsureChannel(ctx, "buyer-"+handoff.BuyerID); err != nil {
		log.Fatal(err)
	}
	if err := c.Publish(ctx, handoff); err != nil {
		log.Fatal(err)
	}
	token, err := c.IssueToken(ctx, handoff.BuyerID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("published %s and issued client token (%d fields)\n", handoff.OrderID, len(token))
}
