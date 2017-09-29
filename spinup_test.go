package spinup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"database/sql"
	_ "github.com/herenow/go-crate"
)

func TestSpinUpCrateDB(t *testing.T) {
	c := NewContainer("crate", 500*time.Millisecond)
	err := c.SpinUp(context.Background())
	defer func() {
		err := c.SpinDown(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Log("Container IP:", c.IPAddress)
	db, err := sql.Open("crate",
		fmt.Sprintf("http://%s:4200/", c.IPAddress))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Retry for 16 seconds, crate usually takes 8 to start.
	seconds := time.Second * 16
	milliseconds := time.Millisecond * 200
	retry := ((time.Millisecond * 1000) / milliseconds) * seconds
	for retry > 0 {
		time.Sleep(milliseconds)

		err = db.Ping()
		if err == nil {
			t.Log("Connected to crate")
			return
		}

		retry -= 1
	}
	t.Fatal("Couldn't connect")
}
