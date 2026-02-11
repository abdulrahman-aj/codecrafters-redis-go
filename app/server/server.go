package server

type request struct {
	command    []string
	responseCh chan<- string
}

type Server struct {
	requests chan *request
}

func New() *Server {
	return &Server{requests: make(chan *request)}
}

func (s *Server) Run() {
	for req := range s.requests {
		req.responseCh <- s.handle(req.command)
	}
}

func (s *Server) Do(command []string) string {
	ch := make(chan string)
	req := &request{
		command:    command,
		responseCh: ch,
	}

	s.requests <- req
	return <-ch
}

func (s *Server) handle(command []string) string {
	return "+PONG\r\n"
}
