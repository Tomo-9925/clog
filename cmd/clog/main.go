package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xapima/clog/pkg/internalinfo"
	"github.com/xapima/clog/pkg/judge"
	"github.com/xapima/clog/pkg/logread"
	"github.com/xapima/clog/pkg/metainfo"
	"github.com/xapima/clog/pkg/runnotify"
	"github.com/xapima/conps/pkg/ps"
)

var auditLogPath = "/var/log/audit/audit.log"
var debugContainerPid = -1

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
}

func main() {

	iapi, err := internalinfo.NewContainerInternalApid()
	if err != nil {
		logrus.WithField("cant create new ContainerInternalApi", err)
	}
	japi := judge.NewJudgeApi()
	auditCh := make(chan logread.AuditLog, 100)
	auditErrCh := make(chan error)
	aapi := logread.NewAuditApi(auditCh, auditErrCh)
	mapi, err := metainfo.NewMetainfoApi()
	if err != nil {
		logrus.WithField("cant create new Metainfo Api", err)
	}
	runCh := make(chan string)
	killCh := make(chan string)
	runErrCh := make(chan error)
	rapi, err := runnotify.NewRunnotifyApi(runCh, killCh, runErrCh)
	if err != nil {
		logrus.Errorf("cant create new RunnotifyApi: %v", err)
	}
	nowTime := time.Now()
	go rapi.Start()
	go aapi.ReadSince(auditLogPath, nowTime)
	// go aapi.Read(auditLogPath)
	for {
		select {
		case cid := <-runCh:
			logrus.Infof("RUN cid: %v", cid)
			minfo, err := mapi.GetMetadata(cid)
			if err != nil {
				logrus.Error(err)
			}
			pid := minfo.State.Pid
			ppid, err := ps.PPid(pid)
			if err != nil {
				logrus.Errorf("cant get container init ppid: %v", err)
			}
			pppid, err := ps.PPid(ppid)
			if err != nil {
				logrus.Errorf("cant get container init pppid: %v", err)
			}
			japi.SetContainerProc(pid, true, cid, ppid)
			japi.SetContainerProc(ppid, true, cid, pppid)
			// logrus.Debug("init set pid %d, ppid %d", pid, ppid)
			// logrus.Debug("containerInfo of  pid %d: %v", pid, japi.GetContainerInfo(pid))
			debugContainerPid = pid
			logrus.Debug("GetInternalInfo cid:", cid)
			iinfo, err := iapi.GetInternalInfo(cid)
			if err != nil {
				logrus.Errorf("cant get internal info: pid: %v, %v", cid, err)
			}
			if err := writeInfos(iinfo, minfo); err != nil {
				logrus.Errorf("cant write  Infos: %v", err)
			}
		case cid := <-killCh:
			logrus.Infof("kill cid: %v", cid)
			mdata, err := mapi.GetMetadata(cid)
			if err != nil {
				logrus.Error(err)
			}
			pid := mdata.State.Pid
			go japi.RemoveCid1mAfter(pid)

		case err := <-runErrCh:
			logrus.Errorf("runnotify error: %v", err)
		case auditlog := <-auditCh:
			// logrus.Debug("AuditLogCh EXE:", auditlog.Exe)
			// logrus.Debug("AuditLogCh PPid:", auditlog.PPid)
			ok, cid, err := japi.IsContainerProc(auditlog.Pid, auditlog.PPid)
			if err != nil {
				logrus.Debug("cant check isContainerProc: %v", err)
			}
			logrus.Debug("isContainerProc:", ok)
			// logrus.Infof("pid %d, ppid %d, commandline %v isContainerProc: %v", auditlog.Pid, auditlog.PPid, auditlog.Commandline, ok)
			if ok {
				if err := writeAuditLog(cid, auditlog); err != nil {
					logrus.Errorf("cant write audit log: %v", err)
				}
			}
		case err := <-auditErrCh:
			logrus.Errorf("logread error: %v", err)

		}

	}
}
