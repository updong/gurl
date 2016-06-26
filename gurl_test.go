package main

import (
    "testing"
)

func TestReverse(t *testing.T) {
    expected := "321"
    actual := reverse("123")
    if actual != expected {
        t.Errorf("actual (%s) expected (%s)", actual, expected)
    }
}
