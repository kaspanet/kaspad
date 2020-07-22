package integration

import (
	"fmt"
	"testing"
)

func TestIntegrationBasicSync(t *testing.T) {
	kaspad1, kaspad2, client1, client2, teardown := setup(t)
	defer teardown()

	// TODO: DELETE THIS BEFORE MERGE
	fmt.Print(kaspad1, kaspad2, client1, client2)
}
