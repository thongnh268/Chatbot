package action

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/subiz/header"
	pb "github.com/subiz/header/common"
	"google.golang.org/protobuf/proto"
)

func FirstMatchCondition(node ActionNode, srcs []*InCondition) []string {
	nextActions := node.GetAction().GetNexts()
	for _, n := range nextActions {
		if n.GetAction().GetId() == "" {
			continue
		}
		if n.GetCondition() == nil || proto.Equal(n.GetCondition(), Condition_empty) {
			return []string{n.GetAction().GetId()}
		}
		if matchCondition(n.GetCondition(), srcs) {
			return []string{n.GetAction().GetId()}
		}
	}
	return Str_empty_arr
}

func matchCondition(cond *header.Condition, srcs []*InCondition) bool {
	if cond.Group == header.Condition_single.String() || cond.Group == "" {
		for _, row := range srcs {
			if compareCond(row, cond) {
				return true // any true in srcs
			}
		}
	}

	if cond.Group == header.Condition_any.String() {
		// valid with first cond
		for _, c := range cond.Conditions {
			subIs := matchCondition(c, srcs)
			if subIs {
				return true
			}
		}
	}

	if cond.Group == header.Condition_all.String() {
		for _, c := range cond.Conditions {
			subIs := matchCondition(c, srcs)
			if !subIs {
				return false
			}
		}
		return true
	}
	return false
}

func compareCond(src *InCondition, dst *header.Condition) bool {
	if src == nil && dst == nil {
		return true
	}
	if src == nil {
		return false
	}
	if dst == nil {
		return true
	}
	srcType := src.Type
	srcValue := src.Value
	if dst.TransformFunction != "" && src.Convs != nil {
		convf, has := src.Convs[dst.TransformFunction]
		if has {
			srcType, srcValue = convf(src)
		}
	}
	if srcType != dst.Type {
		return false
	}
	is := false
	switch srcType {
	case "string":
		str, ok := srcValue.(string)
		is = ok && compareStringParams(dst.GetString_(), str)
	case "number":
		num, ok := srcValue.(float32)
		is = ok && compareNumberParams(dst.GetNumber(), num)
	case "boolean":
		b, ok := srcValue.(bool)
		is = ok && dst.GetBoolean().GetIs() == b
	}
	return is
}

func compareNumberParams(intPar *pb.NumberParams, value float32) bool {
	if intPar.GetEq() != 0 && value != intPar.GetEq() {
		return false
	}
	if intPar.GetNeq() != 0 && value == intPar.GetNeq() {
		return false
	}
	if intPar.GetGt() != 0 && value <= intPar.GetGt() {
		return false
	}
	if intPar.GetGte() != 0 && value < intPar.GetGte() {
		return false
	}
	if intPar.GetLt() != 0 && value >= intPar.GetLt() {
		return false
	}
	if intPar.GetLte() != 0 && value > intPar.GetLte() {
		return false
	}

	return true
}

func compareStringParams(strPar *pb.StringParams, value string) bool {
	if strPar.GetEq() != "" && !compareString(value, strPar.GetEq()) {
		return false
	}
	if len(strPar.GetIn()) > 0 && !stringInSlice(value, strPar.GetIn()) {
		return false
	}
	if strPar.GetContain() != "" && !containString(value, strPar.GetContain()) {
		return false
	}

	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if compareString(b, a) {
			return true
		}
	}
	return false
}

// require trimspace
// require lower
func strToFilter(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
func compareString(str1, str2 string) bool {
	return strToFilter(str1) == strToFilter(str2)
}
func containString(s, substr string) bool {
	return strings.Contains(strToFilter(s), strToFilter(substr))
}

type InCondition struct {
	Key   string
	Type  string
	Value interface{}
	Convs map[string]func(data *InCondition) (string, interface{})
	Req   *header.RunRequest
}

var Condition_conv_invalid = map[string]func(*InCondition) (string, interface{}){
	"valid": func(data *InCondition) (string, interface{}) {
		return "boolean", false
	},
}
var Condition_conv_valid = map[string]func(*InCondition) (string, interface{}){
	"valid": func(data *InCondition) (string, interface{}) {
		return "boolean", true
	},
}
var Condition_empty = &header.Condition{}

type RegexpValidation struct {
	Name           string
	Regexp         *regexp.Regexp
	InvalidMessage []string
	LastMatch      func(validation *RegexpValidation, str string) string // TODO self
}

var RegexpValidation_arr []*RegexpValidation
var RegexpValidation_map map[string]*RegexpValidation

func LastMatch(validation *RegexpValidation, str string) string {
	arr := validation.Regexp.FindAllString(str, -1)
	if len(arr) == 0 {
		return ""
	}

	return arr[len(arr)-1]
}

func LastMatchPhone(validation *RegexpValidation, str string) string {
	noneSpaceStr := removePhoneSpace(str)
	arr := validation.Regexp.FindAllString(noneSpaceStr, -1)
	if len(arr) == 0 {
		return ""
	}

	return arr[len(arr)-1]
}

func removePhoneSpace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) && ch != '.' && ch != '-' && ch != '_' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

const (
	RegexEmail  = `([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+)`
	RegexNumber = `([0-9._-]+)`
	RegexDate   = `(0?[1-9]|[12][0-9]|3[01])/(0?[1-9]|1[012])/((19|20)\\d\\d)`
)

// TODO add more field of node
func init() {
	InvalidMessagePhone := []string{
		"Số điện thoại không đúng rồi. Gửi lại cho mình nhé",
		"Mình không thấy số điện thoại của bạn. Bạn gửi lại được chứ?",
		"Để lại số điện thoại sẽ giúp mình tư vấn bạn nhanh hơn đó ạ.",
		"Bạn ơi, vẫn thiếu số điện thoại. Bạn cung cấp thêm cho mình với.",
		"Mình vẫn chưa thấy số điện thoại, bạn kiểm tra và gửi lại nhé.",
		"Ôi số điện thoại không đúng rồi :(( Bạn thử lại đi ạ.",
		"Số điện thoại luôn được bảo mật. Bạn cứ yên tâm gửi mình nhé.",
	}
	InvalidMessageEmail := []string{
		"Email không đúng rồi. Gửi lại cho mình nhé.",
		"Mình không tìm thấy email của bạn. Bạn gửi lại được chứ?",
		"Để lại email sẽ giúp mình tư vấn bạn tốt hơn đó ạ",
		" Bạn ơi, vẫn còn thiếu email nữa. Bạn cung cấp thêm cho mình với.",
		"Mình vẫn chưa thấy địa chỉ email, bạn kiểm tra và gửi lại nhé.",
		"Ôi email không đúng rồi :(( Bạn thử lại đi ạ.",
		"Email và mọi thông tin luôn được bảo mật. Bạn cứ yên tâm gửi mình nhé.",
	}
	regexpEmail, err := regexp.Compile(RegexEmail)
	if err != nil {
		panic(err)
	}
	regexpPhone, err := regexp.Compile(RegexNumber)
	if err != nil {
		panic(err)
	}
	regexpNumber, err := regexp.Compile(RegexNumber)
	if err != nil {
		panic(err)
	}
	regexpDate, err := regexp.Compile(RegexDate)
	if err != nil {
		panic(err)
	}
	RegexpValidation_arr = []*RegexpValidation{
		{
			Name:           "email",
			Regexp:         regexpEmail,
			InvalidMessage: InvalidMessageEmail,
			LastMatch:      LastMatch,
		},
		{
			Name:           "phone",
			Regexp:         regexpPhone,
			InvalidMessage: InvalidMessagePhone,
			LastMatch:      LastMatchPhone,
		},
		{
			Name:           "number",
			Regexp:         regexpNumber,
			InvalidMessage: []string{"Không tìm thấy số được nhập, vui lòng nhập lại"},
			LastMatch:      LastMatch,
		},
		{
			Name:           "date",
			Regexp:         regexpDate,
			InvalidMessage: []string{"Vui lòng nhập lại theo định dạng dd/mm/yyyy"},
			LastMatch:      LastMatch,
		},
	}
	RegexpValidation_map = make(map[string]*RegexpValidation)
	for _, validation := range RegexpValidation_arr {
		RegexpValidation_map[validation.Name] = validation
	}
}
