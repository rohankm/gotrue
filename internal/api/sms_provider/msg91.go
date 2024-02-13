package sms_provider

import (

	"encoding/json"
	"fmt"
	"net/http"


	"strings"  // Add this import for the "strings" package
	"io" 
	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/utilities"
)

const (
	defaultMsg91ApiBase = "https://control.msg91.com/api/v5/flow"
)

type Msg91Provider struct {
	Config  *conf.Msg91ProviderConfiguration
	APIPath string
}

type Msg91Response struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewMsg91Provider creates a new SmsProvider for Msg91.
func NewMsg91Provider(config conf.Msg91ProviderConfiguration) (SmsProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Msg91Provider{
		Config:  &config,
		APIPath: defaultMsg91ApiBase,
	}, nil
}

// SendMessage implements the SmsProvider interface for Msg91Provider.
func (t *Msg91Provider) SendMessage(phone, message, channel, otp string) (string, error) {
	switch channel {
	case SMSProvider:
		return t.SendSms(phone, message,otp)
	default:
		return "", fmt.Errorf("msg91: channel type %q is not supported", channel)
	}
}

func (t *Msg91Provider) SendSms(phone, message, otp string) (string, error) {
  

	payload := strings.NewReader(fmt.Sprintf("{\"template_id\":\"%s\",\"recipients\":[{\"mobiles\":\"%s\",\"otp\":\"%s\"}]}", t.Config.TemplateId, phone, otp))



	client := &http.Client{Timeout: defaultTimeout}

    req, err := http.NewRequest("POST", t.APIPath, payload)
    if err != nil {
        return "", fmt.Errorf("msg91 error: unable to create request %w", err)
    }


	req.Header.Add("accept", "application/json")
    req.Header.Add("content-type", "application/json")
    req.Header.Add("authkey", t.Config.AuthKey)

    res, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("msg91 error: failed to execute request %w", err)
    }
    defer utilities.SafeClose(res.Body)

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return "", fmt.Errorf("msg91 error: failed to read response body: %w", err)
    }

    fmt.Println(string(body)) // Assuming you want to print the response body

    var resp Msg91Response
    if err := json.Unmarshal(body, &resp); err != nil {
        return "", fmt.Errorf("msg91 error: failed to unmarshal JSON response body (status code %v): %w", res.StatusCode, err)
    }

    if resp.Type != "success" {
        return resp.Message, fmt.Errorf("msg91 error: expected \"success\" but got %q with message %q (code: %v)", resp.Type, resp.Message, res.StatusCode)
    }

    return resp.Message, nil
}

