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
	"fmt"
	"math/rand"
	"testing"

	"github.com/WeiWeiWesley/open-match/pkg/pb"

	"github.com/stretchr/testify/require"
)

func TestMakeMatchesDeduplicate(t *testing.T) {
	require := require.New(t)

	poolNameToTickets := map[string][]*pb.Ticket{
		"pool1": {{Id: "1"}},
		"pool2": {{Id: "1"}},
	}

	matches, err := makeMatches(poolNameToTickets)
	require.Nil(err)
	require.Equal(len(matches), 0)
}

func TestMakeMatches(t *testing.T) {
	require := require.New(t)

	poolNameToTickets := map[string][]*pb.Ticket{
		"pool1": {{Id: "1"}, {Id: "2"}, {Id: "3"}},
		"pool2": {{Id: "4"}},
		"pool3": {{Id: "5"}, {Id: "6"}, {Id: "7"}},
	}

	matches, err := makeMatches(poolNameToTickets)
	require.Nil(err)
	require.Equal(len(matches), 3)

	for _, match := range matches {
		require.Equal(2, len(match.Tickets))
		require.Equal(matchName, match.MatchFunction)
	}
}

func BenchmarkMakeMatches(b *testing.B) {
	for i := 0; i < b.N; i++ {

		poolNameToTickets, ticketsCount := requestCreator()
		equal := int(ticketsCount / 2) //期望總組數

		matches, err := makeMatches(poolNameToTickets)
		if err != nil {
			b.Error(err)
		}

		if len(matches) != equal {
			b.Errorf("should create total num of %d matches but got %d", equal, len(matches))
		}

		for _, match := range matches {
			if len(match.Tickets) != 2 {
				b.Error("match.Tickets count err")
			}

			if match.MatchFunction != matchName {
				b.Error("matchName err")
			}
		}
	}
}

//隨機產生請求
func requestCreator() (map[string][]*pb.Ticket, int) {
	poolNum := rand.Intn(9) + 1 // random numbers of pools 1~10
	poolNameToTickets := map[string][]*pb.Ticket{}

	ticketCount := 1
	for i := 0; i < poolNum; i++ {
		aPoolTicketNum := rand.Intn(99) + 1 // random numbers of tickets 1~100
		poolName := fmt.Sprintf("pool%d", i)

		for j := 0; j < aPoolTicketNum; j++ {
			poolNameToTickets[poolName] = append(poolNameToTickets[poolName], &pb.Ticket{
				Id: fmt.Sprintf("%d", ticketCount),
			})
		}
	}

	return poolNameToTickets, ticketCount - 1
}
