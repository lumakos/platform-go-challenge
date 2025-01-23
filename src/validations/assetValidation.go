package validations

import (
	"encoding/json"
	"fmt"
	"platform-go-challenge/src/models"
	"strings"
)

// Validates chart data
func ValidateChartData(data *models.ChartData) error {
	if data.Title == "" {
		return fmt.Errorf("Title cannot be empty")
	}
	if data.XAxis == "" || data.YAxis == "" {
		return fmt.Errorf("XAxis and YAxis cannot be empty")
	}
	if len(data.Data) == 0 {
		return fmt.Errorf("Chart data cannot be empty")
	}
	return nil
}

// Validates insight data
func ValidateInsightData(data *models.InsightData) error {
	data.Text = strings.TrimSpace(data.Text)
	if data.Text == "" {
		return fmt.Errorf("Insight text cannot be empty")
	}
	if len(data.Text) > 500 {
		return fmt.Errorf("Insight text cannot exceed 500 characters")
	}
	return nil
}

// Validates audience data
func ValidateAudienceData(data *models.AudienceData) error {
	data.Gender = strings.TrimSpace(data.Gender)
	if data.Gender != "male" && data.Gender != "female" {
		return fmt.Errorf("Gender must be 'male' or 'female'")
	}
	if data.AgeRange == "" {
		return fmt.Errorf("Age range cannot be empty")
	}
	if data.PurchasesLastMo <= 0 {
		return fmt.Errorf("Number of purchases must be greater than 0")
	}
	return nil
}

// Validates Asset
func ValidateAsset(asset *models.Asset) error {
	switch asset.Type {
	case models.Chart:
		var chartData models.ChartData
		if err := json.Unmarshal(asset.Data, &chartData); err != nil {
			return fmt.Errorf("Invalid chart data format")
		}
		if err := ValidateChartData(&chartData); err != nil {
			return err
		}
		asset.Data, _ = json.Marshal(chartData)

	case models.Insight:
		var insightData models.InsightData
		if err := json.Unmarshal(asset.Data, &insightData); err != nil {
			return fmt.Errorf("Invalid insight data format")
		}
		if err := ValidateInsightData(&insightData); err != nil {
			return err
		}
		asset.Data, _ = json.Marshal(insightData)

	case models.Audience:
		var audienceData models.AudienceData
		if err := json.Unmarshal(asset.Data, &audienceData); err != nil {
			return fmt.Errorf("Invalid audience data format")
		}
		if err := ValidateAudienceData(&audienceData); err != nil {
			return err
		}
		asset.Data, _ = json.Marshal(audienceData)

	default:
		return fmt.Errorf("Unsupported asset type")
	}
	return nil
}
