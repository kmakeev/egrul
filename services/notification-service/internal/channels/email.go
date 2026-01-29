package channels

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/egrul/notification-service/internal/model"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// EmailChannel реализация канала Email через SMTP
type EmailChannel struct {
	dialer        *gomail.Dialer
	from          string
	fromName      string
	htmlTemplate  *template.Template
	textTemplate  *template.Template
	logger        *zap.Logger
	maxRetries    int
	retryInterval time.Duration
	dryRun        bool // Режим логирования без отправки
}

// EmailConfig конфигурация Email канала
type EmailConfig struct {
	Host          string
	Port          int
	Username      string
	Password      string
	From          string
	FromName      string
	TLS           bool
	MaxRetries    int
	RetryInterval time.Duration
	DryRun        bool // Если true, не отправлять email, только логировать
}

// NewEmailChannel создает новый экземпляр Email канала
func NewEmailChannel(cfg EmailConfig, htmlTmpl, textTmpl *template.Template, logger *zap.Logger) *EmailChannel {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	if !cfg.TLS {
		dialer.TLSConfig = nil
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = 5 * time.Second
	}

	// Логируем режим работы
	if cfg.DryRun {
		logger.Warn("Email channel in DRY RUN mode - emails will NOT be sent, only logged")
	}

	return &EmailChannel{
		dialer:        dialer,
		from:          cfg.From,
		fromName:      cfg.FromName,
		htmlTemplate:  htmlTmpl,
		textTemplate:  textTmpl,
		logger:        logger,
		maxRetries:    cfg.MaxRetries,
		retryInterval: cfg.RetryInterval,
		dryRun:        cfg.DryRun,
	}
}

// Name возвращает название канала
func (c *EmailChannel) Name() string {
	return "email"
}

// Send отправляет уведомление по Email
func (c *EmailChannel) Send(ctx context.Context, notification *model.Notification) error {
	// Подготовка данных для шаблона
	data := c.prepareTemplateData(notification)

	// Рендеринг HTML шаблона
	htmlBody, err := c.renderTemplate(c.htmlTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Рендеринг текстового шаблона
	textBody, err := c.renderTemplate(c.textTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Создание сообщения
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", c.fromName, c.from))
	msg.SetHeader("To", notification.UserEmail)
	subject := c.getSubject(notification.ChangeEvent)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textBody)
	msg.AddAlternative("text/html", htmlBody)

	// DRY RUN режим - только логирование
	if c.dryRun {
		c.logger.Info("🔔 [DRY RUN] Email notification (NOT SENT)",
			zap.String("to", notification.UserEmail),
			zap.String("subject", subject),
			zap.String("entity_type", notification.ChangeEvent.EntityType),
			zap.String("entity_id", notification.ChangeEvent.EntityID),
			zap.String("entity_name", notification.ChangeEvent.EntityName),
			zap.String("change_type", notification.ChangeEvent.ChangeType),
			zap.String("field_name", notification.ChangeEvent.FieldName),
			zap.String("old_value", notification.ChangeEvent.OldValue),
			zap.String("new_value", notification.ChangeEvent.NewValue),
			zap.Bool("is_significant", notification.ChangeEvent.IsSignificant),
			zap.String("change_id", notification.ChangeEvent.ChangeID),
			zap.String("subscription_id", notification.SubscriptionID),
		)

		c.logger.Debug("[DRY RUN] Email text body",
			zap.String("text_body", textBody),
		)

		c.logger.Debug("[DRY RUN] Email HTML body length",
			zap.Int("html_length", len(htmlBody)),
		)

		return nil
	}

	// Реальная отправка с retry механизмом
	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		c.logger.Debug("attempting to send email",
			zap.String("to", notification.UserEmail),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", c.maxRetries),
		)

		if err := c.dialer.DialAndSend(msg); err != nil {
			lastErr = err
			c.logger.Warn("failed to send email",
				zap.String("to", notification.UserEmail),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)

			if attempt < c.maxRetries {
				// Exponential backoff
				waitTime := c.retryInterval * time.Duration(attempt)
				c.logger.Info("retrying after delay",
					zap.Duration("wait_time", waitTime),
				)
				time.Sleep(waitTime)
				continue
			}
		} else {
			// Успешно отправлено
			c.logger.Info("email sent successfully",
				zap.String("to", notification.UserEmail),
				zap.String("change_id", notification.ChangeEvent.ChangeID),
				zap.Int("attempt", attempt),
			)
			return nil
		}
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", c.maxRetries, lastErr)
}

// Close закрывает соединения
func (c *EmailChannel) Close() error {
	// gomail.Dialer не требует явного закрытия
	c.logger.Info("Email channel closed")
	return nil
}

// prepareTemplateData подготавливает данные для email шаблона
func (c *EmailChannel) prepareTemplateData(notification *model.Notification) model.EmailNotificationData {
	event := notification.ChangeEvent

	// Формируем URL для карточки сущности
	var entityURL string
	if event.EntityType == "company" {
		entityURL = fmt.Sprintf("http://localhost:3000/company/%s", event.EntityID)
	} else {
		entityURL = fmt.Sprintf("http://localhost:3000/entrepreneur/%s", event.EntityID)
	}

	return model.EmailNotificationData{
		EntityType:      event.EntityType,
		EntityID:        event.EntityID,
		EntityName:      event.EntityName,
		ChangeType:      event.ChangeType,
		FieldName:       event.FieldName,
		OldValue:        model.FormatValue(event.OldValue, event.FieldName),
		NewValue:        model.FormatValue(event.NewValue, event.FieldName),
		IsSignificant:   event.IsSignificant,
		DetectedAt:      event.DetectedAt,
		EntityURL:       entityURL,
		UnsubscribeURL:  fmt.Sprintf("http://localhost:3000/watchlist?action=unsubscribe&id=%s", notification.SubscriptionID),
		SettingsURL:     "http://localhost:3000/watchlist",
		ChangeTypeLabel: model.GetChangeTypeLabel(event.ChangeType),
		FieldNameLabel:  model.GetFieldNameLabel(event.FieldName),
	}
}

// renderTemplate рендерит шаблон с данными
func (c *EmailChannel) renderTemplate(tmpl *template.Template, data model.EmailNotificationData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// getSubject формирует тему письма
func (c *EmailChannel) getSubject(event *model.ChangeEvent) string {
	changeLabel := model.GetChangeTypeLabel(event.ChangeType)

	if event.IsSignificant {
		return fmt.Sprintf("⚠️ Важное изменение: %s - %s", event.EntityName, changeLabel)
	}

	return fmt.Sprintf("Изменение: %s - %s", event.EntityName, changeLabel)
}
