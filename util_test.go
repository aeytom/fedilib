package fedilib

import "testing"

func TestStripHtmlFromString(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				in: `<p><span class="h-card" translate="no"><a href="https://troet.tay-tec.de/@allsky" class="u-url mention">@<span>allsky</span></a></span> hello</p>`,
			},
			want:    "@allsky hello\n\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StripHtmlFromString(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("StripHtmlFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StripHtmlFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
