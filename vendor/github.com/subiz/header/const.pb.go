// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.11.2
// source: const.proto

package header

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type E int32

const (
	E_Sub                                       E = 0
	E_Email_SendRequest                         E = 10600
	E_LogLogRequested                           E = 11100
	E_LogRequested                              E = 11101
	E_LogSynced                                 E = 11102
	E_InvoiceUpdated                            E = 110410
	E_PaymentMethodUpdated                      E = 110411
	E_BillingUpdated                            E = 110412
	E_LogUpdated                                E = 110413
	E_SubscriptionUpdated                       E = 110414
	E_PromotionCodeUpdated                      E = 110415
	E_PaymentWorkerJob                          E = 110440
	E_AccountResetPasswordEmail                 E = 110533
	E_AccountPasswordChangedEmailRequested      E = 110534
	E_AccountConfirmChangeEmailEmailRequested   E = 110535
	E_AccountCreated                            E = 110545 // account
	E_AccountActivated                          E = 110546 // account
	E_AccountInfoUpdated                        E = 110547
	E_AccountPaymentV3Synced                    E = 110563
	E_AccountDeactivated                        E = 110564
	E_AgentLoggedIn                             E = 110565
	E_AccountWarnInactive                       E = 110566
	E_UnpaidInvoiceEmail                        E = 110567
	E_SubscriptionUpgradedEmail                 E = 110568
	E_BillCreatedEmail                          E = 110569
	E_DowngradedToFreeEmail                     E = 110570
	E_UnpaidInvoice9DaysEmail                   E = 110572
	E_AccountCreatedEmail                       E = 110573
	E_UpdatePlanEmail                           E = 110574
	E_TrialEnding7Email                         E = 110575
	E_TrialEnding3Email                         E = 110576
	E_TrialEnding1Email                         E = 110577
	E_InvoicePaidEmail                          E = 110578
	E_AccountAgentJoinedEmail                   E = 110579
	E_AgentPresenceUpdated                      E = 110580
	E_UserRequested                             E = 111100
	E_UserSynced                                E = 111101
	E_AutomationSynced                          E = 111102
	E_AutomationFired                           E = 111103
	E_AutomationBlockUserFired                  E = 111218
	E_AutomationUpdateUserAttributeFired        E = 111219
	E_AutomationMergeUsersFired                 E = 111220
	E_AutomationUserNotificationFired           E = 111221
	E_AutomationUserNoteFired                   E = 111223
	E_AutomationCreateConversationFired         E = 111224
	E_AutomationConversationMessage2Fired       E = 111225
	E_AutomationConversationWebhookFired        E = 111226
	E_AutomationAddAgentToConversationFired     E = 111227
	E_AutomationCreateTicketFired               E = 111228
	E_AutomationConversationTagFired            E = 111229
	E_UserTotalConvoUpdated                     E = 111232
	E_UserTotalOpenTicketUpdated                E = 111233
	E_TicketUpdated                             E = 111234
	E_AutomationSendEmailFired                  E = 111431
	E_AutomationOpenWidgetScreenFired           E = 111432
	E_AutomationSendChatTranscriptEmailFired    E = 111434
	E_AutomationSendChatTranscriptEmailResolved E = 111435
	E_AutomationUpdateConversationStateFired    E = 111436
	E_TicketSynced                              E = 2001001
	E_BqreportSynced                            E = 3000008
	E_BotSynced                                 E = 4000001
	E_StartBot                                  E = 4000002
	E_TeminateBot                               E = 4000003
)

// Enum value maps for E.
var (
	E_name = map[int32]string{
		0:       "Sub",
		10600:   "Email_SendRequest",
		11100:   "LogLogRequested",
		11101:   "LogRequested",
		11102:   "LogSynced",
		110410:  "InvoiceUpdated",
		110411:  "PaymentMethodUpdated",
		110412:  "BillingUpdated",
		110413:  "LogUpdated",
		110414:  "SubscriptionUpdated",
		110415:  "PromotionCodeUpdated",
		110440:  "PaymentWorkerJob",
		110533:  "AccountResetPasswordEmail",
		110534:  "AccountPasswordChangedEmailRequested",
		110535:  "AccountConfirmChangeEmailEmailRequested",
		110545:  "AccountCreated",
		110546:  "AccountActivated",
		110547:  "AccountInfoUpdated",
		110563:  "AccountPaymentV3Synced",
		110564:  "AccountDeactivated",
		110565:  "AgentLoggedIn",
		110566:  "AccountWarnInactive",
		110567:  "UnpaidInvoiceEmail",
		110568:  "SubscriptionUpgradedEmail",
		110569:  "BillCreatedEmail",
		110570:  "DowngradedToFreeEmail",
		110572:  "UnpaidInvoice9DaysEmail",
		110573:  "AccountCreatedEmail",
		110574:  "UpdatePlanEmail",
		110575:  "TrialEnding7Email",
		110576:  "TrialEnding3Email",
		110577:  "TrialEnding1Email",
		110578:  "InvoicePaidEmail",
		110579:  "AccountAgentJoinedEmail",
		110580:  "AgentPresenceUpdated",
		111100:  "UserRequested",
		111101:  "UserSynced",
		111102:  "AutomationSynced",
		111103:  "AutomationFired",
		111218:  "AutomationBlockUserFired",
		111219:  "AutomationUpdateUserAttributeFired",
		111220:  "AutomationMergeUsersFired",
		111221:  "AutomationUserNotificationFired",
		111223:  "AutomationUserNoteFired",
		111224:  "AutomationCreateConversationFired",
		111225:  "AutomationConversationMessage2Fired",
		111226:  "AutomationConversationWebhookFired",
		111227:  "AutomationAddAgentToConversationFired",
		111228:  "AutomationCreateTicketFired",
		111229:  "AutomationConversationTagFired",
		111232:  "UserTotalConvoUpdated",
		111233:  "UserTotalOpenTicketUpdated",
		111234:  "TicketUpdated",
		111431:  "AutomationSendEmailFired",
		111432:  "AutomationOpenWidgetScreenFired",
		111434:  "AutomationSendChatTranscriptEmailFired",
		111435:  "AutomationSendChatTranscriptEmailResolved",
		111436:  "AutomationUpdateConversationStateFired",
		2001001: "TicketSynced",
		3000008: "BqreportSynced",
		4000001: "BotSynced",
		4000002: "StartBot",
		4000003: "TeminateBot",
	}
	E_value = map[string]int32{
		"Sub":                                       0,
		"Email_SendRequest":                         10600,
		"LogLogRequested":                           11100,
		"LogRequested":                              11101,
		"LogSynced":                                 11102,
		"InvoiceUpdated":                            110410,
		"PaymentMethodUpdated":                      110411,
		"BillingUpdated":                            110412,
		"LogUpdated":                                110413,
		"SubscriptionUpdated":                       110414,
		"PromotionCodeUpdated":                      110415,
		"PaymentWorkerJob":                          110440,
		"AccountResetPasswordEmail":                 110533,
		"AccountPasswordChangedEmailRequested":      110534,
		"AccountConfirmChangeEmailEmailRequested":   110535,
		"AccountCreated":                            110545,
		"AccountActivated":                          110546,
		"AccountInfoUpdated":                        110547,
		"AccountPaymentV3Synced":                    110563,
		"AccountDeactivated":                        110564,
		"AgentLoggedIn":                             110565,
		"AccountWarnInactive":                       110566,
		"UnpaidInvoiceEmail":                        110567,
		"SubscriptionUpgradedEmail":                 110568,
		"BillCreatedEmail":                          110569,
		"DowngradedToFreeEmail":                     110570,
		"UnpaidInvoice9DaysEmail":                   110572,
		"AccountCreatedEmail":                       110573,
		"UpdatePlanEmail":                           110574,
		"TrialEnding7Email":                         110575,
		"TrialEnding3Email":                         110576,
		"TrialEnding1Email":                         110577,
		"InvoicePaidEmail":                          110578,
		"AccountAgentJoinedEmail":                   110579,
		"AgentPresenceUpdated":                      110580,
		"UserRequested":                             111100,
		"UserSynced":                                111101,
		"AutomationSynced":                          111102,
		"AutomationFired":                           111103,
		"AutomationBlockUserFired":                  111218,
		"AutomationUpdateUserAttributeFired":        111219,
		"AutomationMergeUsersFired":                 111220,
		"AutomationUserNotificationFired":           111221,
		"AutomationUserNoteFired":                   111223,
		"AutomationCreateConversationFired":         111224,
		"AutomationConversationMessage2Fired":       111225,
		"AutomationConversationWebhookFired":        111226,
		"AutomationAddAgentToConversationFired":     111227,
		"AutomationCreateTicketFired":               111228,
		"AutomationConversationTagFired":            111229,
		"UserTotalConvoUpdated":                     111232,
		"UserTotalOpenTicketUpdated":                111233,
		"TicketUpdated":                             111234,
		"AutomationSendEmailFired":                  111431,
		"AutomationOpenWidgetScreenFired":           111432,
		"AutomationSendChatTranscriptEmailFired":    111434,
		"AutomationSendChatTranscriptEmailResolved": 111435,
		"AutomationUpdateConversationStateFired":    111436,
		"TicketSynced":                              2001001,
		"BqreportSynced":                            3000008,
		"BotSynced":                                 4000001,
		"StartBot":                                  4000002,
		"TeminateBot":                               4000003,
	}
)

func (x E) Enum() *E {
	p := new(E)
	*p = x
	return p
}

func (x E) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (E) Descriptor() protoreflect.EnumDescriptor {
	return file_const_proto_enumTypes[0].Descriptor()
}

func (E) Type() protoreflect.EnumType {
	return &file_const_proto_enumTypes[0]
}

func (x E) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use E.Descriptor instead.
func (E) EnumDescriptor() ([]byte, []int) {
	return file_const_proto_rawDescGZIP(), []int{0}
}

var File_const_proto protoreflect.FileDescriptor

var file_const_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x2a, 0x96, 0x0e, 0x0a, 0x01, 0x45, 0x12, 0x07, 0x0a, 0x03, 0x53,
	0x75, 0x62, 0x10, 0x00, 0x12, 0x16, 0x0a, 0x11, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x53, 0x65,
	0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x10, 0xe8, 0x52, 0x12, 0x14, 0x0a, 0x0f,
	0x4c, 0x6f, 0x67, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x65, 0x64, 0x10,
	0xdc, 0x56, 0x12, 0x11, 0x0a, 0x0c, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x65, 0x64, 0x10, 0xdd, 0x56, 0x12, 0x0e, 0x0a, 0x09, 0x4c, 0x6f, 0x67, 0x53, 0x79, 0x6e, 0x63,
	0x65, 0x64, 0x10, 0xde, 0x56, 0x12, 0x14, 0x0a, 0x0e, 0x49, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xca, 0xde, 0x06, 0x12, 0x1a, 0x0a, 0x14, 0x50,
	0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x64, 0x10, 0xcb, 0xde, 0x06, 0x12, 0x14, 0x0a, 0x0e, 0x42, 0x69, 0x6c, 0x6c, 0x69,
	0x6e, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xcc, 0xde, 0x06, 0x12, 0x10, 0x0a,
	0x0a, 0x4c, 0x6f, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xcd, 0xde, 0x06, 0x12,
	0x19, 0x0a, 0x13, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xce, 0xde, 0x06, 0x12, 0x1a, 0x0a, 0x14, 0x50, 0x72,
	0x6f, 0x6d, 0x6f, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x64, 0x65, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x64, 0x10, 0xcf, 0xde, 0x06, 0x12, 0x16, 0x0a, 0x10, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e,
	0x74, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4a, 0x6f, 0x62, 0x10, 0xe8, 0xde, 0x06, 0x12, 0x1f,
	0x0a, 0x19, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x65, 0x74, 0x50, 0x61,
	0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xc5, 0xdf, 0x06, 0x12,
	0x2a, 0x0a, 0x24, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f,
	0x72, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x65, 0x64, 0x10, 0xc6, 0xdf, 0x06, 0x12, 0x2d, 0x0a, 0x27, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x65, 0x64, 0x10, 0xc7, 0xdf, 0x06, 0x12, 0x14, 0x0a, 0x0e, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x10, 0xd1, 0xdf, 0x06,
	0x12, 0x16, 0x0a, 0x10, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x63, 0x74, 0x69, 0x76,
	0x61, 0x74, 0x65, 0x64, 0x10, 0xd2, 0xdf, 0x06, 0x12, 0x18, 0x0a, 0x12, 0x41, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xd3,
	0xdf, 0x06, 0x12, 0x1c, 0x0a, 0x16, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x50, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x56, 0x33, 0x53, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x10, 0xe3, 0xdf, 0x06,
	0x12, 0x18, 0x0a, 0x12, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65, 0x61, 0x63, 0x74,
	0x69, 0x76, 0x61, 0x74, 0x65, 0x64, 0x10, 0xe4, 0xdf, 0x06, 0x12, 0x13, 0x0a, 0x0d, 0x41, 0x67,
	0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x67, 0x67, 0x65, 0x64, 0x49, 0x6e, 0x10, 0xe5, 0xdf, 0x06, 0x12,
	0x19, 0x0a, 0x13, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x57, 0x61, 0x72, 0x6e, 0x49, 0x6e,
	0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x10, 0xe6, 0xdf, 0x06, 0x12, 0x18, 0x0a, 0x12, 0x55, 0x6e,
	0x70, 0x61, 0x69, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x45, 0x6d, 0x61, 0x69, 0x6c,
	0x10, 0xe7, 0xdf, 0x06, 0x12, 0x1f, 0x0a, 0x19, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x55, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x64, 0x45, 0x6d, 0x61, 0x69,
	0x6c, 0x10, 0xe8, 0xdf, 0x06, 0x12, 0x16, 0x0a, 0x10, 0x42, 0x69, 0x6c, 0x6c, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xe9, 0xdf, 0x06, 0x12, 0x1b, 0x0a,
	0x15, 0x44, 0x6f, 0x77, 0x6e, 0x67, 0x72, 0x61, 0x64, 0x65, 0x64, 0x54, 0x6f, 0x46, 0x72, 0x65,
	0x65, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xea, 0xdf, 0x06, 0x12, 0x1d, 0x0a, 0x17, 0x55, 0x6e,
	0x70, 0x61, 0x69, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x39, 0x44, 0x61, 0x79, 0x73,
	0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xec, 0xdf, 0x06, 0x12, 0x19, 0x0a, 0x13, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c,
	0x10, 0xed, 0xdf, 0x06, 0x12, 0x15, 0x0a, 0x0f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x6c,
	0x61, 0x6e, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xee, 0xdf, 0x06, 0x12, 0x17, 0x0a, 0x11, 0x54,
	0x72, 0x69, 0x61, 0x6c, 0x45, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x37, 0x45, 0x6d, 0x61, 0x69, 0x6c,
	0x10, 0xef, 0xdf, 0x06, 0x12, 0x17, 0x0a, 0x11, 0x54, 0x72, 0x69, 0x61, 0x6c, 0x45, 0x6e, 0x64,
	0x69, 0x6e, 0x67, 0x33, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xf0, 0xdf, 0x06, 0x12, 0x17, 0x0a,
	0x11, 0x54, 0x72, 0x69, 0x61, 0x6c, 0x45, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x31, 0x45, 0x6d, 0x61,
	0x69, 0x6c, 0x10, 0xf1, 0xdf, 0x06, 0x12, 0x16, 0x0a, 0x10, 0x49, 0x6e, 0x76, 0x6f, 0x69, 0x63,
	0x65, 0x50, 0x61, 0x69, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xf2, 0xdf, 0x06, 0x12, 0x1d,
	0x0a, 0x17, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x4a, 0x6f,
	0x69, 0x6e, 0x65, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0xf3, 0xdf, 0x06, 0x12, 0x1a, 0x0a,
	0x14, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63, 0x65, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0xf4, 0xdf, 0x06, 0x12, 0x13, 0x0a, 0x0d, 0x55, 0x73, 0x65,
	0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x65, 0x64, 0x10, 0xfc, 0xe3, 0x06, 0x12, 0x10,
	0x0a, 0x0a, 0x55, 0x73, 0x65, 0x72, 0x53, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x10, 0xfd, 0xe3, 0x06,
	0x12, 0x16, 0x0a, 0x10, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x79,
	0x6e, 0x63, 0x65, 0x64, 0x10, 0xfe, 0xe3, 0x06, 0x12, 0x15, 0x0a, 0x0f, 0x41, 0x75, 0x74, 0x6f,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xff, 0xe3, 0x06, 0x12,
	0x1e, 0x0a, 0x18, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x55, 0x73, 0x65, 0x72, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf2, 0xe4, 0x06, 0x12,
	0x28, 0x0a, 0x22, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf3, 0xe4, 0x06, 0x12, 0x1f, 0x0a, 0x19, 0x41, 0x75, 0x74,
	0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x55, 0x73, 0x65, 0x72,
	0x73, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf4, 0xe4, 0x06, 0x12, 0x25, 0x0a, 0x1f, 0x41, 0x75,
	0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x55, 0x73, 0x65, 0x72, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf5, 0xe4,
	0x06, 0x12, 0x1d, 0x0a, 0x17, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x55,
	0x73, 0x65, 0x72, 0x4e, 0x6f, 0x74, 0x65, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf7, 0xe4, 0x06,
	0x12, 0x27, 0x0a, 0x21, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xf8, 0xe4, 0x06, 0x12, 0x29, 0x0a, 0x23, 0x41, 0x75, 0x74,
	0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x46, 0x69, 0x72, 0x65, 0x64,
	0x10, 0xf9, 0xe4, 0x06, 0x12, 0x28, 0x0a, 0x22, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x57, 0x65,
	0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xfa, 0xe4, 0x06, 0x12, 0x2b,
	0x0a, 0x25, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x64, 0x64, 0x41,
	0x67, 0x65, 0x6e, 0x74, 0x54, 0x6f, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xfb, 0xe4, 0x06, 0x12, 0x21, 0x0a, 0x1b, 0x41,
	0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54,
	0x69, 0x63, 0x6b, 0x65, 0x74, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xfc, 0xe4, 0x06, 0x12, 0x24,
	0x0a, 0x1e, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x76,
	0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x61, 0x67, 0x46, 0x69, 0x72, 0x65, 0x64,
	0x10, 0xfd, 0xe4, 0x06, 0x12, 0x1b, 0x0a, 0x15, 0x55, 0x73, 0x65, 0x72, 0x54, 0x6f, 0x74, 0x61,
	0x6c, 0x43, 0x6f, 0x6e, 0x76, 0x6f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0x80, 0xe5,
	0x06, 0x12, 0x20, 0x0a, 0x1a, 0x55, 0x73, 0x65, 0x72, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x4f, 0x70,
	0x65, 0x6e, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10,
	0x81, 0xe5, 0x06, 0x12, 0x13, 0x0a, 0x0d, 0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x64, 0x10, 0x82, 0xe5, 0x06, 0x12, 0x1e, 0x0a, 0x18, 0x41, 0x75, 0x74, 0x6f,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x6e, 0x64, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x46,
	0x69, 0x72, 0x65, 0x64, 0x10, 0xc7, 0xe6, 0x06, 0x12, 0x25, 0x0a, 0x1f, 0x41, 0x75, 0x74, 0x6f,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x65, 0x6e, 0x57, 0x69, 0x64, 0x67, 0x65, 0x74,
	0x53, 0x63, 0x72, 0x65, 0x65, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xc8, 0xe6, 0x06, 0x12,
	0x2c, 0x0a, 0x26, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x6e,
	0x64, 0x43, 0x68, 0x61, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x45,
	0x6d, 0x61, 0x69, 0x6c, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xca, 0xe6, 0x06, 0x12, 0x2f, 0x0a,
	0x29, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x6e, 0x64, 0x43,
	0x68, 0x61, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x45, 0x6d, 0x61,
	0x69, 0x6c, 0x52, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x64, 0x10, 0xcb, 0xe6, 0x06, 0x12, 0x2c,
	0x0a, 0x26, 0x41, 0x75, 0x74, 0x6f, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x46, 0x69, 0x72, 0x65, 0x64, 0x10, 0xcc, 0xe6, 0x06, 0x12, 0x12, 0x0a, 0x0c,
	0x54, 0x69, 0x63, 0x6b, 0x65, 0x74, 0x53, 0x79, 0x6e, 0x63, 0x65, 0x64, 0x10, 0xe9, 0x90, 0x7a,
	0x12, 0x15, 0x0a, 0x0e, 0x42, 0x71, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x79, 0x6e, 0x63,
	0x65, 0x64, 0x10, 0xc8, 0x8d, 0xb7, 0x01, 0x12, 0x10, 0x0a, 0x09, 0x42, 0x6f, 0x74, 0x53, 0x79,
	0x6e, 0x63, 0x65, 0x64, 0x10, 0x81, 0x92, 0xf4, 0x01, 0x12, 0x0f, 0x0a, 0x08, 0x53, 0x74, 0x61,
	0x72, 0x74, 0x42, 0x6f, 0x74, 0x10, 0x82, 0x92, 0xf4, 0x01, 0x12, 0x12, 0x0a, 0x0b, 0x54, 0x65,
	0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x42, 0x6f, 0x74, 0x10, 0x83, 0x92, 0xf4, 0x01, 0x42, 0x19,
	0x5a, 0x17, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75, 0x62,
	0x69, 0x7a, 0x2f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_const_proto_rawDescOnce sync.Once
	file_const_proto_rawDescData = file_const_proto_rawDesc
)

func file_const_proto_rawDescGZIP() []byte {
	file_const_proto_rawDescOnce.Do(func() {
		file_const_proto_rawDescData = protoimpl.X.CompressGZIP(file_const_proto_rawDescData)
	})
	return file_const_proto_rawDescData
}

var file_const_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_const_proto_goTypes = []interface{}{
	(E)(0), // 0: header.E
}
var file_const_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_const_proto_init() }
func file_const_proto_init() {
	if File_const_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_const_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_const_proto_goTypes,
		DependencyIndexes: file_const_proto_depIdxs,
		EnumInfos:         file_const_proto_enumTypes,
	}.Build()
	File_const_proto = out.File
	file_const_proto_rawDesc = nil
	file_const_proto_goTypes = nil
	file_const_proto_depIdxs = nil
}