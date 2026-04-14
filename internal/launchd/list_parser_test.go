package launchd

import "testing"

func TestParseListOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []ListEntry
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "header only",
			input: "PID\tStatus\tLabel\n",
			want:  nil,
		},
		{
			name: "single running job",
			input: "PID\tStatus\tLabel\n" +
				"584\t0\tcom.example.myapp\n",
			want: []ListEntry{
				{Label: "com.example.myapp", PID: 584, LastExitStatus: 0},
			},
		},
		{
			name: "single stopped job",
			input: "PID\tStatus\tLabel\n" +
				"-\t0\tcom.example.stopped\n",
			want: []ListEntry{
				{Label: "com.example.stopped", PID: 0, LastExitStatus: 0},
			},
		},
		{
			name: "single error job",
			input: "PID\tStatus\tLabel\n" +
				"-\t78\tcom.example.broken\n",
			want: []ListEntry{
				{Label: "com.example.broken", PID: 0, LastExitStatus: 78},
			},
		},
		{
			name: "mixed running stopped error",
			input: "PID\tStatus\tLabel\n" +
				"584\t0\tcom.example.myapp\n" +
				"-\t0\tcom.example.stopped\n" +
				"-\t78\tcom.example.broken\n",
			want: []ListEntry{
				{Label: "com.example.myapp", PID: 584, LastExitStatus: 0},
				{Label: "com.example.stopped", PID: 0, LastExitStatus: 0},
				{Label: "com.example.broken", PID: 0, LastExitStatus: 78},
			},
		},
		{
			name: "negative exit status (signal termination)",
			input: "PID\tStatus\tLabel\n" +
				"-\t-9\tcom.example.killed\n" +
				"-\t-15\tcom.example.termed\n",
			want: []ListEntry{
				{Label: "com.example.killed", PID: 0, LastExitStatus: -9},
				{Label: "com.example.termed", PID: 0, LastExitStatus: -15},
			},
		},
		{
			name: "malformed line too few columns skipped",
			input: "PID\tStatus\tLabel\n" +
				"584\t0\tcom.example.good\n" +
				"bad-line\n" +
				"-\t0\tcom.example.also-good\n",
			want: []ListEntry{
				{Label: "com.example.good", PID: 584, LastExitStatus: 0},
				{Label: "com.example.also-good", PID: 0, LastExitStatus: 0},
			},
		},
		{
			name: "malformed PID column skipped",
			input: "PID\tStatus\tLabel\n" +
				"abc\t0\tcom.example.badpid\n" +
				"584\t0\tcom.example.good\n",
			want: []ListEntry{
				{Label: "com.example.good", PID: 584, LastExitStatus: 0},
			},
		},
		{
			name: "malformed status column skipped",
			input: "PID\tStatus\tLabel\n" +
				"584\tabc\tcom.example.badstatus\n" +
				"-\t0\tcom.example.good\n",
			want: []ListEntry{
				{Label: "com.example.good", PID: 0, LastExitStatus: 0},
			},
		},
		{
			name: "trailing newline handled",
			input: "PID\tStatus\tLabel\n" +
				"584\t0\tcom.example.myapp\n" +
				"\n",
			want: []ListEntry{
				{Label: "com.example.myapp", PID: 584, LastExitStatus: 0},
			},
		},
		{
			name: "no header present (raw data)",
			input: "584\t0\tcom.example.myapp\n" +
				"-\t78\tcom.example.broken\n",
			want: []ListEntry{
				{Label: "com.example.myapp", PID: 584, LastExitStatus: 0},
				{Label: "com.example.broken", PID: 0, LastExitStatus: 78},
			},
		},
		{
			name: "label with dots dashes underscores",
			input: "PID\tStatus\tLabel\n" +
				"100\t0\tcom.apple.some-daemon_v2.3\n",
			want: []ListEntry{
				{Label: "com.apple.some-daemon_v2.3", PID: 100, LastExitStatus: 0},
			},
		},
		{
			name: "large PID and exit status",
			input: "PID\tStatus\tLabel\n" +
				"99999\t255\tcom.example.big\n",
			want: []ListEntry{
				{Label: "com.example.big", PID: 99999, LastExitStatus: 255},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseListOutput(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("length mismatch: got %d, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}

			for i, w := range tt.want {
				g := got[i]
				if g.Label != w.Label {
					t.Errorf("[%d] Label: got %q, want %q", i, g.Label, w.Label)
				}
				if g.PID != w.PID {
					t.Errorf("[%d] PID: got %d, want %d", i, g.PID, w.PID)
				}
				if g.LastExitStatus != w.LastExitStatus {
					t.Errorf("[%d] LastExitStatus: got %d, want %d", i, g.LastExitStatus, w.LastExitStatus)
				}
			}
		})
	}
}

func TestListEntry_DeriveStatus(t *testing.T) {
	tests := []struct {
		name  string
		entry ListEntry
		want  JobStatus
	}{
		{"running", ListEntry{PID: 584, LastExitStatus: 0}, StatusRunning},
		{"stopped", ListEntry{PID: 0, LastExitStatus: 0}, StatusStopped},
		{"error", ListEntry{PID: 0, LastExitStatus: 78}, StatusError},
		{"signal killed", ListEntry{PID: 0, LastExitStatus: -9}, StatusError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveStatus(tt.entry.PID, tt.entry.LastExitStatus)
			if got != tt.want {
				t.Errorf("DeriveStatus(%d, %d) = %q, want %q",
					tt.entry.PID, tt.entry.LastExitStatus, got, tt.want)
			}
		})
	}
}
