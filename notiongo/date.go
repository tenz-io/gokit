package notiongo

type Date struct {
	Start    string `json:"start,omitempty"`
	End      string `json:"end,omitempty"`
	TimeZone string `json:"time_zone,omitempty"`
}

func (d *Date) FromMap(dateMap map[string]any) *Date {
	d.Start = ValueOrDefault(dateMap["start"], "")
	d.End = ValueOrDefault(dateMap["end"], "")
	d.TimeZone = ValueOrDefault(dateMap["time_zone"], "")
	return d
}

type DateProperty struct {
	*property
	Date *Date `json:"date,omitempty"`
}

// NewDatePropertyWithBase creates a new date property.
func NewDatePropertyWithBase(base *property) *DateProperty {
	return &DateProperty{
		property: base,
	}
}

// FromMap converts a map to a DateProperty.
func (d *DateProperty) FromMap(m map[string]any) *DateProperty {
	if dataMap := ValueOrDefault(m["date"], map[string]any{}); len(dataMap) >= 0 {
		d.Date = (&Date{}).FromMap(dataMap)
	}
	return d
}
