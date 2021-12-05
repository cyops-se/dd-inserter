package types

// Imported meta data
type Meta struct {
	TagId       int     `json:"tag_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Type        string  `json:"type"`
	Unit        string  `json:"unit"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
}
