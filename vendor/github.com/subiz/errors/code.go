package errors

//go:generate stringer -type=Code -trimprefix=E_
// Code represence error code
type Code int

const (
	E_no_error                  Code = iota
	E_unknown                   Code = iota
	E_internal_error            Code = iota
	E_missing_account_id        Code = iota
	E_proto_marshal_error       Code = iota
	E_database_error            Code = iota
	E_data_not_found            Code = iota
	E_access_deny               Code = iota
	E_json_marshal_error        Code = iota
	E_idempotency_input_changed Code = iota
	E_account_not_found         Code = iota
	E_invalid_credential        Code = iota
	E_http_call_error           Code = iota
	E_agent_group_not_found     Code = iota
	E_token_type_is_invalid     Code = iota
	E_jwt_token_is_invalid      Code = iota
	E_agent_not_found           Code = iota
	E_account_fullname_is_empty Code = iota
	E_password_is_empty         Code = iota
	E_password_too_weak         Code = iota
	E_email_is_empty            Code = iota
	E_email_is_invalid          Code = iota
	E_email_is_used             Code = iota
	E_content_type_is_not_json  Code = iota
	E_missing_url               Code = iota

	E_empty_alias Code = iota

	E_token_expired                         Code = iota
	E_agent_is_not_active                   Code = iota
	E_wrong_password                        Code = iota
	E_change_own_state                      Code = iota
	E_owner_state_cannot_be_changed         Code = iota
	E_pending_agent_state_cannot_be_changed Code = iota
	E_deleted_agent_state_cannot_be_changed Code = iota
	E_invalid_form_web_type                 Code = iota
	E_invalid_form_mobile_type              Code = iota
	E_invalid_form_email_type               Code = iota
	E_permission_not_enough                 Code = iota
	E_token_is_invalid                      Code = iota
	E_invalid_datetime                      Code = iota

	E_body_parsing_error       Code = iota
	E_query_parsing_error      Code = iota
	E_form_parsing_error       Code = iota
	E_invalid_http_method      Code = iota
	E_connection_id_is_invalid Code = iota

	E_agent_id_is_invalid Code = iota

	E_invalid_filestore_credential  Code = iota
	E_filestore_error               Code = iota
	E_filestore_write_error         Code = iota
	E_filestore_read_error          Code = iota
	E_filestore_acl_error           Code = iota
	E_redis_client_is_uninitialized Code = iota
	E_cannot_connect_to_redis       Code = iota
	E_redis_error                   Code = iota

	E_scrypt_error               Code = iota
	E_randomize_error            Code = iota
	E_scrypt_file_not_found      Code = iota
	E_invalid_mask               Code = iota
	E_form_is_invalid            Code = iota
	E_invalid_access_token       Code = iota
	E_invalid_apikey             Code = iota
	E_invalid_password           Code = iota
	E_whitelist_domain_not_found Code = iota
	E_blacklist_ip_not_found     Code = iota
	E_blocked_user_not_found     Code = iota
	E_user_id_is_invalid         Code = iota
	E_account_id_is_invalid      Code = iota
	E_ip_is_blocked              Code = iota
	E_answer_is_wrong            Code = iota
	E_url_is_not_whitelisted     Code = iota
	E_user_is_banned             Code = iota
	E_domain_is_not_whitelisted  Code = iota

	E_attribute_name_is_empty        Code = iota
	E_attribute_type_is_empty        Code = iota
	E_attribute_key_is_empty         Code = iota
	E_attribyte_type_is_unsupported  Code = iota
	E_attribute_list_is_empty        Code = iota
	E_attribute_key_is_invalid       Code = iota
	E_too_many_attribute             Code = iota
	E_attribute_key_not_found        Code = iota
	E_attribute_list_item_is_invalid Code = iota
	E_attribute_datetime_is_invalid  Code = iota
	E_attribute_boolean_is_invalid   Code = iota
	E_attribute_number_is_invalid    Code = iota
	E_attribute_text_is_invalid      Code = iota

	E_text_too_long Code = iota

	E_automation_not_found                Code = iota
	E_segmentation_not_found              Code = iota
	E_segmentation_id_is_invalid          Code = iota
	E_event_id_is_invalid                 Code = iota
	E_note_target_id_is_invalid           Code = iota
	E_note_id_is_invalid                  Code = iota
	E_unknown_segmentation_condition_type Code = iota

	E_kv_key_not_found                           Code = iota
	E_subscription_not_found                     Code = iota
	E_payment_method_not_found                   Code = iota
	E_missing_primary_payment_method             Code = iota
	E_payment_method_state_is_failed             Code = iota
	E_primary_payment_method_is_not_credit_card  Code = iota
	E_missing_payment_method_id                  Code = iota
	E_plan_not_found                             Code = iota
	E_invalid_plan_cannot_buy                    Code = iota
	E_missing_next_billing_cycle_month           Code = iota
	E_trial_package_cannot_buy                   Code = iota
	E_missing_billing_cycle_month                Code = iota
	E_primary_payment_method_must_be_credit_card Code = iota
	E_bank_transfer_payment_method_not_found     Code = iota
	E_invoice_not_found                          Code = iota
	E_invoice_template_error                     Code = iota
	E_execute_invoice_template_error             Code = iota
	E_missing_stripe_token                       Code = iota
	E_missing_stripe_info                        Code = iota
	E_stripe_call_error                          Code = iota
	E_missing_stripe_customer_id                 Code = iota
	E_invalid_invoice_id                         Code = iota
	E_invalid_payment_log_id                     Code = iota
	E_invalid_bill_id                            Code = iota
	E_unable_to_load_permission                  Code = iota
	E_auto_charge_is_not_enabled                 Code = iota
	E_stripe_customer_is_not_valid               Code = iota
	E_invoice_duedate_empty                      Code = iota

	E_kafka_rpc_timeout                  Code = iota
	E_kafka_error                        Code = iota
	E_kafka_rpc_handler_definition_error Code = iota

	E_s3_call_error                   Code = iota
	E_pdf_generate_error              Code = iota
	E_subscription_started_is_invalid Code = iota
	E_subscription_is_nil             Code = iota

	E_ticket_list_anchor_is_invalid  Code = iota
	E_elastic_search_error           Code = iota
	E_conversation_not_found         Code = iota
	E_chain_is_active                Code = iota
	E_invalid_conversation_no_user   Code = iota
	E_invalid_accepter_must_be_agent Code = iota
	E_conversation_ended             Code = iota

	E_user_is_the_last_one_in_conversation           Code = iota
	E_inviter_is_not_agent                           Code = iota
	E_event_must_be_message_event                    Code = iota
	E_conversation_id_missing                        Code = iota
	E_missing_conversation_id                        Code = iota
	E_ticket_id_missing                              Code = iota
	E_ticket_id_is_invalid                           Code = iota
	E_conversation_id_is_invalid                     Code = iota
	E_conversation_already_had_a_ticket              Code = iota
	E_ticket_not_found                               Code = iota
	E_webhook_not_found                              Code = iota
	E_webhook_is_disabled                            Code = iota
	E_invalid_conversation_state                     Code = iota
	E_event_is_invalid                               Code = iota
	E_limit_is_negative                              Code = iota
	E_invalid_integration                            Code = iota
	E_message_not_found                              Code = iota
	E_missing_pong                                   Code = iota
	E_invalid_pong_type                              Code = iota
	E_invalid_message_type                           Code = iota
	E_missing_message_id                             Code = iota
	E_user_is_not_in_the_conversation                Code = iota
	E_duplicated_message_received_error              Code = iota
	E_message_id_is_invalid                          Code = iota
	E_integration_id_is_invalid                      Code = iota
	E_integration_state_is_invalid                   Code = iota
	E_integration_not_found                          Code = iota
	E_undefined_router_condition_key                 Code = iota
	E_conversation_closed                            Code = iota
	E_user_has_already_in_conversation               Code = iota
	E_remover_is_not_agent                           Code = iota
	E_caller_is_not_leaver                           Code = iota
	E_leaver_is_the_last_one_in_conversation         Code = iota
	E_empty_message                                  Code = iota
	E_too_large_message                              Code = iota
	E_unknown_message_format                         Code = iota
	E_too_many_attachments                           Code = iota
	E_too_many_fields                                Code = iota
	E_too_large_attachment                           Code = iota
	E_too_long_field                                 Code = iota
	E_conversation_closer_is_not_in_the_conversation Code = iota
	E_invalid_channel_id                             Code = iota
	E_conversation_tag_not_found                     Code = iota

	E_client_not_found             Code = iota
	E_missing_redirect_url         Code = iota
	E_missing_client_name          Code = iota
	E_corrupted_user_data          Code = iota
	E_auth_token_expired           Code = iota
	E_hash_bcrypt_error            Code = iota
	E_invalid_cipher_size          Code = iota
	E_invalid_google_auth_response Code = iota
	E_challenge_not_found          Code = iota
	E_inactive_agent               Code = iota
	E_date_format_is_invalid       Code = iota
	E_pipeline_not_found           Code = iota
	E_pipeline_is_invalid          Code = iota
	E_stage_is_invalid             Code = iota
	E_currency_not_found           Code = iota
	E_currency_is_invalid          Code = iota
	E_exchange_rate_not_found      Code = iota
	E_exchange_rate_is_invalid     Code = iota
	E_working_day_is_invalid       Code = iota
	E_holiday_is_invalid           Code = iota

	E_service_level_agreement_not_found  Code = iota
	E_service_level_agreement_is_invalid Code = iota

	E_content_id_missing Code = iota

	E_fcm_token_not_found Code = iota

	E_grpc_agent_call_error        Code = iota
	E_grpc_account_call_error      Code = iota
	E_grpc_user_call_error         Code = iota
	E_grpc_conversation_call_error Code = iota
	E_grpc_content_call_error      Code = iota

	E_template_message_key_not_found  Code = iota
	E_template_message_not_is_creator Code = iota

	E_promotion_subcription_not_is_trial        Code = iota
	E_promotion_not_found                       Code = iota
	E_promotion_code_not_found                  Code = iota
	E_promotion_only_use_for_credit             Code = iota
	E_promotion_only_use_for_discount           Code = iota
	E_promotion_used_by_account                 Code = iota
	E_promotion_max_number_of_uses              Code = iota
	E_promotion_is_deleted                      Code = iota
	E_promotion_account_invalid                 Code = iota
	E_promotion_plan_invalid                    Code = iota
	E_promotion_number_of_agent_invalid         Code = iota
	E_promotion_referral_program_agent_inactive Code = iota
	E_promotion_add_log_bill_error              Code = iota
	E_promotion_add_log_account_invited_error   Code = iota
	E_promotion_referral_program_not_found      Code = iota
	E_referrer_code_invalid                     Code = iota
	E_transaction_exists                        Code = iota
	E_amount_invalid                            Code = iota
	E_transaction_is_missing                    Code = iota

	E_invalid_partition_version                  Code = iota
	E_invalid_partition_cluster                  Code = iota
	E_invalid_partition_term                     Code = iota
	E_partition_rebalance_timeout                Code = iota
	E_error_from_partition_peer                  Code = iota
	E_partition_node_have_not_joined_the_cluster Code = iota
	E_wrong_partition_host                       Code = iota
	E_duplicated_partition_term                  Code = iota
	E_worker_denied                              Code = iota

	E_file_key_is_missing     Code = iota
	E_file_size_is_missing    Code = iota
	E_file_size_invalid       Code = iota
	E_file_not_found          Code = iota
	E_file_key_secret_invalid Code = iota
	E_file_type_invalid       Code = iota

	E_file_error     Code = iota
	E_tempfile_error Code = iota
	E_invalid_css    Code = iota

	E_send_to_abandon_chan    Code = iota
	E_send_to_channel_timeout Code = iota

	E_too_many_subs_for_a_topic Code = iota
	E_too_many_topics_for_a_sub Code = iota

	E_endpoint_not_found Code = iota
	E_invalid_url        Code = iota

	E_facebook_call_error Code = iota

	E_zalo_call_error Code = iota

	E_invalid_oauth_scope    Code = iota
	E_setting_bot_exists     Code = iota
	E_setting_bot_not_found  Code = iota
	E_unconverted_v3_account Code = iota

	E_invalid_event_type Code = iota

	E_server_error Code = iota

	E_user_not_login Code = iota

	E_invalid_report_range  Code = iota
	E_invalid_report_metric Code = iota

	E_subiz_call_error Code = iota

	E_endchat_bot_setting_after_agent_message_too_low  Code = iota
	E_endchat_bot_setting_after_agent_message_too_high Code = iota
	E_endchat_bot_setting_after_user_message_too_low   Code = iota
	E_endchat_bot_setting_after_user_message_too_high  Code = iota
	E_endchat_bot_setting_after_any_message_too_low    Code = iota
	E_endchat_bot_setting_after_any_message_too_high   Code = iota
	E_endchat_bot_setting_age_too_low                  Code = iota
	E_endchat_bot_setting_age_too_high                 Code = iota

	E_webhook_verification_error Code = iota
	E_wrong_verification_key     Code = iota
	E_unreachable_webhook        Code = iota
	E_webhook_timeout            Code = iota

	E_goal_status_invalid   Code = iota
	E_not_found_campaign_id Code = iota

	E_invalid_poll_token               Code = iota
	E_expired_poll_token               Code = iota
	E_dead_poll_connection             Code = iota
	E_free_account_cannot_invite_agent Code = iota
)
