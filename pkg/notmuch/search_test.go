package notmuch

import (
	"github.com/emersion/go-imap"
	"net/textproto"
	"testing"
)

func TestIMAPSearchToNotmuch(t *testing.T) {
	tests := []struct {
		name    string
		search  *imap.SearchCriteria
		want    string
		wantErr bool
	}{
		{
			name: "search with OR",
			search: &imap.SearchCriteria{
				Or: [][2]*imap.SearchCriteria{
					{
						{
							Body: []string{"Retreat schedule"},
						},
						{
							Header: textproto.MIMEHeader{
								"From": []string{"Plum Village"},
							},
						},
					},
				},
			},
			want: `( body:"Retreat schedule" or from:"Plum Village" )`,
		},
		{
			name: "search with NOT",
			search: &imap.SearchCriteria{
				Not: []*imap.SearchCriteria{
					{
						Header: textproto.MIMEHeader{
							"From": []string{"Plum Village"},
						},
					},
				},
			},
			want: `not from:"Plum Village"`,
		},
		{
			name: "search with unsupported fields",
			search: &imap.SearchCriteria{
				Smaller: uint32(256),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IMAPSearchToNotmuch(tt.search, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("IMAPSearchToNotmuch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IMAPSearchToNotmuch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
