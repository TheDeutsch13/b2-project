package http

import (
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CdekHandler struct {
	clientID     string
	clientSecret string
	logger       *zap.Logger
}

func NewCdekHandler(clientID, clientSecret string, logger *zap.Logger) *CdekHandler {
	return &CdekHandler{
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
	}
}

type cdekPointResponse struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	WorkTime  string  `json:"work_time"`
	Phone     string  `json:"phone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ListPoints godoc
// @Summary List CDEK pickup points by city
// @Tags cdek
// @Produce json
// @Param city query string true "City name"
// @Success 200 {array} cdekPointResponse
// @Router /api/cdek/points [get]
func (h *CdekHandler) ListPoints(ctx *gin.Context) {
	city := strings.TrimSpace(ctx.Query("city"))
	if city == "" {
		ctx.JSON(stdhttp.StatusBadRequest, gin.H{"error": "city is required"})
		return
	}

	if h.clientID != "" && h.clientSecret != "" {
		points, err := h.fetchFromCdekAPI(ctx, city)
		if err == nil && len(points) > 0 {
			ctx.JSON(stdhttp.StatusOK, points)
			return
		}

		if err != nil {
			h.logger.Warn("cdek api failed, using demo points", zap.Error(err))
		}
	}

	ctx.JSON(stdhttp.StatusOK, demoCdekPoints(city))
}

func demoCdekPoints(city string) []cdekPointResponse {
	normalized := strings.ToLower(city)

	if strings.Contains(normalized, "саратов") {
		return []cdekPointResponse{
			{
				Code: "SAR115", Name: "SAR115, Саратов, ул. Блинова",
				Address: "410007, Россия, Саратовская область, Саратов, ул. Блинова, 23",
				City: "Саратов", WorkTime: "Пн-Пт 10:00-20:00, Сб-Вс 10:00-18:00",
				Phone: "+7 (8452) 00-00-00", Latitude: 51.533, Longitude: 46.034,
			},
			{
				Code: "SAR42", Name: "SAR42, Саратов, ул. Московская",
				Address: "410012, Россия, Саратовская область, Саратов, ул. Московская, 57",
				City: "Саратов", WorkTime: "Пн-Вс 09:00-21:00",
				Phone: "+7 (8452) 11-11-11", Latitude: 51.526, Longitude: 46.017,
			},
			{
				Code: "SAR88", Name: "SAR88, Саратов, пр-т Строителей",
				Address: "410065, Россия, Саратовская область, Саратов, пр-т Строителей, 1",
				City: "Саратов", WorkTime: "Пн-Пт 10:00-19:00",
				Phone: "+7 (8452) 22-22-22", Latitude: 51.589, Longitude: 45.954,
			},
		}
	}

	if strings.Contains(normalized, "москв") {
		return []cdekPointResponse{
			{
				Code: "MSK101", Name: "MSK101, Москва, ул. Тверская",
				Address: "125009, Россия, Москва, ул. Тверская, 12",
				City: "Москва", WorkTime: "Пн-Вс 10:00-22:00",
				Phone: "+7 (495) 00-00-00", Latitude: 55.757, Longitude: 37.614,
			},
			{
				Code: "MSK205", Name: "MSK205, Москва, ул. Профсоюзная",
				Address: "117437, Россия, Москва, ул. Профсоюзная, 56",
				City: "Москва", WorkTime: "Пн-Пт 09:00-20:00",
				Phone: "+7 (495) 11-11-11", Latitude: 55.677, Longitude: 37.562,
			},
		}
	}

	return []cdekPointResponse{
		{
			Code: "DEMO1", Name: fmt.Sprintf("ПВЗ СДЭК, %s", city),
			Address: fmt.Sprintf("Россия, %s, ул. Центральная, 1", city),
			City: city, WorkTime: "Пн-Пт 10:00-20:00",
			Phone: "+7 (800) 000-00-00", Latitude: 55.75, Longitude: 37.62,
		},
	}
}

type cdekTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type cdekAPIPoint struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Location struct {
		City        string  `json:"city"`
		Address     string  `json:"address"`
		AddressFull string  `json:"address_full"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	} `json:"location"`
	WorkTime string `json:"work_time"`
	Phones   []struct {
		Number string `json:"number"`
	} `json:"phones"`
}

func (h *CdekHandler) fetchFromCdekAPI(ctx *gin.Context, city string) ([]cdekPointResponse, error) {
	token, err := h.getCdekToken(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://api.cdek.ru/v2/deliverypoints?type=PVZ&city=%s&size=50",
		url.QueryEscape(city),
	)

	req, err := stdhttp.NewRequestWithContext(
		ctx.Request.Context(),
		stdhttp.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != stdhttp.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cdek points status %d: %s", resp.StatusCode, string(body))
	}

	var raw []cdekAPIPoint
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	points := make([]cdekPointResponse, 0, len(raw))
	for _, item := range raw {
		phone := ""
		if len(item.Phones) > 0 {
			phone = item.Phones[0].Number
		}

		address := item.Location.AddressFull
		if address == "" {
			address = item.Location.Address
		}

		points = append(points, cdekPointResponse{
			Code:      item.Code,
			Name:      item.Name,
			Address:   address,
			City:      item.Location.City,
			WorkTime:  item.WorkTime,
			Phone:     phone,
			Latitude:  item.Location.Latitude,
			Longitude: item.Location.Longitude,
		})
	}

	return points, nil
}

func (h *CdekHandler) getCdekToken(ctx *gin.Context) (string, error) {
	req, err := stdhttp.NewRequestWithContext(
		ctx.Request.Context(),
		stdhttp.MethodPost,
		"https://api.cdek.ru/v2/oauth/token",
		strings.NewReader("grant_type=client_credentials&client_id="+h.clientID+"&client_secret="+h.clientSecret),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &stdhttp.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != stdhttp.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cdek auth status %d: %s", resp.StatusCode, string(body))
	}

	var payload cdekTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	return payload.AccessToken, nil
}
