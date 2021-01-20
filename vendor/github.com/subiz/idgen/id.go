package idgen

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	USER_PREFIX              = "us"
	ACCOUNT_PREFIX           = "ac"
	AGENT_GROUP_PREFIX       = "gr"
	AGENT_PREFIX             = "ag"
	INVITATION_PREFIX        = "iv"
	EVENT_PREFIX             = "ev"
	CONVERSATION_PREFIX      = "cs"
	SCHEDULE_ITEM_PREFIX     = "si"
	WEBSEND_PREFIX           = "wd"
	AUTH_TOKEN_PREFIX        = "au"
	REFRESH_TOKEN_PREFIX     = "rt"
	RULE_PREFIX              = "ru"
	CLIENT_PREFIX            = "cl"
	ERROR_PREFIX             = "er"
	WEBHOOK_PREFIX           = "wh"
	TAG_PREFIX               = "tg"
	CANNED_RESPONSE_PREFIX   = "cn"
	TEMPLATE_PREFIX          = "tp"
	FILE_PREFIX              = "fi"
	BUTTON_PREFIX            = "bt"
	CHALLENGE_PREFIX         = "ch"
	SEGMENTATION_PREFIX      = "sg"
	SUBIZ_PREFIX             = "su"
	USER_FIELD_PREFIX        = "uf"
	REQUEST_PREFIX           = "rq"
	AUTOMATION_PREFIX        = "at"
	USER_SESSION_PREFIX      = "ss"
	BILLING_PREFIX           = "bi"
	INVOICE_PREFIX           = "ic"
	SUBSCRIPTION_PREFIX      = "sc"
	CUSTOMER_PREFIX          = "cr"
	PAYMENT_LOG_PREFIX       = "pl"
	PAYMENT_METHOD_PREFIX    = "pm"
	ATTRIBUTE_PREFIX         = "ab"
	IDEMPOTENCY_KEY_PREFIX   = "ik"
	PAYMENT_COMMENT_PREFIX   = "cm"
	WS_PREFIX                = "ws"
	USER_NOTE_PREFIX         = "nt"
	TICKET_PREFIX            = "tk"
	PIPELINE_PREFIX          = "pl"
	STAGE_PREFIX             = "st"
	CURRENCY_PREFIX          = "cr"
	EXCHANGE_RATE_PREFIX     = "ex"
	SLA_PREFIX               = "sa"
	REFER_PREFIX             = "rf"
	PROMOTION_CODE_PREFIX    = "pc"
	BOT_REPLY                = "br"
	BIZBOT_PREFIX            = "bb"
	AUTOMATION_ACTION_PREFIX = "aa"
	BACKOFF_PREFIX           = "bo"
	GOALSTATUS_PREFIX        = "gs"
	CONVERSION_PREFIX        = "cv"
	IMPRESSION_PREFIX        = "ip"
	CAMPAIGN_PREFIX          = "cp"
	POLLINGCONNECTION_PREFIX = "co"
)

// New return new random ID
func New() string { return generateID("", 2) }

func init() { rand.Seed(time.Now().UnixNano()) }

// letterRunes (read-only) contains all runes which can be used in an ID
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func generateID(sign string, randomfactor int) string {
	var sb strings.Builder
	sb.WriteString(sign)

	nowstr := formatInt(time.Now().UnixNano(), 26)
	sb.WriteString(nowstr)
	for i := 0; i < randomfactor; i++ {
		sb.WriteRune(letterRunes[rand.Intn(len(letterRunes))])
	}
	return sb.String()
}

func GetCreated(id, prefix string) (int64, error) {
	if len(id)-len(prefix) < 13 {
		return 0, errors.New("id too short")
	}
	// 12 or 13??
	return parseInt26(id[len(prefix) : 13+len(prefix)])
}

func New0() string {
	return generateID("00", 0)
}

func NewBizbotID() string {
	return generateID(BIZBOT_PREFIX, 3)
}

func NewBotID() string {
	return generateID(BOT_REPLY, 3)
}

func NewReferrerID() string {
	return generateID(REFER_PREFIX, 3)
}

func NewAgentGroupID() string {
	return generateID(AGENT_GROUP_PREFIX, 3)
}

func NewAccountID() string {
	return generateID(ACCOUNT_PREFIX, 5)
}

func NewAgentID() string {
	return generateID(AGENT_PREFIX, 3)
}

func NewInvitationID() string {
	return generateID(INVITATION_PREFIX, 4)
}

func NewEventID() string {
	return generateID(EVENT_PREFIX, 10)
}

func NewConversationID() string {
	return generateID(CONVERSATION_PREFIX, 3)
}

func NewScheduleItemID() string {
	return generateID(SCHEDULE_ITEM_PREFIX, 20)
}

func NewPromotionCodeID() string {
	return generateID(PROMOTION_CODE_PREFIX, 2)
}

func NewWebsendID() string {
	return generateID(WEBSEND_PREFIX, 20) // websend
}

// NewAuthToken return new oauth2 authorization token
func NewAuthToken() string {
	return generateID(AUTH_TOKEN_PREFIX, 20)
}

func NewRefreshToken() string {
	return generateID(REFRESH_TOKEN_PREFIX, 30)
}

func NewRuleID() string {
	return generateID(RULE_PREFIX, 6)
}

func NewClientID() string {
	return generateID(CLIENT_PREFIX, 6)
}

func NewErrorID() string {
	return generateID(ERROR_PREFIX, 6)
}

func NewWebhookID() string {
	return generateID(WEBHOOK_PREFIX, 6)
}

func NewPollingConnId(host, accid, userid string) string {
	return POLLINGCONNECTION_PREFIX + "_" + host + "_" + accid + "_" + userid + "_" + generateID("", 4)
}

func ExtractPollingConnId(connid string) (host, accid, userid string) {
	if !IsPollingConnID(connid) {
		return "", "", ""
	}
	sp := strings.Split(connid, "_")
	return sp[1], sp[2], sp[3]
}

func NewWsConnID(partition int32) string {
	return strconv.Itoa(int(partition)) + "p" + generateID(WS_PREFIX, 20)
}

func GetPartitionFromWsConnID(id string) int32 {
	for i := 0; i < len(id); i++ {
		if id[i] != 'p' {
			continue
		}
		par := id[:i]
		partition, err := strconv.Atoi(par)
		if err != nil {
			return -1
		}
		return int32(partition)
	}
	return -1
}

func NewUserID() string {
	return generateID(USER_PREFIX, 6)
}

func NewTagID() string {
	return generateID(TAG_PREFIX, 5)
}

func NewCannedResponseID() string {
	return generateID(CANNED_RESPONSE_PREFIX, 5)
}

func NewTemplateID() string {
	return generateID(TEMPLATE_PREFIX, 5)
}

func NewFileID() string {
	return generateID(FILE_PREFIX, 5)
}

func NewButtonID() string {
	return generateID(BUTTON_PREFIX, 8)
}

func NewChallengeID() string {
	return generateID(CHALLENGE_PREFIX, 40)
}

func NewSegmentationID() string {
	return generateID(SEGMENTATION_PREFIX, 5)
}

func NewSubizID() string {
	return generateID(SUBIZ_PREFIX, 8)
}

func NewUserFieldID() string {
	return generateID(USER_FIELD_PREFIX, 3)
}

func NewRequestID() string {
	return generateID(REQUEST_PREFIX, 10)
}

func NewAutomationID() string {
	return generateID(AUTOMATION_PREFIX, 5)
}

func NewUserSessionID() string {
	return generateID(USER_SESSION_PREFIX, 6)
}

func NewBillingID() string {
	return generateID(BILLING_PREFIX, 6)
}

func NewInvoiceID() string {
	return generateID(INVOICE_PREFIX, 6)
}

func NewSubscriptionID() string {
	return generateID(SUBSCRIPTION_PREFIX, 6)
}

func NewCustomerID() string {
	return generateID(CUSTOMER_PREFIX, 6)
}

func NewPaymentLogID() string {
	return generateID(PAYMENT_LOG_PREFIX, 6)
}

func NewPaymentMethodID() string {
	return generateID(PAYMENT_METHOD_PREFIX, 3)
}

func NewAttributeID() string {
	return generateID(ATTRIBUTE_PREFIX, 3)
}

func NewIdempotencyKey() string {
	return generateID(IDEMPOTENCY_KEY_PREFIX, 20)
}

func NewPaymentCommentID() string {
	return generateID(PAYMENT_COMMENT_PREFIX, 10)
}

func NewPipelineID() string {
	return generateID(PIPELINE_PREFIX, 5)
}

func NewStageID() string {
	return generateID(STAGE_PREFIX, 3)
}

func NewCurrencyID() string {
	return generateID(CURRENCY_PREFIX, 5)
}

func NewExchangeRateID() string {
	return generateID(EXCHANGE_RATE_PREFIX, 5)
}

func NewServiceLevelAgreementID() string {
	return generateID(SLA_PREFIX, 5)
}

func NewBackoffID() string {
	return generateID(BACKOFF_PREFIX, 10)
}

func IsUserID(id string) bool {
	if !strings.HasPrefix(id, USER_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, USER_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAccountID(id string) bool {
	if !strings.HasPrefix(id, ACCOUNT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, ACCOUNT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAgentGroupID(id string) bool {
	if !strings.HasPrefix(id, AGENT_GROUP_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, AGENT_GROUP_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAgentID(id string) bool {
	if !strings.HasPrefix(id, AGENT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, AGENT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsInvitationID(id string) bool {
	if !strings.HasPrefix(id, INVITATION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, INVITATION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsEventID(id string) bool {
	if !strings.HasPrefix(id, EVENT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, EVENT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsConversationID(id string) bool {
	if !strings.HasPrefix(id, CONVERSATION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CONVERSATION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsScheduleItemID(id string) bool {
	if !strings.HasPrefix(id, SCHEDULE_ITEM_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, SCHEDULE_ITEM_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsWebsendID(id string) bool {
	if !strings.HasPrefix(id, WEBSEND_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, WEBSEND_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAuthToken(id string) bool {
	if !strings.HasPrefix(id, AUTH_TOKEN_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, AUTH_TOKEN_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsRefreshToken(id string) bool {
	if !strings.HasPrefix(id, REFRESH_TOKEN_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, REFRESH_TOKEN_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsRuleID(id string) bool {
	if !strings.HasPrefix(id, RULE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, RULE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsClientID(id string) bool {
	if !strings.HasPrefix(id, CLIENT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CLIENT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsErrorID(id string) bool {
	if !strings.HasPrefix(id, ERROR_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, ERROR_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsWebhookID(id string) bool {
	if !strings.HasPrefix(id, WEBHOOK_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, WEBHOOK_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsWsConnID(id string) bool {
	par := GetPartitionFromWsConnID(id)
	if par < 0 {
		return false
	}
	id = strings.TrimPrefix(id, strconv.Itoa(int(par))+"p")
	if !strings.HasPrefix(id, WS_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, WS_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsTagID(id string) bool {
	if !strings.HasPrefix(id, TAG_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, TAG_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsCannedResponseID(id string) bool {
	if !strings.HasPrefix(id, CANNED_RESPONSE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CANNED_RESPONSE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsTemplateID(id string) bool {
	if !strings.HasPrefix(id, TEMPLATE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, TEMPLATE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsFileID(id string) bool {
	if !strings.HasPrefix(id, FILE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, FILE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsButtonID(id string) bool {
	if !strings.HasPrefix(id, BUTTON_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, BUTTON_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsChallengeID(id string) bool {
	if !strings.HasPrefix(id, CHALLENGE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CHALLENGE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsSegmentationID(id string) bool {
	if !strings.HasPrefix(id, SEGMENTATION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, SEGMENTATION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsSubizID(id string) bool {
	if !strings.HasPrefix(id, SUBIZ_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, SUBIZ_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsUserFieldID(id string) bool {
	if !strings.HasPrefix(id, USER_FIELD_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, USER_FIELD_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsRequestID(id string) bool {
	if !strings.HasPrefix(id, REQUEST_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, REQUEST_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAutomationID(id string) bool {
	if !strings.HasPrefix(id, AUTOMATION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, AUTOMATION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsUserSessionID(id string) bool {
	if !strings.HasPrefix(id, USER_SESSION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, USER_SESSION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsBillingID(id string) bool {
	if !strings.HasPrefix(id, BILLING_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, BILLING_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsInvoiceID(id string) bool {
	if !strings.HasPrefix(id, INVOICE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, INVOICE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsSubscriptionID(id string) bool {
	if !strings.HasPrefix(id, SUBSCRIPTION_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, SUBSCRIPTION_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsCustomerID(id string) bool {
	if !strings.HasPrefix(id, CUSTOMER_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CUSTOMER_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsPaymentLogID(id string) bool {
	if !strings.HasPrefix(id, PAYMENT_LOG_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, PAYMENT_LOG_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsPaymentMethodID(id string) bool {
	if !strings.HasPrefix(id, PAYMENT_METHOD_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, PAYMENT_METHOD_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsAttributeID(id string) bool {
	if !strings.HasPrefix(id, ATTRIBUTE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, ATTRIBUTE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsIdempotencyKey(id string) bool {
	if !strings.HasPrefix(id, IDEMPOTENCY_KEY_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, IDEMPOTENCY_KEY_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsPaymentCommentID(id string) bool {
	if !strings.HasPrefix(id, PAYMENT_COMMENT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, PAYMENT_COMMENT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func NewUserNoteID() string {
	return generateID(USER_NOTE_PREFIX, 10)
}

func IsUserNoteID(id string) bool {
	if !strings.HasPrefix(id, USER_NOTE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, USER_NOTE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func NewTicketID() string {
	return generateID(TICKET_PREFIX, 10)
}

func IsTicketID(id string) bool {
	if !strings.HasPrefix(id, TICKET_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, TICKET_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsPipelineID(id string) bool {
	if !strings.HasPrefix(id, PIPELINE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, PIPELINE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsStageID(id string) bool {
	if !strings.HasPrefix(id, STAGE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, STAGE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsPollingConnID(id string) bool {
	sp := strings.Split(id, "_")
	if len(sp) != 5 {
		return false
	}

	if sp[0] != POLLINGCONNECTION_PREFIX {
		return false
	}
	return true
}

func IsCurrencyID(id string) bool {
	if !strings.HasPrefix(id, CURRENCY_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, CURRENCY_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsExchangeRateID(id string) bool {
	if !strings.HasPrefix(id, EXCHANGE_RATE_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, EXCHANGE_RATE_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsServiceLevelAgreementID(id string) bool {
	if !strings.HasPrefix(id, SLA_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, SLA_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func IsBizbotID(id string) bool {
	if !strings.HasPrefix(id, BIZBOT_PREFIX) {
		return false
	}
	ts, err := GetCreated(id, BIZBOT_PREFIX)
	if err != nil || ts <= 0 {
		return false
	}
	return true
}

func NewAutomationActionID() string {
	return generateID(AUTOMATION_ACTION_PREFIX, 6)
}

func NewCampaignId() string {
	return generateID(CAMPAIGN_PREFIX, 6)
}

func NewConversionId() string {
	return generateID(CONVERSION_PREFIX, 6)
}

func NewGoalStatusId() string {
	return generateID(GOALSTATUS_PREFIX, 6)
}

func NewImpressionId() string {
	return generateID(IMPRESSION_PREFIX, 6)
}
