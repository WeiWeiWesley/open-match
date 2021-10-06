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
	"log"

	"github.com/WeiWeiWesley/open-match/pkg/matchfunction"
	"github.com/WeiWeiWesley/open-match/pkg/pb"
	"google.golang.org/grpc"
)

var (
	matchName = "3v3_normal_battle_royale_matchfunction"
)

// matchFunctionService implements pb.MatchFunctionServer, the server generated
// by compiling the protobuf, by fulfilling the pb.MatchFunctionServer interface.
type matchFunctionService struct {
	grpc               *grpc.Server
	queryServiceClient pb.QueryServiceClient
	port               int
}

func makeMatches(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
	var matches []*pb.Match

	team := &pb.Match{
		MatchFunction: matchName,
	}
	for i := range p.Pools {
		poolName := p.Pools[i].Name

		if tickets, ok := poolTickets[poolName]; ok {
			for j := range tickets {
				if len(team.Tickets) < 3 {
					team.Tickets = append(team.Tickets, tickets[j])
					if len(team.Tickets) == 3 {
						matches = append(matches, team)
						team = &pb.Match{}
					}
				}
			}
		}
	}

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
	low := tickets[0].SearchFields.DoubleArgs["level"]
	for _, ticket := range tickets {
		if high < ticket.SearchFields.DoubleArgs["level"] {
			high = ticket.SearchFields.DoubleArgs["level"]
		}
		if low > ticket.SearchFields.DoubleArgs["level"] {
			low = ticket.SearchFields.DoubleArgs["level"]
		}
	}
	quality = high - low

	return quality
}
