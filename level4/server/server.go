package server

import (
	"fmt"
    "github.com/coreos/raft"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"bytes"
	"encoding/json"
	"net/http"
	"stripe-ctf.com/sqlcluster/log"
	"stripe-ctf.com/sqlcluster/sql"
	"stripe-ctf.com/sqlcluster/transport"
	"stripe-ctf.com/sqlcluster/util"
	"time"
    "strings"
)

type Server struct {
	name       string
	path       string
	listen     string
	router     *mux.Router
	httpServer *http.Server
	sql        *sql.SQL
	client     *transport.Client
	cluster    *Cluster
	raftServer raft.Server
}

// Creates a new server.
func New(path, listen string) (*Server, error) {
    rand.Seed(time.Now().UTC().UnixNano())
	cs, err := transport.Encode(listen)
	if err != nil {
		return nil, err
	}

    sqlPath := ":memory:"
	s := &Server{
		name:    listen,
		path:    path,
		listen:  listen,
		sql:     sql.NewSQL(sqlPath),
		router:  mux.NewRouter(),
		client:  transport.NewClient(),
		cluster: NewCluster(path, cs),
	}
	return s, nil
}

// Returns the connection string.
func (s *Server) connectionString() string {
	return _connectionString(s.listen)
}

func _connectionString(listen string) string {
	cStr, _ := transport.Encode(listen)
	return cStr
}

// Starts the server.
func (s *Server) ListenAndServe(leader string) error {
	var err error

	// Initialize and start Raft server.
	log.Printf("Initializing Raft Server: %s", s.path)

	transporter := raft.NewHTTPTransporter("/raft")
	transporter.Transport.Dial = transport.UnixDialer

	s.raftServer, err = raft.NewServer(s.name, s.path, transporter, nil, s.sql, "")
	if err != nil {
		log.Fatal(err)
	}
	s.raftServer.SetElectionTimeout(250 * time.Millisecond)
	//s.raftServer.SetHeartbeatInterval(50 * time.Millisecond)
	//raft.SetLogLevel(raft.Debug)

	transporter.Install(s.raftServer, s)
	s.raftServer.Start()

	if leader != "" {
		// Join to leader if specified.

		log.Println("Attempting to join leader:", leader)

		if !s.raftServer.IsLogEmpty() {
			log.Fatal("Cannot join with an existing log")
		}
        for {
		if err := s.Join(leader); err != nil {
			log.Printf("Unable to join cluster: %s", err)
            time.Sleep(100 * time.Millisecond)
			//log.Fatal(err)
            continue
		}
        break
        }
	} else if s.raftServer.IsLogEmpty() {
		// Initialize the server by joining itself.

		log.Printf("Initializing new cluster with name %s\n", s.raftServer.Name())

		_, err := s.raftServer.Do(&raft.DefaultJoinCommand{
			Name:             s.raftServer.Name(),
			ConnectionString: s.connectionString(),
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Println("Recovered from log")
	}

	// Initialize and start HTTP server.
	s.httpServer = &http.Server{
		Handler: s.router,
	}

	s.router.HandleFunc("/sql", s.sqlHandler).Methods("POST")
	s.router.HandleFunc("/join", s.joinHandler).Methods("POST")

	// Start Unix transport
	l, err := transport.Listen(s.listen)
	if err != nil {
		log.Fatal(err)
	}
	//log.Debugf("[%s] state: %s leader: %s peers: %#v", s.listen, s.raftServer.State(), s.raftServer.Leader(), s.raftServer.MemberCount())
	return s.httpServer.Serve(l)
}

// Client operations

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

// Join an existing cluster
func (s *Server) Join(leader string) error {

	command := &raft.DefaultJoinCommand{
		Name:             s.raftServer.Name(),
		ConnectionString: s.connectionString(),
	}

	b := util.JSONEncode(command)
	cs, err := transport.Encode(leader)
	if err != nil {
		return err
	}
	_, err = s.client.SafePost(cs, "/join", b)

	if err != nil {
		return err
	}

	return nil
}

// Server handlers
func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
	command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := s.raftServer.Do(command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// This is the only user-facing function, and accordingly the body is
// a raw string rather than JSON.
func (s *Server) sqlHandler(w http.ResponseWriter, req *http.Request) {

	leader := s.raftServer.Leader()
	if leader == "" {
		http.Error(w, "no leader yet", http.StatusBadRequest)
		return
	}
	query, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	outputInterface, err := s.raftServer.Do(sql.NewSqlCommand(string(query), s.name))
	if err != nil {
		switch err {
		case raft.NotLeaderError:
            q := string(query)
            if !strings.Contains(q, "/*") {
                nonce := fmt.Sprintf("%07x", rand.Int())[0:7]
                q = string(query) + "/*" +  nonce + "*/"
            }
			s.fwdToLeader(w, req, []byte(q))
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	output := outputInterface.(*sql.Output)
	if err != nil {
		msg := err.Error()
		w.Write([]byte(msg))
	} else {
		formatted := fmt.Sprintf("SequenceNumber: %d\n%s",
			output.SequenceNumber, output.Stdout)

		w.Write([]byte(formatted))
	}
}

func (s *Server) fwdToLeader(w http.ResponseWriter, req *http.Request, data []byte) {
	leader := s.raftServer.Leader()
    if leader == "" {
        // XXX: tune this
        time.Sleep(10 * time.Millisecond)
        s.fwdToLeader(w, req, data)
        return
    }
	cs, err := transport.Encode(leader)
	reader, err := s.client.SafePost(cs, "/sql", bytes.NewReader([]byte(data)))

	if err != nil {
        time.Sleep(10 * time.Millisecond)
        s.fwdToLeader(w, req, data)
	} else if reader == nil {
        time.Sleep(10 * time.Millisecond)
        s.fwdToLeader(w, req, data)
	} else {
		b := reader.(*bytes.Buffer)
		w.Write(b.Bytes())
	}
}
