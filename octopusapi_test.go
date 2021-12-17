package octopusapi

import "testing"

func TestGetBySCode(t *testing.T) {
	tests := []struct {
		Name string
		URL  string
	}{
		{
			Name: "dark-magician-girl",
			URL:  "https://cuddlyoctopus.com/product/dark-magician-girl/",
		}, {
			Name: "Asuna",
			URL:  "https://cuddlyoctopus.com/product/asuna/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := GetProductByURL(tt.URL)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
