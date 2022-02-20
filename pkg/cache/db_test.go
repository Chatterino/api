package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
)

func TestDBCache(t *testing.T) {
	ctx := context.Background()
	const url = "user=pajlada host=/var/run/postgresql dbname=chatterino-api sslmode=disable"
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	newTime := time.Now().Add(5 * time.Hour)
	oldTime := time.Now().Add(-5 * time.Second)

	if _, err := conn.Exec(ctx, "DELETE FROM tooltips;"); err != nil {
		fmt.Println(err)
		t.Errorf("Unexpected error %v", err)
	}

	fmt.Println("a")

	inputRows := [][]interface{}{}

	for i := 0; i < 50000; i++ {
		inputRows = append(inputRows, []interface{}{fmt.Sprintf("newkey%d", i), "", newTime})
	}
	for i := 0; i < 50000; i++ {
		inputRows = append(inputRows, []interface{}{fmt.Sprintf("oldkey%d", i), "", oldTime})
	}

	copyCount, err := conn.CopyFrom(ctx, pgx.Identifier{"tooltips"}, []string{"url", "tooltip", "cached_until"}, pgx.CopyFromRows(inputRows))
	if err != nil {
		t.Errorf("Unexpected error from xd %v", err)
	}

	conn.Exec(ctx, "ANALYZE tooltips;")

	before := time.Now()
	conn.Exec(ctx, "DELETE FROM tooltips WHERE now() > cached_until;")
	after := time.Now()

	fmt.Println(before)
	fmt.Println(after)
	fmt.Println(after.Sub(before))

	fmt.Println(copyCount)
}
