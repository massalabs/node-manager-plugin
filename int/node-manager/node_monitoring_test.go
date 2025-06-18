package nodeManager

import (
	"testing"
)

func TestCheckDesync(t *testing.T) {
	tests := []struct {
		name           string
		prometheusData string
		wantDesync     bool
		wantErr        bool
		errMsg         string
	}{
		{
			name: "node is desynced",
			prometheusData: `# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period 100
# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period 85`,
			wantDesync: true,
			wantErr:    false,
		},
		{
			name: "node is not desynced",
			prometheusData: `# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period 100
# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period 95`,
			wantDesync: false,
			wantErr:    false,
		},
		{
			name: "metrics in different order but node is synced",
			prometheusData: `# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period 95
# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period 100`,
			wantDesync: false,
			wantErr:    false,
		},
		{
			name: "missing active cursor metric",
			prometheusData: `# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period 95`,
			wantDesync: false,
			wantErr:    true,
			errMsg:     "failed to find active_cursor_period metric in prometheus data",
		},
		{
			name: "missing final cursor metric",
			prometheusData: `# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period 100`,
			wantDesync: false,
			wantErr:    true,
			errMsg:     "failed to find final_cursor_period metric in prometheus data",
		},
		{
			name: "invalid active cursor value",
			prometheusData: `# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period invalid
# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period 95`,
			wantDesync: false,
			wantErr:    true,
			errMsg:     "failed to convert active_cursor_period string value invalid to int: strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name: "invalid final cursor value",
			prometheusData: `# HELP active_cursor_period Current active cursor period
# TYPE active_cursor_period gauge
active_cursor_period 100
# HELP final_cursor_period Current final cursor period
# TYPE final_cursor_period gauge
final_cursor_period invalid`,
			wantDesync: false,
			wantErr:    true,
			errMsg:     "failed to convert final_cursor_period string value invalid to int: strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := NewNodeMonitor()
			gotDesync, err := nm.checkDesync([]byte(tt.prometheusData))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Test%s: expected error but got none", tt.name)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Test %s: got error msg \"%v\", want error msg \"%v\"", tt.name, err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Test %s: unexpected error = %v", tt.name, err)
				return
			}

			if gotDesync != tt.wantDesync {
				t.Errorf("Test %s: got result \"%v\", expected result \"%v\"", tt.name, gotDesync, tt.wantDesync)
			}
		})
	}
}
