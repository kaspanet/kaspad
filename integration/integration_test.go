package integration

import (
	"testing"
	"time"
)

func TestIntegrationBasicSync(t *testing.T) {
	kaspad1, kaspad2, client1, client2, teardown := setup(t)
	defer teardown()
	<-time.After(1 * time.Second)

	connect(t, kaspad1, kaspad2, client1, client2)

}
