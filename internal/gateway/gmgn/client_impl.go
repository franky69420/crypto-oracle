package gmgn

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	http_client "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
)

// clientImpl est l'implémentation concrète du client GMGN
type clientImpl struct {
	config      ClientConfig
	tlsClient   tls_client.HttpClient
	jar         *cookiejar.Jar
	lastRequest time.Time
}

// newClientImpl crée une nouvelle instance de l'implémentation du client GMGN
func newClientImpl(config ClientConfig) *clientImpl {
	// Initialiser le cookie jar
	jar, _ := cookiejar.New(nil)

	// Configurer les options du client TLS
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(int(config.RequestTimeout)),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithCookieJar(jar),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	// Créer le client TLS
	tlsClient, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	return &clientImpl{
		config:      config,
		tlsClient:   tlsClient,
		jar:         jar,
		lastRequest: time.Now().Add(-config.RateLimitDelay * time.Millisecond),
	}
}

// getHeaders retourne les en-têtes HTTP à utiliser pour les requêtes
func (c *clientImpl) getHeaders(referer string) http_client.Header {
	headers := http_client.Header{
		"authority":           []string{"gmgn.ai"},
		"accept":              []string{"application/json, text/plain, */*"},
		"accept-language":     []string{"en-US,en;q=0.9,fr;q=0.8"},
		"sec-ch-ua":           []string{`"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`},
		"sec-ch-ua-mobile":    []string{"?0"},
		"sec-ch-ua-platform":  []string{`"Windows"`},
		"sec-fetch-dest":      []string{"empty"},
		"sec-fetch-mode":      []string{"cors"},
		"sec-fetch-site":      []string{"same-origin"},
		"user-agent":          []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		"x-requested-with":    []string{"XMLHttpRequest"},
	}

	if referer != "" {
		headers["referer"] = []string{referer}
		headers["origin"] = []string{"https://gmgn.ai"}
	}

	return headers
}

// buildQueryParams construit la chaîne de paramètres de requête pour les requêtes GMGN
func (c *clientImpl) buildQueryParams() string {
	return fmt.Sprintf(
		"device_id=%s&client_id=%s&from_app=%s&app_ver=%s&tz_name=%s&tz_offset=%s&app_lang=%s",
		c.config.DeviceID, c.config.ClientID, c.config.FromApp, c.config.AppVer, c.config.TzName, c.config.TzOffset, c.config.AppLang,
	)
}

// prepareSession effectue une requête préliminaire pour établir une session valide
func (c *clientImpl) prepareSession() error {
	req, err := http_client.NewRequest(http_client.MethodGet, "https://gmgn.ai", nil)
	if err != nil {
		return err
	}

	req.Header = c.getHeaders("")
	resp, err := c.tlsClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// makeRequest effectue une requête à l'API GMGN et vérifie les erreurs
func (c *clientImpl) makeRequest(url string) (Response, error) {
	// Respecter le taux de requêtes
	elapsed := time.Since(c.lastRequest)
	if elapsed < c.config.RateLimitDelay * time.Millisecond {
		time.Sleep(c.config.RateLimitDelay * time.Millisecond - elapsed)
	}
	c.lastRequest = time.Now()

	// Préparer la session si nécessaire
	if err := c.prepareSession(); err != nil {
		return Response{}, fmt.Errorf("échec de la préparation de la session: %w", err)
	}

	// Créer et exécuter la requête
	req, err := http_client.NewRequest(http_client.MethodGet, url, nil)
	if err != nil {
		return Response{}, fmt.Errorf("échec de la création de la requête: %w", err)
	}

	req.Header = c.getHeaders(url)
	resp, err := c.tlsClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("échec de la requête: %w", err)
	}
	defer resp.Body.Close()

	// Lire le corps de la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("échec de la lecture de la réponse: %w", err)
	}

	// Vérifier si la réponse est au format attendu
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return Response{
			Code: 1,
			Msg:  "Réponse invalide du serveur",
		}, nil
	}

	// Analyser la réponse JSON
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{
			Code: 40000300,
			Msg:  "argument invalide",
			Data: nil,
		}, nil
	}

	// Vérifier les erreurs dans la réponse
	if response.Code != 0 {
		return response, fmt.Errorf("erreur API: %d - %s", response.Code, response.Msg)
	}

	return response, nil
}

// GetTokenStat récupère les statistiques d'un token
func (c *clientImpl) GetTokenStat(tokenAddress string) (*TokenStatResponse, error) {
	url := fmt.Sprintf("%s/api/v1/token_stat/sol/%s?%s", 
		c.config.BaseURL, tokenAddress, c.buildQueryParams())
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var tokenStat TokenStatResponse
	if err := json.Unmarshal(data, &tokenStat); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &tokenStat, nil
}

// GetTokenTrades récupère l'historique des transactions d'un token
func (c *clientImpl) GetTokenTrades(tokenAddress string, limit int, tag string) (*TradeHistoryResponse, error) {
	url := fmt.Sprintf("%s/api/v1/token_trades/sol/%s?%s", 
		c.config.BaseURL, tokenAddress, c.buildQueryParams())
	
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if tag != "" {
		url += fmt.Sprintf("&tag=%s", tag)
	}
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var tradeHistory TradeHistoryResponse
	if err := json.Unmarshal(data, &tradeHistory); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &tradeHistory, nil
}

// GetTokenPrice récupère les données de prix d'un token
func (c *clientImpl) GetTokenPrice(tokenAddress string, timeframe string) (*KlineDataResponse, error) {
	if timeframe == "" {
		timeframe = "5m"
	}

	now := time.Now()
	endTime := now.Unix()
	startTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	url := fmt.Sprintf("%s/api/v1/token_kline/sol/%s?%s", 
		c.config.BaseURL, tokenAddress, c.buildQueryParams())
	url += fmt.Sprintf("&resolution=%s", timeframe)
	url += fmt.Sprintf("&from=%d&to=%d", startTime, endTime)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var klineData KlineDataResponse
	if err := json.Unmarshal(data, &klineData); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &klineData, nil
}

// GetAllTokenTraders récupère tous les traders d'un token avec pagination
func (c *clientImpl) GetAllTokenTraders(tokenAddress string) ([]Trader, error) {
	var allTraders []Trader
	nextToken := ""
	
	for {
		// Construire URL avec pagination
		url := fmt.Sprintf("%s/vas/api/v1/token_traders/sol/%s?%s", 
			c.config.BaseURL, tokenAddress, c.buildQueryParams())
		if nextToken != "" {
			url = fmt.Sprintf("%s&next=%s", url, nextToken)
		}
		
		resp, err := c.makeRequest(url)
		if err != nil {
			return nil, fmt.Errorf("échec de la récupération des traders: %w", err)
		}
		
		// Extraire les données des traders
		data, err := json.Marshal(resp.Data)
		if err != nil {
			return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
		}
		
		var tradersResponse struct {
			List []Trader `json:"list"`
			Next string   `json:"next"`
		}
		if err := json.Unmarshal(data, &tradersResponse); err != nil {
			return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
		}
		
		// Ajouter traders à la liste complète
		allTraders = append(allTraders, tradersResponse.List...)
		
		// Vérifier si plus de pages
		if tradersResponse.Next == "" {
			break
		}
		
		// Préparer pour la prochaine page
		nextToken = tradersResponse.Next
	}
	
	return allTraders, nil
}

// GetTokenHolderStat récupère les statistiques des détenteurs d'un token
func (c *clientImpl) GetTokenHolderStat(tokenAddress string) (*TokenHolderStatResponse, error) {
	url := fmt.Sprintf("%s/vas/api/v1/token_holder_stat/sol/%s?%s", 
		c.config.BaseURL, tokenAddress, c.buildQueryParams())
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var holderStat TokenHolderStatResponse
	if err := json.Unmarshal(data, &holderStat); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &holderStat, nil
}

// GetTokenWalletTagsStat récupère les statistiques des tags de wallets pour un token
func (c *clientImpl) GetTokenWalletTagsStat(tokenAddress string) (*TokenWalletTagsStatResponse, error) {
	url := fmt.Sprintf("%s/api/v1/token_wallet_tags_stat/sol/%s?%s", 
		c.config.BaseURL, tokenAddress, c.buildQueryParams())
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var tagsStat TokenWalletTagsStatResponse
	if err := json.Unmarshal(data, &tagsStat); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &tagsStat, nil
}

// GetWalletInfo récupère les informations d'un wallet
func (c *clientImpl) GetWalletInfo(walletAddress string) (*WalletInfoResponse, error) {
	url := fmt.Sprintf("%s/defi/quotation/v1/smartmoney/sol/walletNew/%s?%s", 
		c.config.BaseURL, walletAddress, c.buildQueryParams())
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var walletInfo WalletInfoResponse
	if err := json.Unmarshal(data, &walletInfo); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &walletInfo, nil
}

// GetAllWalletHoldings récupère tous les holdings d'un wallet avec pagination
func (c *clientImpl) GetAllWalletHoldings(walletAddress string) ([]Holding, error) {
	var allHoldings []Holding
	nextToken := ""
	
	for {
		// Construire URL avec pagination
		url := fmt.Sprintf("%s/api/v1/wallet_holdings/sol/%s?%s", 
			c.config.BaseURL, walletAddress, c.buildQueryParams())
		if nextToken != "" {
			url = fmt.Sprintf("%s&next=%s", url, nextToken)
		}
		
		resp, err := c.makeRequest(url)
		if err != nil {
			return nil, fmt.Errorf("échec de la récupération des holdings: %w", err)
		}
		
		// Extraire les données des holdings
		data, err := json.Marshal(resp.Data)
		if err != nil {
			return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
		}
		
		var holdingsResponse struct {
			Holdings []Holding `json:"holdings"`
			Next     string    `json:"next"`
		}
		if err := json.Unmarshal(data, &holdingsResponse); err != nil {
			return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
		}
		
		// Ajouter holdings à la liste complète
		allHoldings = append(allHoldings, holdingsResponse.Holdings...)
		
		// Vérifier si plus de pages
		if holdingsResponse.Next == "" {
			break
		}
		
		// Préparer pour la prochaine page
		nextToken = holdingsResponse.Next
	}
	
	return allHoldings, nil
}

// GetWalletStat récupère les statistiques d'un wallet
func (c *clientImpl) GetWalletStat(walletAddress string, period string) (*WalletStatResponse, error) {
	if period == "" {
		period = "all"
	}

	url := fmt.Sprintf("%s/api/v1/wallet_stat/sol/%s/%s?%s", 
		c.config.BaseURL, walletAddress, period, c.buildQueryParams())
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var walletStat WalletStatResponse
	if err := json.Unmarshal(data, &walletStat); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}

	return &walletStat, nil
}

// GetTrending récupère les tokens en tendance
func (c *clientImpl) GetTrending(timeframe string, orderBy string, direction string, filters []string) (*TrendingResponse, error) {
	if timeframe == "" {
		timeframe = "5m"
	}

	url := fmt.Sprintf("%s/defi/quotation/v1/rank/sol/swaps/%s?%s", 
		c.config.BaseURL, timeframe, c.buildQueryParams())

	if orderBy != "" {
		url += "&orderby=" + orderBy
	}
	if direction != "" {
		url += "&direction=" + direction
	}
	for _, filter := range filters {
		url += "&filters[]=" + filter
	}
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	var trendingResp TrendingResponse
	trendingResp.Code = resp.Code
	trendingResp.Msg = resp.Msg
	
	// Extraire les données des tokens en tendance
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var rankData struct {
		Rank []TrendingToken `json:"rank"`
	}
	if err := json.Unmarshal(data, &rankData); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}
	
	trendingResp.Data.Rank = rankData.Rank
	
	return &trendingResp, nil
}

// GetCompletedCoins récupère les tokens complétés
func (c *clientImpl) GetCompletedCoins(limit string, orderBy string, direction string) (*CompletedTokensResponse, error) {
	url := fmt.Sprintf("%s/defi/quotation/v1/rank/sol/completed?%s", 
		c.config.BaseURL, c.buildQueryParams())

	if limit != "" {
		url += "&limit=" + limit
	}
	if orderBy != "" {
		url += "&orderby=" + orderBy
	}
	if direction != "" {
		url += "&direction=" + direction
	}
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}

	// Convertir les données de réponse en structure attendue
	var completedResp CompletedTokensResponse
	completedResp.Code = resp.Code
	completedResp.Msg = resp.Msg
	
	// Extraire les données des tokens complétés
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("échec de la sérialisation des données: %w", err)
	}

	var rankData struct {
		Rank []CompletedToken `json:"rank"`
	}
	if err := json.Unmarshal(data, &rankData); err != nil {
		return nil, fmt.Errorf("échec de la désérialisation des données: %w", err)
	}
	
	completedResp.Data.Rank = rankData.Rank
	
	return &completedResp, nil
} 