package alerting

import (
	"context"
	"fmt"
	"time"

	"github.com/franko/crypto-oracle/pkg/models"
	"github.com/sirupsen/logrus"
)

// Manager gère les alertes pour les tokens et wallets
type Manager struct {
	logger *logrus.Logger
	alerts []models.TokenAlert
}

// NewManager crée un nouveau gestionnaire d'alertes
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		logger: logger,
		alerts: make([]models.TokenAlert, 0),
	}
}

// Start démarre le service d'alertes
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting Alert Manager")
	return nil
}

// Shutdown arrête le service d'alertes
func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("Shutting down Alert Manager")
	return nil
}

// CreateAlert crée une nouvelle alerte
func (m *Manager) CreateAlert(tokenAddress, tokenSymbol, alertType, severity, message string) (*models.TokenAlert, error) {
	alert := models.TokenAlert{
		ID:               fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		TokenAddress:     tokenAddress,
		TokenSymbol:      tokenSymbol,
		AlertType:        alertType,
		Severity:         severity,
		Message:          message,
		DetectedAt:       time.Now(),
		ConfirmationCount: 0,
		IsConfirmed:      false,
	}

	m.alerts = append(m.alerts, alert)
	m.logger.WithFields(logrus.Fields{
		"token_address": tokenAddress,
		"token_symbol":  tokenSymbol,
		"alert_type":    alertType,
		"severity":      severity,
	}).Info("Alert created")

	return &alert, nil
}

// GetAlerts récupère toutes les alertes
func (m *Manager) GetAlerts() []models.TokenAlert {
	return m.alerts
}

// ConfirmAlert confirme une alerte existante
func (m *Manager) ConfirmAlert(alertID string) error {
	for i, alert := range m.alerts {
		if alert.ID == alertID {
			m.alerts[i].ConfirmationCount++
			m.alerts[i].IsConfirmed = true
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// CreateTokenAlert crée une alerte pour un token basée sur des critères
func (m *Manager) CreateTokenAlert(token models.Token, xScore float64, walletAnalysis *models.WalletAnalysis) error {
	// Alerte pour token à score élevé
	if xScore > 80 {
		_, err := m.CreateAlert(
			token.Address,
			token.Symbol,
			"HIGH_SCORE",
			"URGENT",
			fmt.Sprintf("Token %s has a high X-Score of %.2f", token.Symbol, xScore),
		)
		return err
	}

	// Alerte pour forte présence de smart money
	if walletAnalysis != nil && walletAnalysis.TrustMetrics.SmartMoneyRatio > 0.3 {
		_, err := m.CreateAlert(
			token.Address,
			token.Symbol,
			"SMART_MONEY",
			"ALERT",
			fmt.Sprintf("Token %s has high smart money presence (%.1f%%)", 
				token.Symbol, walletAnalysis.TrustMetrics.SmartMoneyRatio*100),
		)
		return err
	}

	return nil
}

// CreateDumpAlert crée une alerte de dump potentiel
func (m *Manager) CreateDumpAlert(tokenAddress, tokenSymbol string, severity float64) error {
	severityText := "LOW"
	if severity > 70 {
		severityText = "CRITICAL"
	} else if severity > 50 {
		severityText = "HIGH"
	} else if severity > 30 {
		severityText = "MEDIUM"
	}

	_, err := m.CreateAlert(
		tokenAddress,
		tokenSymbol,
		"DUMP_DETECTED",
		severityText,
		fmt.Sprintf("Potential dump detected for %s (severity: %.1f)", tokenSymbol, severity),
	)
	return err
}

// CreateReactivationAlert crée une alerte de réactivation de token
func (m *Manager) CreateReactivationAlert(tokenAddress, tokenSymbol string, reactivationScore float64) error {
	_, err := m.CreateAlert(
		tokenAddress,
		tokenSymbol,
		"REACTIVATION",
		"ALERT",
		fmt.Sprintf("Token %s is reactivating with score %.1f", tokenSymbol, reactivationScore),
	)
	return err
} 