package sql

import (
    //"bytes"
    "strings"
    "database/sql"
    _ "github.com/mattn/go-sqlite3" 
    "stripe-ctf.com/sqlcluster/log"
    "sync"
)

type SQL struct {
    path           string
    sequenceNumber int
    mutex          sync.Mutex
    conn           *sql.DB
    cmds           map[string]*Output
}

type Output struct {
    Stdout         string
    SequenceNumber int
}

func NewSQL(path string) *SQL {
    s := &SQL{
        path: path,
    }
    
    conn, err := sql.Open("sqlite3", path)
    if err != nil {
            log.Fatal(err)
    }
    s.conn = conn
    s.cmds = make(map[string]*Output)
    return s
}

func (s *SQL) Execute(query string) (*Output, error) {
    orig_query := query
    // TODO: make sure I can catch non-lock issuez
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if val, ok := s.cmds[query]; ok {
        return val, nil
    }

    defer func() { 
        s.sequenceNumber += 1 
    }()

    if strings.Contains(query, ";") {
        query2 := query
        query = ""
        for _, q := range strings.Split(query2, ";") {
            if strings.Contains(q, "SELECT") {
                query = q
                break
            }
            s.conn.Exec(q)
        }
    }
    
    if query == "" {
        output := &Output{
            Stdout:         "",
            SequenceNumber: s.sequenceNumber,
        }

        s.cmds[orig_query] = output
        return output, nil        
    }
    
    rows, err := s.conn.Query(query)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    out := ""
    cols, _ := rows.Columns()
    rawResult := make([][]byte, len(cols))
    dest := make([]interface{}, len(cols))
    for i, _ := range rawResult {
        dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
    }    
    
    for rows.Next() {
        err := rows.Scan(dest...)
        if err != nil {
            panic(err)
        }    
            
        line := ""
        for _, v := range rawResult {
            if line != "" { line += "|" }
            line += string(v)
        }
        line += "\n"
        out += line
    }
    rows.Close()
    
    output := &Output{
        Stdout:         out,
        SequenceNumber: s.sequenceNumber,
    }
        
    s.cmds[orig_query] = output
    return output, nil    
}
