// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mmf

import (
	"testing"

	"github.com/WeiWeiWesley/open-match/pkg/pb"

	"github.com/stretchr/testify/require"
)

func TestMakeMatchesDeduplicate(t *testing.T) {
	require := require.New(t)

	poolNameToTickets := map[string][]*pb.Ticket{
		"pool1": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 3,
					},
				},
			},
		},
		"pool2": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 24,
					},
				},
			},
		},
	}

	p := &pb.MatchProfile{
		Name: "MatchProfile",
		Pools: []*pb.Pool{
			{
				Name: "A pool",
				DoubleRangeFilters: []*pb.DoubleRangeFilter{
					{
						DoubleArg: "level",
						Max:       30,
						Min:       25,
					},
				},
			},
		},
	}

	matches, err := makeMatches(p, poolNameToTickets)
	require.Nil(err)
	require.Equal(len(matches), 0)
}

func TestMakeMatches(t *testing.T) {
	require := require.New(t)

	poolNameToTickets := map[string][]*pb.Ticket{
		"pool1": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 23,
					},
				},
			},
			{
				Id: "2",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 24,
					},
				},
			},
		},
		"pool2": {
			{
				Id: "3",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 25,
					},
				},
			},
			{
				Id: "4",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 26,
					},
				},
			},
		},
		"pool3": {
			{
				Id: "5",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 30,
					},
				},
			},
			{
				Id: "6",
				SearchFields: &pb.SearchFields{
					DoubleArgs: map[string]float64{
						"level": 30,
					},
				},
			},
		},
	}

	p := &pb.MatchProfile{
		Name: "MatchProfile",
		Pools: []*pb.Pool{
			{
				Name: "A pool",
				DoubleRangeFilters: []*pb.DoubleRangeFilter{
					{
						DoubleArg: "level",
						Max:       30,
						Min:       25,
					},
				},
			},
		},
	}

	matches, err := makeMatches(p, poolNameToTickets)
	require.Nil(err)
	require.Equal(len(matches), 3)

	for _, match := range matches {
		require.Equal(2, len(match.Tickets))
		require.Equal(matchName, match.MatchFunction)
	}
}
