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

// Package mmf provides a sample match function that uses the GRPC harness to set up 1v1 matches.
// This sample is a reference to demonstrate the usage of the GRPC harness and should only be used as
// a starting point for your match function. You will need to modify the
// matchmaking logic in this function based on your game's requirements.
package mmf

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/WeiWeiWesley/open-match/pkg/matchfunction"
	"github.com/WeiWeiWesley/open-match/pkg/pb"
	"github.com/bwmarrin/snowflake"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
)

var (
	normalMatchName      = "3v3_normal_battle_royale_matchfunction"
	rankMatchName        = "3v3_rank_battle_royale_matchfunction"
	normalMatchIDCreator *snowflake.Node
	rankMatchIDCreator   *snowflake.Node
)

// matchFunctionService implements pb.MatchFunctionServer, the server generated
// by compiling the protobuf, by fulfilling the pb.MatchFunctionServer interface.
type matchFunctionService struct {
	grpc               *grpc.Server
	queryServiceClient pb.QueryServiceClient
	port               int
}

func init() {
	node, err := snowflake.NewNode(rand.Int63n(50))
	if err != nil {
		panic(err)
	}
	normalMatchIDCreator = node

	node2, err := snowflake.NewNode(51 + rand.Int63n(49))
	if err != nil {
		panic(err)
	}
	rankMatchIDCreator = node2
}

func makeMatches(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	var matches []*pb.Match

	nm, err := normalMatch(p, poolTickets)
	if err != nil {
		log.Println(err)
		return matches, err
	}

	matches = append(matches, nm...)

	rm, err := rankMatch(p, poolTickets)
	if err != nil {
		log.Println(err)
		return matches, err
	}

	matches = append(matches, rm...)

	return matches, nil
}

// Run is this match function's implementation of the gRPC call defined in api/matchfunction.proto.
func (s *matchFunctionService) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	// Fetch tickets for the pools specified in the Match Profile.
	log.Printf("Generating proposals for function %v", req.GetProfile().GetName())

	poolTickets, err := matchfunction.QueryPools(stream.Context(), s.queryServiceClient, req.GetProfile().GetPools())
	if err != nil {
		log.Printf("Failed to query tickets for the given pools, got %s", err.Error())
		return err
	}

	// Generate proposals.
	proposals, err := makeMatches(req.GetProfile(), poolTickets)
	if err != nil {
		log.Printf("Failed to generate matches, got %s", err.Error())
		return err
	}

	log.Printf("Streaming %v proposals to Open Match", len(proposals))
	// Stream the generated proposals back to Open Match.
	for _, proposal := range proposals {
		if err := stream.Send(&pb.RunResponse{Proposal: proposal}); err != nil {
			log.Printf("Failed to stream proposals to Open Match, got %s", err.Error())
			return err
		}
	}

	return nil
}

// Compute the quality as a difference in the highest and lowest player skill levels. This can be used to determine if the match is outside a given skill differential
func computeQuality(tickets []*pb.Ticket) float64 {
	quality := 0.0
	high := 0.0
	low := tickets[0].SearchFields.DoubleArgs["score"]
	for _, ticket := range tickets {
		if high < ticket.SearchFields.DoubleArgs["score"] {
			high = ticket.SearchFields.DoubleArgs["score"]
		}
		if low > ticket.SearchFields.DoubleArgs["score"] {
			low = ticket.SearchFields.DoubleArgs["score"]
		}
	}
	quality = high - low

	return quality
}

func normalMatch(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	var matches []*pb.Match
	if p.Name != "3v3_normal_battle_royale" {
		return nil, nil
	}

	team := &pb.Match{}
	count := 0
	roleInTeam := []string{}

	for i := range p.Pools {
		poolName := p.Pools[i].Name

		if tickets, ok := poolTickets[poolName]; ok {
			for j := range tickets {
				if len(team.Tickets) < 3 {
					//check deduplicated role
					if stringInArr(tickets[j].SearchFields.StringArgs["role"], roleInTeam) {
						continue
					}

					team.Tickets = append(team.Tickets, tickets[j])
					roleInTeam = append(roleInTeam, tickets[j].SearchFields.StringArgs["role"])

					if len(team.Tickets) == 3 {
						// Compute the match quality/score
						matchQuality := computeQuality(team.Tickets)
						evaluationInput, err := ptypes.MarshalAny(&pb.DefaultEvaluationCriteria{
							Score: matchQuality,
						})
						if err != nil {
							return nil, err
						}
						team.MatchId = fmt.Sprintf("profile-%v-time-%v-%d", poolName, time.Now().Format("2006-01-02T15:04:05.00"), rankMatchIDCreator.Generate().Int64()+int64(count))
						team.MatchFunction = normalMatchName
						team.MatchProfile = poolName
						team.Extensions = map[string]*any.Any{
							"evaluation_input": evaluationInput,
						}

						matches = append(matches, team)

						//reset
						team = &pb.Match{}
						roleInTeam = []string{}
						count++
					}
				}
			}
		}
	}

	return matches, nil

}

func rankMatch(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	var matches []*pb.Match

	if p.Name != "3v3_rank_battle_royale" {
		return matches, nil
	}

	for i := range p.Pools {
		poolName := p.Pools[i].Name

		m, err := rankTeam(p, poolName, poolTickets)
		if err != nil {
			return matches, err
		}
		matches = append(matches, m...)

	}

	return matches, nil

}

func rankTeam(p *pb.MatchProfile, poolName string, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	matches := []*pb.Match{}
	team := &pb.Match{}
	roleInTeam := []string{}
	count := 0

	if tickets, ok := poolTickets[poolName]; ok {
		for j := range tickets {
			if len(team.Tickets) < 3 {
				//check deduplicated role
				if stringInArr(tickets[j].SearchFields.StringArgs["role"], roleInTeam) {
					continue
				}

				team.Tickets = append(team.Tickets, tickets[j])
				roleInTeam = append(roleInTeam, tickets[j].SearchFields.StringArgs["role"])

				if len(team.Tickets) == 3 {
					// Compute the match quality/score
					matchQuality := computeQuality(team.Tickets)
					evaluationInput, err := ptypes.MarshalAny(&pb.DefaultEvaluationCriteria{
						Score: matchQuality,
					})
					if err != nil {
						return nil, err
					}

					team.MatchId = fmt.Sprintf("profile-%v-time-%v-%d", poolName, time.Now().Format("2006-01-02T15:04:05.00"), rankMatchIDCreator.Generate().Int64()+int64(count))
					team.MatchFunction = rankMatchName
					team.MatchProfile = p.GetName()
					team.Extensions = map[string]*any.Any{
						"evaluation_input": evaluationInput,
					}

					matches = append(matches, team)
					team = &pb.Match{}
					roleInTeam = []string{}
					count++
				}
			}
		}
	}

	return matches, nil
}

//StringInArr check if string in array
func stringInArr(find string, array []string) (exists bool) {
	for i := range array {
		if find == array[i] {
			return true
		}
	}
	return
}
