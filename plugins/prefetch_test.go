package plugins

import (
	"github.com/forensicanalysis/forensicstore/goforensicstore"
	"log"
	"path/filepath"
	"testing"
)

func TestPrefetchPlugin_Run(t *testing.T) {
	log.Println("Start setup")
	storeDir, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Setup done")
	defer cleanup(storeDir)

	type args struct {
		storeName string
		data      Data
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{"Prefetch Test", args{"example1.forensicstore", nil}, 261, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PrefetchPlugin{}

			url := filepath.Join(storeDir, tt.args.storeName)
			if err := pr.Run(url, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			store, err := goforensicstore.NewJSONLite(url)
			if err != nil {
				t.Errorf("goforensicstore.NewJSONLite() error = %v, wantErr %v", err, tt.wantErr)
			}
			items, err := store.Select("prefetch")
			if err != nil {
				t.Errorf("store.All() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(items) != tt.wantCount {
				t.Errorf("len(items) = %v, wantCount %v", len(items), tt.wantCount)
			}

		})
	}
}
