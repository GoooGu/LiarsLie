package liars_network

import (
	"log"
	"net"
	"google.golang.org/grpc"
)

type Agent struct {
	value int32
	port_number string
}

func (agent *Agent) Init(port_number string, agent_value int32) {
	agent.port_number = port_number
	agent.value = agent_value

	conn, err := net.Listen("tcp", port_number)
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", port_number, err)
	}
	
	grpcServer := grpc.NewServer()

	if err := grpcServer.Serve(conn); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %v: %v", port_number, err)
	}
}