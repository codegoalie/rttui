package rtm

import "testing"

func TestSign(t *testing.T) {
	// Example from RTM API docs: shared secret "BANANAS", params a=1, b=2, c=3
	// sorted: a1b2c3 → prepend secret → "BANANASa1b2c3" → MD5
	params := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	got := sign("BANANAS", params)
	// pre-computed: echo -n "BANANASa1b2c3" | md5sum
	want := "e596fc0195c86fe958f509943c361754"
	if got != want {
		t.Errorf("sign() = %q, want %q", got, want)
	}
}
