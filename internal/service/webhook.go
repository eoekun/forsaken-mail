package service

import "forsaken-mail/internal/webhook"

// WebhookTestRequest is the typed webhook test input model.
type WebhookTestRequest struct {
	Service string
	Config  string
	Message string
	Token   string
}

type webhookTester interface {
	SendTest(service, configStr, message, lang string) (*webhook.Result, error)
}

// WebhookService owns webhook test behavior and compatibility handling.
type WebhookService struct {
	tester webhookTester
}

// NewWebhookService creates a webhook test service.
func NewWebhookService(tester webhookTester) *WebhookService {
	return &WebhookService{tester: tester}
}

// SendTest sends a test notification using the supplied payload.
func (s *WebhookService) SendTest(req WebhookTestRequest, lang string) (*webhook.Result, error) {
	service := req.Service
	config := req.Config
	if service == "" && req.Token != "" {
		service = "dingtalk"
		config = `{"token":"` + req.Token + `"}`
	}

	return s.tester.SendTest(service, config, req.Message, lang)
}
