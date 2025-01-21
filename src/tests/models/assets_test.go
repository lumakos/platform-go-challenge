package models

import (
	"encoding/json"
	"testing"

	"platform-go-challenge/src/models"

	"github.com/stretchr/testify/assert"
)

func TestChartData_GetDescription(t *testing.T) {
	data := models.ChartData{
		Title: "Sales Growth",
		XAxis: "Months",
		YAxis: "Revenue",
		Data:  []float64{100, 200, 300},
	}

	description := data.GetDescription()
	assert.Equal(t, "Sales Growth Months Revenue", description)
}

func TestInsightData_GetDescription(t *testing.T) {
	data := models.InsightData{
		Text: "This quarter saw a 15% increase in sales.",
	}

	description := data.GetDescription()
	assert.Equal(t, "This quarter saw a 15% increase in sales.", description)
}

func TestAudienceData_GetDescription(t *testing.T) {
	data := models.AudienceData{
		Gender:           "Female",
		BirthCountry:     "USA",
		AgeRange:         "25-34",
		SocialMediaUsage: 3.5,
		PurchasesLastMo:  5,
	}

	description := data.GetDescription()
	expected := "Female USA 25-34 3.5000005"
	assert.Equal(t, expected, description)
}

func TestRetrieveDescription(t *testing.T) {
	chart := models.ChartData{
		Title: "Sales Growth",
		XAxis: "Months",
		YAxis: "Revenue",
		Data:  []float64{100, 200, 300},
	}

	insight := models.InsightData{
		Text: "This quarter saw a 15% increase in sales.",
	}

	audience := models.AudienceData{
		Gender:           "Female",
		BirthCountry:     "USA",
		AgeRange:         "25-34",
		SocialMediaUsage: 3.5,
		PurchasesLastMo:  5,
	}

	assert.Equal(t, "Sales Growth Months Revenue", models.RetrieveDescription(chart))
	assert.Equal(t, "This quarter saw a 15% increase in sales.", models.RetrieveDescription(insight))
	assert.Equal(t, "Female USA 25-34 3.5000005", models.RetrieveDescription(audience))
}

func TestAsset_JSONMarshalling(t *testing.T) {
	asset := models.Asset{
		ID:   1,
		Type: models.Chart,
		Description: models.ChartData{
			Title: "Sales Growth",
			XAxis: "Months",
			YAxis: "Revenue",
			Data:  []float64{100, 200, 300},
		},
		Data: json.RawMessage(`{"title":"Sales Growth","x_axis":"Months","y_axis":"Revenue","data":[100,200,300]}`),
	}

	// Marshalling
	marshalled, err := json.Marshal(asset)
	assert.NoError(t, err)

	// Unmarshalling
	var unmarshalled models.Asset
	err = json.Unmarshal(marshalled, &unmarshalled)
	assert.NoError(t, err)

	// Validate unmarshalled data
	assert.Equal(t, asset.ID, unmarshalled.ID)
	assert.Equal(t, asset.Type, unmarshalled.Type)

	// RawMessage comparison (ensure it's equal as JSON)
	assert.JSONEq(t, string(asset.Data), string(unmarshalled.Data))
}

func TestAsset_DescriptionPolymorphism(t *testing.T) {
	chart := models.Asset{
		ID:   1,
		Type: models.Chart,
		Description: models.ChartData{
			Title: "Sales Growth",
			XAxis: "Months",
			YAxis: "Revenue",
			Data:  []float64{100, 200, 300},
		},
	}

	insight := models.Asset{
		ID:   2,
		Type: models.Insight,
		Description: models.InsightData{
			Text: "This quarter saw a 15% increase in sales.",
		},
	}

	audience := models.Asset{
		ID:   3,
		Type: models.Audience,
		Description: models.AudienceData{
			Gender:           "Female",
			BirthCountry:     "USA",
			AgeRange:         "25-34",
			SocialMediaUsage: 3.5,
			PurchasesLastMo:  5,
		},
	}

	assert.Equal(t, "Sales Growth Months Revenue", models.RetrieveDescription(chart.Description.(models.ChartData)))
	assert.Equal(t, "This quarter saw a 15% increase in sales.", models.RetrieveDescription(insight.Description.(models.InsightData)))
	assert.Equal(t, "Female USA 25-34 3.5000005", models.RetrieveDescription(audience.Description.(models.AudienceData)))
}
