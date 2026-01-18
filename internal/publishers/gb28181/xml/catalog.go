package xml

import (
	"encoding/xml"
	"fmt"
)

// CatalogQuery represents a Catalog query request
type CatalogQuery struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

// ParseCatalogQuery parses a Catalog query from XML
func ParseCatalogQuery(data []byte) (*CatalogQuery, error) {
	var q CatalogQuery
	if err := xml.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("parse catalog query: %w", err)
	}
	return &q, nil
}

// CatalogResponse represents a Catalog query response
type CatalogResponse struct {
	XMLName    xml.Name        `xml:"Response"`
	CmdType    CmdType         `xml:"CmdType"`
	SN         int             `xml:"SN"`
	DeviceID   string          `xml:"DeviceID"`
	SumNum     int             `xml:"SumNum"`
	DeviceList CatalogItemList `xml:"DeviceList"`
}

// CatalogItemList wraps the device list
type CatalogItemList struct {
	Num   int           `xml:"Num,attr"`
	Items []CatalogItem `xml:"Item"`
}

// CatalogItem represents a device or channel in the catalog
type CatalogItem struct {
	DeviceID     string `xml:"DeviceID"`
	Name         string `xml:"Name"`
	Manufacturer string `xml:"Manufacturer"`
	Model        string `xml:"Model"`
	Owner        string `xml:"Owner"`
	CivilCode    string `xml:"CivilCode"`
	Address      string `xml:"Address"`
	Parental     int    `xml:"Parental"`
	ParentID     string `xml:"ParentID"`
	SafetyWay    int    `xml:"SafetyWay"`
	RegisterWay  int    `xml:"RegisterWay"`
	Secrecy      int    `xml:"Secrecy"`
	Status       string `xml:"Status"`
	Longitude    string `xml:"Longitude,omitempty"`
	Latitude     string `xml:"Latitude,omitempty"`
}

// NewCatalogResponse creates a Catalog response
func NewCatalogResponse(deviceID string, sn int, items []CatalogItem) *CatalogResponse {
	return &CatalogResponse{
		CmdType:  CmdTypeCatalog,
		SN:       sn,
		DeviceID: deviceID,
		SumNum:   len(items),
		DeviceList: CatalogItemList{
			Num:   len(items),
			Items: items,
		},
	}
}

// NewCatalogItem creates a catalog item for a drone
func NewCatalogItem(deviceID, name, parentID, civilCode string, online bool) CatalogItem {
	status := "OFF"
	if online {
		status = "ON"
	}
	return CatalogItem{
		DeviceID:     deviceID,
		Name:         name,
		Manufacturer: "UAV-Gateway",
		Model:        "OUTB",
		Owner:        "UAV-Gateway",
		CivilCode:    civilCode,
		Address:      "UAV Location",
		Parental:     0, // No children
		ParentID:     parentID,
		SafetyWay:    0,
		RegisterWay:  1, // Standard registration
		Secrecy:      0,
		Status:       status,
	}
}

// Marshal serializes the catalog response to XML with declaration
func (r *CatalogResponse) Marshal() (string, error) {
	data, err := xml.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal catalog response: %w", err)
	}
	return XMLDeclaration + "\r\n" + string(data), nil
}
