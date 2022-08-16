package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func setOnBlockAddedHandler(t *testing.T, harness *appHarness, handler func(notification *appmessage.BlockAddedNotificationMessage)) {
	err := harness.rpcClient.RegisterForBlockAddedNotifications(handler)
	if err != nil {
		t.Fatalf("Error from RegisterForBlockAddedNotifications: %s", err)
	}
}

func TestNotificationIDs(t *testing.T) {
	kaspad1, kaspad2, kaspad3, teardown := standardSetupWithUtxoindex(t)

	defer teardown()

	ID1 := "kaspad1"
	ID2 := "kaspad2"
	ID3 := "kaspad3"

	err := kaspad1.rpcClient.RegisterForBlockAddedNotificationsWithID(
		func(notification *appmessage.BlockAddedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForBlockAddedNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForBlockAddedNotificationsWithID(
		func(notification *appmessage.BlockAddedNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForBlockAddedNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForBlockAddedNotificationsWithID(
		func(notification *appmessage.BlockAddedNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForBlockAddedNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForVirtualSelectedParentBlueScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForUTXOsChangedNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForUTXOsChangedNotificationsWithID(
		[]string{kaspad1.miningAddress, kaspad2.miningAddress, kaspad3.miningAddress},
		func(notification *appmessage.UTXOsChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForUTXOsChangedNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForUTXOsChangedNotificationsWithID(
		[]string{kaspad1.miningAddress, kaspad2.miningAddress, kaspad3.miningAddress},
		func(notification *appmessage.UTXOsChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForUTXOsChangedNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForUTXOsChangedNotificationsWithID(
		[]string{kaspad1.miningAddress, kaspad2.miningAddress, kaspad3.miningAddress},
		func(notification *appmessage.UTXOsChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForUTXOsChangedNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForNewBlockTemplateNotificationsWithID(
		func(notification *appmessage.NewBlockTemplateNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForNewBlockTemplateNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForNewBlockTemplateNotificationsWithID(
		func(notification *appmessage.NewBlockTemplateNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForNewBlockTemplateNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForNewBlockTemplateNotificationsWithID(
		func(notification *appmessage.NewBlockTemplateNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForNewBlockTemplateNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForVirtualDaaScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualDaaScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualDaaScoreChangedNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForVirtualDaaScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualDaaScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualDaaScoreChangedNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForVirtualDaaScoreChangedNotificationsWithID(
		func(notification *appmessage.VirtualDaaScoreChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualDaaScoreChangedNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForVirtualSelectedParentChainChangedNotificationsWithID(
		false,
		func(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualSelectedParentChainChangedNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForVirtualSelectedParentChainChangedNotificationsWithID(
		false,
		func(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualSelectedParentChainChangedNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForVirtualSelectedParentChainChangedNotificationsWithID(
		false,
		func(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForVirtualSelectedParentChainChangedNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterForFinalityConflictsNotificationsWithID(
		func(notification *appmessage.FinalityConflictNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		func(notification *appmessage.FinalityConflictResolvedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForFinalityConflictsNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterForFinalityConflictsNotificationsWithID(
		func(notification *appmessage.FinalityConflictNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		func(notification *appmessage.FinalityConflictResolvedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForFinalityConflictsNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterForFinalityConflictsNotificationsWithID(
		func(notification *appmessage.FinalityConflictNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		func(notification *appmessage.FinalityConflictResolvedNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterForFinalityConflictsNotificationsWithID: %s", err)
	}

	err = kaspad1.rpcClient.RegisterPruningPointUTXOSetNotificationsWithID(
		func(notification *appmessage.PruningPointUTXOSetOverrideNotificationMessage) {
			checkIDs(t, notification.ID, ID1)
		},
		ID1,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterPruningPointUTXOSetNotificationsWithID: %s", err)
	}

	err = kaspad2.rpcClient.RegisterPruningPointUTXOSetNotificationsWithID(
		func(notification *appmessage.PruningPointUTXOSetOverrideNotificationMessage) {
			checkIDs(t, notification.ID, ID2)
		},
		ID2,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterPruningPointUTXOSetNotificationsWithID: %s", err)
	}

	err = kaspad3.rpcClient.RegisterPruningPointUTXOSetNotificationsWithID(
		func(notification *appmessage.PruningPointUTXOSetOverrideNotificationMessage) {
			checkIDs(t, notification.ID, ID3)
		},
		ID3,
	)
	if err != nil {
		t.Fatalf("Failed to register with RegisterPruningPointUTXOSetNotificationsWithID: %s", err)
	}

	const approxBlockAmountToMine = 100

	for i := 0; i < approxBlockAmountToMine/3; i++ {

		mineNextBlock(t, kaspad1)
		mineNextBlock(t, kaspad2)
		mineNextBlock(t, kaspad3)
	}
}

func checkIDs(t *testing.T, notificationID string, expectedID string) {
	if expectedID == "" {
		t.Fatalf("the kaspad with assigned id %s is using the default id %s - cannot test id assignment!", expectedID, "")
	}
	if notificationID != expectedID {
		t.Fatalf("the kaspad with assigned id %s got a notification with id %s", expectedID, notificationID)
	}
}
