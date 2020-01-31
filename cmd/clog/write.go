package main

import (
	"encoding/json"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/xapima/clog/pkg/internalinfo"
	"github.com/xapima/clog/pkg/logread"
	"github.com/xapima/conps/pkg/util"
)

var metainfoPath = "/var/log/clog/meta.log"
var eventLogPath = "/var/log/clog/event.log"

type CAuditClog struct {
	Cid   string
	Event logread.AuditLog
}

func writeInfos(cid string, iinfo internalinfo.InternalInfo, minfo types.ContainerJSON) error {
	f, err := os.OpenFile(metainfoPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	defer f.Close()
	iiJson, err := json.Marshal(iinfo)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	f.Write(iiJson)
	f.Write([]byte("\n"))
	miJson, err := json.Marshal(minfo)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	f.Write(miJson)
	f.Write([]byte("\n"))
	return nil
}

func writeAuditLog(cid string, auditlog logread.AuditLog) error {
	f, err := os.OpenFile(eventLogPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	defer f.Close()
	Clog := CAuditClog{Cid: cid, Event: auditlog}
	clogJson, err := json.Marshal(Clog)
	if err != nil {
		return util.ErrorWrapFunc(err)
	}
	f.Write(clogJson)
	f.Write([]byte("\n"))
	return nil

}
