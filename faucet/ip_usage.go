package main

import (
	"github.com/daglabs/btcd/faucet/database"
	"github.com/daglabs/btcd/httpserverutils"
	"net"
	"net/http"
	"time"
)

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

func validateIPUsage(r *http.Request) *httpserverutils.HandlerError {
	db, err := database.DB()
	if err != nil {
		return httpserverutils.NewInternalServerHandlerError(err.Error())
	}
	now := time.Now()
	timeBefore24Hours := now.Add(-time.Hour * 24)
	var count int
	ip, err := ipFromRequest(r)
	if err != nil {
		return httpserverutils.NewInternalServerHandlerError(err.Error())
	}
	dbResult := db.Model(&ipUse{}).Where(&ipUse{IP: ip}).Where("last_use BETWEEN ? AND ?", now, timeBefore24Hours).Count(&count)
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when checking the last use of an IP:", dbResult.GetErrors())
	}
	if count != 0 {
		return httpserverutils.NewHandlerError(http.StatusForbidden, "A user is allowed to to have one request from the faucet every 24 hours.")
	}
	return nil
}

func updateIPUsage(r *http.Request) *httpserverutils.HandlerError {
	db, err := database.DB()
	if err != nil {
		return httpserverutils.NewInternalServerHandlerError(err.Error())
	}

	dbTx := db.Begin()
	ip, err := ipFromRequest(r)
	if err != nil {
		return httpserverutils.NewInternalServerHandlerError(err.Error())
	}
	dbResult := db.Where(&ipUse{IP: ip}).Delete(&ipUse{})
	dbErrors := dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when checking the last use of an IP:", dbResult.GetErrors())
	}
	requestIPUse := &ipUse{IP: ip, LastUse: time.Now()}
	dbResult = dbTx.Save(requestIPUse)
	dbErrors = dbResult.GetErrors()
	if httpserverutils.HasDBError(dbErrors) {
		return httpserverutils.NewHandlerErrorFromDBErrors("Some errors were encountered when checking the last use of an IP:", dbResult.GetErrors())
	}
	dbTx.Commit()
	return nil
}
