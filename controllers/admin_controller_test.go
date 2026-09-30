package controllers

import "testing"

func TestValidRequestAccountEmail(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		requestType string
		want        bool
	}{
		{
			name:        "student account uses student domain",
			address:     "fikrisujud.241011028@mahasiswa.ith.ac.id",
			requestType: "Mahasiswa ITH",
			want:        true,
		},
		{
			name:        "student account rejects staff domain",
			address:     "fikrisujud.241011028@ith.ac.id",
			requestType: "Mahasiswa ITH",
			want:        false,
		},
		{
			name:        "staff account uses primary domain",
			address:     "fikrisujud@ith.ac.id",
			requestType: "Dosen / Pegawai ITH",
			want:        true,
		},
		{
			name:        "staff account rejects student domain",
			address:     "fikrisujud@mahasiswa.ith.ac.id",
			requestType: "Dosen / Pegawai ITH",
			want:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validRequestAccountEmail(test.address, test.requestType); got != test.want {
				t.Errorf("validRequestAccountEmail(%q, %q) = %t, want %t", test.address, test.requestType, got, test.want)
			}
		})
	}
}
