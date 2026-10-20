package notify

import "testing"

func TestOrderHandoffChannel(t *testing.T) {
	cases := []struct{ name, buyer, want string }{{"buyer channel", "b-17", "buyer-b-17"}}
	for _, tc := range cases {
		n := OrderHandoff{BuyerID: tc.buyer}
		got := "buyer-" + n.BuyerID
		if got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}
