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

package clients

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc"

	"github.com/WeiWeiWesley/open-match/examples/demo/components"
	"github.com/WeiWeiWesley/open-match/examples/demo/updater"
	"github.com/WeiWeiWesley/open-match/pkg/pb"
	"github.com/bwmarrin/snowflake"
)

var idCreator *snowflake.Node

func init() {
	node, err := snowflake.NewNode(int64(rand.Intn(100)))
	if err != nil {
		panic(err)
	}

	idCreator = node
}

func Run(ds *components.DemoShared) {
	u := updater.NewNested(ds.Ctx, ds.Update)

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("fakeplayer_%d", i)
		go func() {
			for !isContextDone(ds.Ctx) {
				runScenario(ds.Ctx, name, u.ForField(name))
			}
		}()
	}
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

type status struct {
	Status     string
	Assignment *pb.Assignment
}

func runScenario(ctx context.Context, name string, update updater.SetFunc) {
	defer func() {
		r := recover()
		if r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}

			update(status{Status: fmt.Sprintf("Encountered error: %s", err.Error())})
			time.Sleep(time.Second * 10)
		}
	}()

	s := status{}

	//////////////////////////////////////////////////////////////////////////////
	s.Status = "Main Menu"
	update(s)

	time.Sleep(time.Duration(rand.Int63()) % (time.Second * 15))

	//////////////////////////////////////////////////////////////////////////////
	s.Status = "Connecting to Open Match frontend"
	update(s)

	// See https://open-match.dev/site/docs/guides/api/
	conn, err := grpc.Dial("open-match-frontend.open-match.svc.cluster.local:50504", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fe := pb.NewFrontendServiceClient(conn)

	//////////////////////////////////////////////////////////////////////////////
	s.Status = "Creating Open Match Ticket"
	update(s)

	var ticketId string
	{
		req := &pb.CreateTicketRequest{
			Ticket: ticketCreator(),
		}

		resp, err := fe.CreateTicket(ctx, req)
		if err != nil {
			panic(err)
		}
		ticketId = resp.Id
	}

	//////////////////////////////////////////////////////////////////////////////
	s.Status = fmt.Sprintf("Waiting match with ticket Id %s", ticketId)
	update(s)

	var assignment *pb.Assignment
	{
		req := &pb.WatchAssignmentsRequest{
			TicketId: ticketId,
		}

		stream, err := fe.WatchAssignments(ctx, req)
		for assignment.GetConnection() == "" {
			resp, err := stream.Recv()
			if err != nil {
				// For now we don't expect to get EOF, so that's still an error worthy of panic.
				panic(err)
			}

			assignment = resp.Assignment
		}

		err = stream.CloseSend()
		if err != nil {
			panic(err)
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	s.Status = fmt.Sprintf("Sleeping (pretend this %s is playing a match...)", ticketId)
	s.Assignment = assignment
	update(s)

	time.Sleep(time.Second * 10)
}

func ticketCreator() (ticket *pb.Ticket) {
	rank := rand.Intn(6) // 0未分類 1銅 2銀 3金 4白金 5鑽石 6大師
	score, avgDmg, avgKd, level := creatScoreByRank(rank)

	switch time.Now().Unix() % 2 {
	case 0:
		/*
			台服
			一般場
			無隊友
			無連勝
		*/
		ticket = &pb.Ticket{
			SearchFields: &pb.SearchFields{
				StringArgs: map[string]string{
					"mode":        "3v3_normal_battle_royale",
					"server":      "Taiwan_GCE2",
					"role":        roleCreator(),
					"user_id":     idCreator.Generate().String(),
					"team_member": `[]`,
					"black_list":  `[]`,
				},
				DoubleArgs: map[string]float64{
					"level":             level,
					"rank":              float64(rank),
					"team_member_count": 0,
					"score":             score,
					"avg_dmg":           avgDmg,
					"avg_kd":            avgKd,
					"win_streak":        0,
				},
			},
		}
	case 1:
		/*
			台服
			排位場
			無隊友
			無連勝
		*/
		ticket = &pb.Ticket{
			SearchFields: &pb.SearchFields{
				StringArgs: map[string]string{
					"mode":        "3v3_rank_battle_royale",
					"server":      "Taiwan_GCE2",
					"role":        roleCreator(),
					"user_id":     idCreator.Generate().String(),
					"team_member": `[]`,
					"black_list":  `[]`,
				},
				DoubleArgs: map[string]float64{
					"level":             level,
					"rank":              float64(rank),
					"team_member_count": 0,
					"score":             score,
					"avg_dmg":           avgDmg,
					"avg_kd":            avgKd,
					"win_streak":        0,
				},
			},
		}
	}

	return
}

func roleCreator() (role string) {
	switch rand.Intn(7) {
	case 0:
		role = "bang"
	case 1:
		role = "hound"
	case 2:
		role = "gibby"
	case 3:
		role = "valk"
	case 4:
		role = "crypto"
	case 5:
		role = "caustic"
	case 6:
		role = "horizon"
	case 7:
		role = "octane"
	}

	return
}

func creatScoreByRank(rank int) (score float64, avgDmg float64, avgKd float64, level float64) {
	switch rank {
	case 0, 1:
		//Bronze
		return float64(rand.Intn(1199)), float64(rand.Intn(180)), float64(rand.Intn(30)) / 100, float64(1 + rand.Intn(499))
	case 2:
		//Silver
		return float64(1200 + rand.Intn(1599)), float64(rand.Intn(230)), float64(rand.Intn(50)) / 100, float64(11 + rand.Intn(489))
	case 3:
		//Gold
		return float64(2800 + rand.Intn(1999)), float64(200 + rand.Intn(180)), 0.2 + float64(rand.Intn(50))/100, float64(50 + rand.Intn(449))
	case 4:
		//Platinum
		return float64(4800 + rand.Intn(2399)), float64(280 + rand.Intn(200)), 0.5 + float64(rand.Intn(50))/100, float64(50 + rand.Intn(449))
	case 5:
		//Diamond
		return float64(7200 + rand.Intn(2799)), float64(380 + rand.Intn(300)), 0.7 + float64(rand.Intn(30))/100, float64(50 + rand.Intn(449))
	case 6:
		//Master
		return float64(10000 + rand.Intn(3500)), float64(480 + rand.Intn(500)), 1 + float64(rand.Intn(30))/100, float64(50 + rand.Intn(449))
	}

	return
}
