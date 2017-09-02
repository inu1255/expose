package expose

import (
	"log"
	"testing"
)

func TestOnline(t *testing.T) {
	client := NewClientServer("59a7c165d929ceb31857ffed")
	err := client.Run()
	log.Println(err)
}
