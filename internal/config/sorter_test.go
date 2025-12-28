package config

import (
	"testing"
)

func TestSortResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []Resource
		wantErr   bool
		wantLevel int
	}{
		{
			name: "Basit Doğrusal Bağımlılık",
			resources: []Resource{
				{Name: "app", DependsOn: []string{"pkg"}},
				{Name: "pkg", DependsOn: []string{}},
			},
			wantErr:   false,
			wantLevel: 2,
		},
		{
			name: "Karmaşık Paralel Bağımlılık",
			resources: []Resource{
				{Name: "service", DependsOn: []string{"config"}},
				{Name: "config", DependsOn: []string{"pkg"}},
				{Name: "pkg", DependsOn: []string{}},
				{Name: "unrelated", DependsOn: []string{}},
			},
			wantErr:   false,
			wantLevel: 3,
		},
		{
			name: "Döngüsel Bağımlılık Hatası",
			resources: []Resource{
				{Name: "A", DependsOn: []string{"B"}},
				{Name: "B", DependsOn: []string{"A"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SortResources(tt.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("SortResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got) != tt.wantLevel {
				t.Errorf("%s: Katman sayısı hatalı. Got %d, Want %d", tt.name, len(got), tt.wantLevel)
			}
		})
	}
}
