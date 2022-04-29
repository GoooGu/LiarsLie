package liars_network

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	port_number int
	grpc_server *grpc.Server
	value       int32
	UnimplementedLieServiceServer
}

func (agent *Agent) Init(port_number chan int, agent_value int32, wg *sync.WaitGroup) {
	agent.value = agent_value

	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalln("Failed to find the next available port: ", err)
	}
	// This port number is the next arbitrary free available one in the network
	agent.port_number = conn.Addr().(*net.TCPAddr).Port
	port_number <- agent.port_number
	// Creates a grpc server over the port that was just found
	agent.grpc_server = grpc.NewServer()
	RegisterLieServiceServer(agent.grpc_server, agent)
	wg.Done()
	if err := agent.grpc_server.Serve(conn); err != nil {
		log.Fatalln("Failed to serve gRPC server over port: ", agent.port_number, err)
	}
}

func (agent *Agent) LieQuery(ctx context.Context, lie_request *LieRequest) (*LieResponse, error) {
	if lie_request.GetExpertMode() {
		var collected_agent_values []int32
		for _, port_number := range lie_request.GetOtherAgentIds() {
			conn, err := grpc.Dial(":"+port_number, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			defer conn.Close()
			internal_client := NewLieServiceClient(conn)
			response, err := internal_client.LieQuery(context.Background(), new(LieRequest))
			if err != nil {
				return nil, err
			}
			collected_agent_values = append(collected_agent_values, response.AgentValue)
		}
		return &LieResponse{CollectedAgentValues: collected_agent_values, AgentValue: agent.value}, nil
	}
	return &LieResponse{AgentValue: agent.value}, nil
}

func (agent *Agent) Stop() {
	fmt.Println("Stopping grpc server on port number: ", agent.port_number)
	agent.grpc_server.Stop()
}

func (agent *Agent) UpdateValue(value int32) {
	agent.value = value
}

func (agent *Agent) IsMatchingPortNumber(port_number int) bool {
	return agent.port_number == port_number
}

func (agent *Agent) RetrievePortNum() string {
	return strconv.FormatInt(int64(agent.port_number), 10)
}
