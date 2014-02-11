package sql

import (
//    "github.com/goraft/raft"
    "github.com/coreos/raft"
    "stripe-ctf.com/sqlcluster/log"
)

type SqlCommand struct {
    Tag   string `json:"tag"`
    Query string `json:"query"`
}

func init() {
    raft.RegisterCommand(&SqlCommand{})
}

func NewSqlCommand(query, tag string) *SqlCommand {
    return &SqlCommand{
        Query: query,
        Tag:   tag,
    }
}

func (c *SqlCommand) CommandName() string {
    return "Sql"
}

func (c *SqlCommand) Apply(server raft.Server) (interface{}, error) {
    sql_db := server.Context().(*SQL)
    output, err := sql_db.Execute(c.Query)

    if err != nil {
        log.Printf("Error in SQL query: %s\n", err.Error())
    }

    return output, nil
}
