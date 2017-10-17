package pip_service

import (
	"fmt"
	"google.golang.org/grpc"
)

var pIPServiceEndPoints = map[string]string {
	"mcafee-ts": "127.0.0.1:5368",
}

// This will be replaced by service discovery
// This will also need to be able to handle load balancing of multiple
//    endpoint per service
func GetPIPServiceEndpoint(service string) (string, error) {
	endpoint, ok := pIPServiceEndPoints[service]
	if !ok {
		return "", fmt.Errorf("Cannot find PIP Service '%s'", service)
	}

	return endpoint, nil
}

type Connection struct {
	endpoint string
	conn     *grpc.ClientConn
}

type ConnectionManager struct {
	connections map[string]Connection
}

func NewConnectionManager() *ConnectionManager {
	fmt.Printf("Making ConnectionManager\n")
	return &ConnectionManager{connections: make(map[string]Connection, 0)}
}

func (c *ConnectionManager) GetConnection(serviceName string) (*grpc.ClientConn, error) {
	connection, ok := c.connections[serviceName]
	if !ok {
		fmt.Printf("Making connection to pip '%s'\n", serviceName)
		endpoint, err := GetPIPServiceEndpoint(serviceName)
		if err != nil {
			return nil, fmt.Errorf("Cannot get service endpoint for PIP %s'", serviceName)
		}
		conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("Cannot connect to service endpoint for PIP '%s'", endpoint)
		}

		connection = Connection{endpoint: endpoint, conn: conn}
		c.connections[serviceName] = connection
	}

	return connection.conn, nil
}
