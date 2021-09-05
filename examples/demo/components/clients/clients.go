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
	"open-match.dev/open-match/pkg/pb"
)

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

		ticket := &pb.Ticket{}

		//隨機產生不同條件
		//兩個地區
		//兩種職業
		//等級 0~30
		switch time.Now().UnixNano() % 4 {
		case 0:
			ticket.SearchFields = &pb.SearchFields{
				StringArgs: map[string]string{
					"location": "Asia/Taiwan",
					"role":     "knight",
				},
				DoubleArgs: map[string]float64{
					"level": float64(rand.Int31n(30)),
				},
			}
		case 1:
			ticket.SearchFields = &pb.SearchFields{
				StringArgs: map[string]string{
					"location": "Asia/Taiwan",
					"role":     "archer",
				},
				DoubleArgs: map[string]float64{
					"level": float64(rand.Int31n(30)),
				},
			}
		case 2:
			ticket.SearchFields = &pb.SearchFields{
				StringArgs: map[string]string{
					"location": "Asia/Japan",
					"role":     "knight",
				},
				DoubleArgs: map[string]float64{
					"level": float64(rand.Int31n(30)),
				},
			}
		case 3:
			ticket.SearchFields = &pb.SearchFields{
				StringArgs: map[string]string{
					"location": "Asia/Japan",
					"role":     "archer",
				},
				DoubleArgs: map[string]float64{
					"level": float64(rand.Int31n(30)),
				},
			}
		}

		fmt.Println(ticket.SearchFields.StringArgs["role"])

		req := &pb.CreateTicketRequest{Ticket: ticket}

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
	s.Status = "Sleeping (pretend this is playing a match...)"
	s.Assignment = assignment
	update(s)

	time.Sleep(time.Second * 10)
}
