package main

import (
	"github.com/kaspanet/kaspad/faucet/database"
	"github.com/kaspanet/kaspad/httpserverutils"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"time"
)

const minRequestInterval = time.Hour * 24

type ipUse struct {
	IP      string
	LastUse time.Time
}

func ipFromRequest(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func validateIPUsage(r *http.Request) error {
	db, err := database.DB()
	if err != nil {
		return err
	}
	now := time.Now()
	timeBeforeMinRequestInterval := now.Add(-minRequestInterval)
	var count int
	ip, err := ipFromRequest(r)
	if err != nil {
		return err
	}
	dbResult := db.Model(&ipUse{}).Where(&ipUse{IP: ip}).Where("last_use BETWEEN ? AND ?", timeBeforeMinRequestInterval, now).Count(&count)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("Some errors were encountered when checking the last use of an IP:", dbResult.GetErrors())
	}
	if count != 0 {
		return httpserverutils.NewHandlerError(http.StatusForbidden, errors.New("A user is allowed to to have one request from the faucet every 24 hours"))
	}
	return nil
}

func updateIPUsage(r *http.Request) error {
	db, err := database.DB()
	if err != nil {
		return err
	}

	ip, err := ipFromRequest(r)
	if err != nil {
		return err
	}
	dbResult := db.Where(&ipUse{IP: ip}).Assign(&ipUse{LastUse: time.Now()}).FirstOrCreate(&ipUse{})
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewErrorFromDBErrors("Some errors were encountered when upserting the IP to the new date:", dbResult.GetErrors())
	}
	return nil
}
