package action

import (
	"testing"

	"github.com/subiz/header"
	"github.com/subiz/header/common"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestLongestMatch(t *testing.T) {
	var exo, aco string
	var inStr string
	var inValidation *RegexpValidation

	inStr = "test@mail.com"
	exo = "test@mail.com"
	inValidation = RegexpValidation_map["email"]
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inStr = "test1@mail.com test2@mail.com"
	exo = "test2@mail.com"
	inValidation = RegexpValidation_map["email"]
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inStr = " abc test@mail.com -./;\\'validemail@mail.com 123 ,,,"
	exo = "validemail@mail.com"
	inValidation = RegexpValidation_map["email"]
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inStr = "email cua toi la ducck92@gmail.com"
	exo = "ducck92@gmail.com"
	inValidation = RegexpValidation_map["email"]
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inValidation = RegexpValidation_map["phone"]
	inStr = "0123456789"
	exo = "0123456789"
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inValidation = RegexpValidation_map["phone"]
	inStr = "0123456789, 9876543210"
	exo = "9876543210"
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inValidation = RegexpValidation_map["phone"]
	inStr = " abc 123123 -./;\\'0123 456 789 abc ,,,"
	exo = "0123456789"
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inValidation = RegexpValidation_map["phone"]
	inStr = " abc 123123 -./;\\'0123 456 789 abc .-_"
	exo = "0123456789"
	aco = inValidation.LastMatch(inValidation, inStr)
	if aco != exo {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}

func TestCompareCond(t *testing.T) {
	var exo, aco bool
	var inCond *header.Condition
	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: true},
	}
	exo = true
	aco = compareCond(&InCondition{Type: "boolean", Key: "last_response", Value: true}, inCond)
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}

	inCond = &header.Condition{
		Group:   header.Condition_single.String(),
		Type:    "boolean",
		Key:     "last_response",
		Boolean: &common.BooleanParams{Is: false},
	}
	exo = false
	aco = compareCond(&InCondition{}, inCond)
	if exo != aco {
		t.Errorf("want %#v, actual %#v", exo, aco)
	}
}
