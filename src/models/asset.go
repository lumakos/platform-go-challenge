package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type AssetType string

func (a AssetType) String() {
	panic("unimplemented")
}

const (
	Chart    AssetType = "Chart"
	Insight  AssetType = "Insight"
	Audience AssetType = "Audience"
)

type BaseAsset interface {
	GetDescription() string
}

type Asset struct {
	ID          uint      `json:"id"`
	Type        AssetType `json:"type"`
	Description interface{}
	Data        json.RawMessage `json:"data,omitempty"`
}

type ChartData struct {
	Title string    `json:"title,omitempty"`
	XAxis string    `json:"x_axis"`
	YAxis string    `json:"y_axis"`
	Data  []float64 `json:"data"`
}

func (c ChartData) GetDescription() string {
	return c.Title + " " + c.XAxis + " " + c.YAxis
}

type InsightData struct {
	Text string `json:"text"`
}

func (i InsightData) GetDescription() string {
	return i.Text
}

type AudienceData struct {
	Gender           string  `json:"gender"`
	BirthCountry     string  `json:"birth_country"`
	AgeRange         string  `json:"age_range"`
	SocialMediaUsage float64 `json:"social_media_usage"`
	PurchasesLastMo  uint    `json:"purchases_last_month"`
}

func (a AudienceData) GetDescription() string {
	return a.Gender + " " + a.BirthCountry + " " + a.AgeRange + " " + fmt.Sprintf("%f", a.SocialMediaUsage) + strconv.FormatUint(uint64(a.PurchasesLastMo), 10)
}

func RetrieveDescription(b BaseAsset) string {
	return b.GetDescription()
}
