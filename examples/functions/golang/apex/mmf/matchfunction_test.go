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
		"3v3_normal_battle_royale": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "bang",
						"user_id":     "123456789",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             202,
						"rank":              3,
						"team_member_count": 0,
						"score":             3923,
						"avg_dmg":           346,
						"avg_kd":            0.22,
						"win_streak":        0,
					},
				},
			},
			{
				Id: "2",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "valk",
						"user_id":     "123456333",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             203,
						"rank":              3,
						"team_member_count": 0,
						"score":             3701,
						"avg_dmg":           311,
						"avg_kd":            0.25,
						"win_streak":        0,
					},
				},
			},
		},
	}

	p := &pb.MatchProfile{
		Name: "3v3_normal_battle_royale",
		Pools: []*pb.Pool{
			{
				Name: "3v3_normal_battle_royale",
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
		"3v3_normal_battle_royale": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "bang",
						"user_id":     "123456789",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             202,
						"rank":              3,
						"team_member_count": 0,
						"score":             3923,
						"avg_dmg":           346,
						"avg_kd":            0.22,
						"win_streak":        0,
					},
				},
			},
			{
				Id: "2",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "valk",
						"user_id":     "123456333",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             203,
						"rank":              3,
						"team_member_count": 0,
						"score":             3701,
						"avg_dmg":           311,
						"avg_kd":            0.25,
						"win_streak":        0,
					},
				},
			},
			{
				Id: "3",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "caustic",
						"user_id":     "123456333",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             203,
						"rank":              3,
						"team_member_count": 0,
						"score":             3701,
						"avg_dmg":           311,
						"avg_kd":            0.25,
						"win_streak":        0,
					},
				},
			},
			{
				Id: "4",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "horizon",
						"user_id":     "1771193153",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             197,
						"rank":              3,
						"team_member_count": 0,
						"score":             3691,
						"avg_dmg":           289,
						"avg_kd":            0.27,
						"win_streak":        0,
					},
				},
			},
		},
	}

	p := &pb.MatchProfile{
		Name: "3v3_normal_battle_royale",
		Pools: []*pb.Pool{
			{
				Name: "3v3_normal_battle_royale",
			},
		},
	}

	matches, err := makeMatches(p, poolNameToTickets)
	require.Nil(err)
	require.Equal(1, len(matches))

	for _, match := range matches {
		require.Equal(3, len(match.Tickets))
		require.Equal(normalMatchName, match.MatchFunction)
	}
}

func TestNoMatches(t *testing.T) {
	require := require.New(t)

	poolNameToTickets := map[string][]*pb.Ticket{
		"3v3_normal_battle_royale": {
			{
				Id: "1",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "bang",
						"user_id":     "123456789",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             202,
						"rank":              3,
						"team_member_count": 0,
						"score":             3923,
						"avg_dmg":           346,
						"avg_kd":            0.22,
						"win_streak":        0,
					},
				},
			},
			{
				Id: "2",
				SearchFields: &pb.SearchFields{
					StringArgs: map[string]string{
						"mode":        "3v3_normal_battle_royale",
						"server":      "Taiwan_GCE2",
						"role":        "valk",
						"user_id":     "123456333",
						"team_member": `[]`,
						"black_list":  `[]`,
					},
					DoubleArgs: map[string]float64{
						"level":             203,
						"rank":              3,
						"team_member_count": 0,
						"score":             3701,
						"avg_dmg":           311,
						"avg_kd":            0.25,
						"win_streak":        0,
					},
				},
			},
		},
	}

	p := &pb.MatchProfile{
		Name: "3v3_normal_battle_royale",
		Pools: []*pb.Pool{
			{
				Name: "3v3_normal_battle_royale",
			},
		},
	}

	matches, err := makeMatches(p, poolNameToTickets)
	require.Nil(err)
	require.Equal(0, len(matches))

	for _, match := range matches {
		require.Equal(3, len(match.Tickets))
		require.Equal(normalMatchName, match.MatchFunction)
	}
}
