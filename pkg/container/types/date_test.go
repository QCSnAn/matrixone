// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateCast(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// 1. yyyy-mm-dd
		{
			name: "yyyy-mm-dd",
			args: args{
				s: "2005-02-23",
			},
			want: "2005-02-23",
		},
		// 2. yyyymmdd
		{
			name: "yyyymmdd",
			args: args{
				s: "20050223",
			},
			want: "2005-02-23",
		},
		// 3. range test
		{
			name: "leap_year",
			args: args{
				s: "19990229",
			},
			wantErr: true,
		},
		{
			name: "month_range1",
			args: args{
				s: "20001329",
			},
			wantErr: true,
		},
		{
			name: "month_range2",
			args: args{
				s: "20000029",
			},
			wantErr: true,
		},
		{
			name: "day_range1",
			args: args{
				s: "20000431",
			},
			wantErr: true,
		},
		{
			name: "day_range2",
			args: args{
				s: "20000400",
			},
			wantErr: true,
		},
		// 4. yyyy-m-dd
		{
			name: "yyyy-m-dd",
			args: args{
				s: "2005-2-23",
			},
			want: "2005-02-23",
		},
		// 5. yyyy-mm-d
		{
			name: "yyyy-mm-d",
			args: args{
				s: "2005-02-2",
			},
			want: "2005-02-02",
		},
		// 6. yyyy-m-d
		{
			name: "yyyy-m-d",
			args: args{
				s: "2005-2-3",
			},
			want: "2005-02-03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateCast(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateCast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				return
			}
			if got.String() != tt.want {
				t.Errorf("ParseDateCast() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_date_toBytes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// 1. yyyy-mm-dd
		{
			name: "yyyy-mm-dd",
			args: args{
				s: "2005-02-23",
			},
			want: "2005-02-23",
		},
		// 2. yyyymmdd
		{
			name: "yyyymmdd",
			args: args{
				s: "20050223",
			},
			want: "2005-02-23",
		},
		// 4. yyyy-m-dd
		{
			name: "yyyy-m-dd",
			args: args{
				s: "2005-2-23",
			},
			want: "2005-02-23",
		},
		// 5. yyyy-mm-d
		{
			name: "yyyy-mm-d",
			args: args{
				s: "2005-02-2",
			},
			want: "2005-02-02",
		},
		// 6. yyyy-m-d
		{
			name: "yyyy-m-d",
			args: args{
				s: "2005-2-3",
			},
			want: "2005-02-03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateCast(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateCast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				return
			}
			var dBytes [DateToBytesLength]byte
			dSlice := got.ToBytes(dBytes[:0])
			s := string(dSlice)
			if s != tt.want {
				t.Errorf("ParseDateCast() got = %v, want %v", s, tt.want)
			}
		})
	}
}

func BenchmarkParseDate(b *testing.B) {
	s := "2020-12-21"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDateCast(s)
		require.NoError(b, err)
	}
}

func Test_date_String(t *testing.T) {
	for i := 0; i < 100; i++ {
		x, y := i/10, i%10
		chs := hundredToChars[i]
		assert.Equal(t, uint8(x+'0'), chs[0])
		assert.Equal(t, uint8(y+'0'), chs[1])
	}
}
