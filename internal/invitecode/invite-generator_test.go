package invitecode

import (
	"regexp"
	"testing"
)

func TestGenerateNumberStringEmpty(t *testing.T) {
	empty := GenerateNumberString(0)

	if empty != "" {
		t.Errorf(`expected "" but got: %s`, empty)
	}
}

func TestGenerateNumberStringTen(t *testing.T) {
	ten := GenerateNumberString(10)

	if len(ten) != 10 {
		t.Errorf(`expected "" but got %s`, ten)
	}

	if m, err := regexp.Match(`[A-Z,0-9]{10}`, []byte(ten)); !m || err != nil {
		t.Errorf(`generated string did not match regexp but got %s`, ten)
	}
}
