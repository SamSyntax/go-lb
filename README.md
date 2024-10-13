Go Load Balancer with Dynamic Server Spawning
This Go project implements a simple load balancer that forwards requests to multiple backend servers. The project dynamically spawns HTTP servers and balances the incoming requests using round-robin scheduling.

Overview
The project consists of:

A server spawner that creates multiple backend servers listening on different ports.
A load balancer that receives incoming requests on a designated port and forwards them to one of the spawned servers using a round-robin mechanism.
Each server responds with a message indicating the port it is serving on.
Features
Dynamic server spawning with a given number of servers.
Load balancing using the round-robin technique.
Each server runs independently and responds with a simple message.
Basic reverse proxy functionality for forwarding requests to backend servers.
Project Structure
Files
main.go: The entry point of the application that spawns servers and starts the load balancer.
loadbalancer.go: Contains the logic for the load balancer and the backend server configurations.
Code Explanation
Server Spawner
The Spawner function spawns a specified number of backend HTTP servers. Each server is assigned a unique port and name. Servers are spawned concurrently in goroutines to avoid blocking the main thread.

```go
func Spawner(amt int) []LbServer {
    servers := make([]LbServer, 0, amt)
    for i := 0; i < amt; i++ {
        name := fmt.Sprintf("Server %v", i)
        srv := Server(":"+strconv.Itoa(8000+i), name)
        servers = append(servers, srv)
    }
    return servers
}
```
Server Initialization
Each server is initialized using the Server function. It creates an HTTP server that listens on the given port and serves a simple response message.

```go
func Server(port, name string) LbServer {
    srv := NewLbServer("http://localhost" + port)
    srv.name = name
    // Serve HTTP requests in a goroutine
    go func() {
        http.ListenAndServe(port, mux)
    }()
    return *srv
}
```
Load Balancer
The LoadBalancer struct holds a list of backend servers and distributes incoming requests among them using a round-robin algorithm.

```go
func (lb *LoadBalancer) GetNextAvailableServer() LbServer {
    server := lb.servers[lb.roundRobinCount % len(lb.servers)]
    lb.roundRobinCount++
    return server
}
```
The load balancer listens on a specified port (e.g., 7000), and forwards requests to the next available backend server.

```go
func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
    targetServer := lb.GetNextAvailableServer()
    fmt.Printf("forwarding to %q\n", targetServer.Address())
    targetServer.Serve(w, r)
}
```
Running the Code
Clone this repository or copy the source files into your Go workspace.

Run the code using:

```bash
go build *.go -o ./lb && ./lb
```
OR

```bash
make build
```
The load balancer will listen on port 7000. Open your browser or use curl to send a request to localhost:7000:

```bash
curl http://localhost:7000
```
The request will be forwarded to one of the backend servers, and you will see a response indicating the server's port.

Example Output
```bash Serving requests at localhost:7000
Spawning server: Server 0 at localhost:8000
Spawning server: Server 1 at localhost:8001
Spawning server: Server 2 at localhost:8002
Spawning server: Server 3 at localhost:8003
Spawning server: Server 4 at localhost:8004
forwarding to "localhost:8001"
forwarding to "localhost:8002"
forwarding to "localhost:8003"
forwarding to "localhost:8004"
forwarding to "localhost:8000"
```
Customization
Number of Servers: You can modify the number of servers spawned by changing the argument passed to the Spawner function in main.go.
Ports: The backend servers listen on ports 8000 and higher. You can modify the port range in the Spawner function.
Dependencies
This project does not require any third-party dependencies. It only uses the Go standard library.

License
This project is licensed under the MIT License.
