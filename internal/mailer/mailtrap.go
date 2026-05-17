package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"text/template"
)

type mailtrapClient struct {
	fromEmail string
	apiKey    string
	sandboxID string
	client    *http.Client
}

func NewMailTrapClient(apiKey, sandboxID, fromEmail string) (mailtrapClient, error) {
	if apiKey == "" {
		return mailtrapClient{}, errors.New("mailtrap api key is required")
	}

	if sandboxID == "" {
		return mailtrapClient{}, errors.New("mailtrap sandbox id is required")
	}

	if fromEmail == "" {
		return mailtrapClient{}, errors.New("from email is required")
	}

	return mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		sandboxID: sandboxID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (m mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	payload := struct {
		From struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"from"`
		To []struct {
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		} `json:"to"`
		Subject  string `json:"subject"`
		Text     string `json:"text,omitempty"`
		HTML     string `json:"html,omitempty"`
		Category string `json:"category,omitempty"`
	}{
		Subject:  subject.String(),
		HTML:     body.String(),
		Category: "User Invitation",
	}

	payload.From.Email = m.fromEmail
	payload.From.Name = FromName
	payload.To = append(payload.To, struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}{
		Email: email,
		Name:  username,
	})

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return -1, err
	}

	url := fmt.Sprintf("https://sandbox.api.mailtrap.io/api/send/%s", m.sandboxID)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var lastErr error
	for i := 0; i < maxRetires; i++ {
		resp, err := m.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp.StatusCode, nil
		}

		lastErr = fmt.Errorf("mailtrap api returned status %d", resp.StatusCode)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return -1, fmt.Errorf("failed to send email after %d attempts: %w", maxRetires, lastErr)
}
