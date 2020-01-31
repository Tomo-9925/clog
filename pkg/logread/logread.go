package logread

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xapima/conps/pkg/util"
)

type AuditApi struct {
	reader      *Reader
	auditCh     chan AuditLog
	auditErrCh  chan error
	logCh       chan string
	logErrCh    chan error
	nowAuditLog *AuditLog
	nowId       uint64
}

type AuditLog struct {
	Pid         int
	PPid        int
	Uid         int
	Gid         int
	Success     bool
	Syscall     int
	Tty         string
	Exe         string
	Commandline []string
	Cwd         string
	addFlag     uint
}

func NewAuditApi(auditCh chan AuditLog, errCh chan error) *AuditApi {
	a := AuditApi{auditCh: auditCh, auditErrCh: errCh}
	a.logCh = make(chan string, 100)
	a.logErrCh = make(chan error)
	a.reader = NewReader(a.logCh, a.logErrCh)
	a.nowAuditLogFlush()
	return &a
}

func (a *AuditApi) Read(auditLogPath string) {
	defer close(a.auditCh)
	defer close(a.auditErrCh)

	go a.reader.Read(auditLogPath)
	for {
		select {
		case logline := <-a.logCh:
			if err := a.parseAuditLog(logline); err != nil {
				a.auditErrCh <- errors.Wrap(err, "audit log parse error")
			}
		case err := <-a.logErrCh:
			a.auditErrCh <- errors.Wrap(err, "read logfile error")
		}
	}
}

func (a *AuditApi) parseAuditLog(logline string) error {
	tags := strings.Split(logline, " ")
	switch strings.Split(tags[0], "=")[1] {
	case "SYSCALL":
		for _, tag := range tags {
			switch tagval := strings.Split(tag, "="); tagval[0] {
			case "pid":
				pid, err := strconv.Atoi(tagval[1])
				if err != nil {
					return util.ErrorWrapFunc(err)
				}
				a.nowAuditLog.Pid = pid
			case "ppid":
				ppid, err := strconv.Atoi(tagval[1])
				if err != nil {
					return util.ErrorWrapFunc(err)
				}
				a.nowAuditLog.PPid = ppid
			case "syscall":
				syscall, err := strconv.Atoi(tagval[1])
				if err != nil {
					return util.ErrorWrapFunc(err)
				}
				a.nowAuditLog.Syscall = syscall
			case "uid":
				uid, err := strconv.Atoi(tagval[1])
				if err != nil {
					return util.ErrorWrapFunc(err)
				}
				a.nowAuditLog.Uid = uid
			case "gid":
				gid, err := strconv.Atoi(tagval[1])
				if err != nil {
					return util.ErrorWrapFunc(err)
				}
				a.nowAuditLog.Gid = gid
			case "success":
				success := false
				if tagval[1] == "yes" {
					success = true
				}
				a.nowAuditLog.Success = success
			case "tty":
				a.nowAuditLog.Tty = strings.Trim(tagval[1], "\"")
			case "exe":
				a.nowAuditLog.Exe = strings.Trim(tagval[1], "\"")
				// case "comm":
				// 	a.nowAuditLog.Comm = tagval[1]
			}
		}
		a.nowAuditLog.addFlag |= AddSyscallFlag
	case "EXECVE":
		if a.nowAuditLog.addFlag&AddSyscallFlag == 0 {
			break
		}
		for _, tag := range tags[2:] {
			arg := strings.Split(tag, "=")[1]
			a.nowAuditLog.Commandline = append(a.nowAuditLog.Commandline, strings.Trim(arg, "\""))
			// a.nowAuditLog.Commandline = append(a.nowAuditLog.Commandline, arg[1:len(arg)-1])
		}
		a.nowAuditLog.addFlag |= AddExecFlag
	case "CWD":
		if a.nowAuditLog.addFlag&AddSyscallFlag == 0 {
			break
		}
		cwd := strings.Split(tags[3], "=")[1]
		a.nowAuditLog.Cwd = strings.Trim(cwd, "\"")
		a.nowAuditLog.addFlag |= AddCwdFlag
	}

	if a.isAuditLogPaesed() {
		a.auditCh <- *a.nowAuditLog
		a.nowAuditLogFlush()
	}
	return nil
}

func (a *AuditApi) isAuditLogPaesed() bool {
	if a.nowAuditLog.addFlag == AddSyscallFlag|AddCwdFlag|AddExecFlag {
		return true
	}
	return false
}

func (a *AuditApi) nowAuditLogFlush() {
	a.nowAuditLog = &AuditLog{}
}

func (a *AuditApi) ReadSince(auditLogPath string, nowTime time.Time) {
	defer close(a.auditCh)
	defer close(a.auditErrCh)

	go a.reader.Read(auditLogPath)
	for {
		select {
		case logline := <-a.logCh:
			// logrus.Infof("logline: %v", logline)
			// ココでエラーが出ることが有る。これは、  0755 ouid=0 ogid=0 rdev=00:00 nametype=NORMAL のような、ログの一部分のみを読み取ってしまっている。
			// ToDo : 原因の調査と、対策。
			if !strings.Contains(logline, "msg=audit") {
				logrus.Errorf("PASS: LOGLINE: %v", logline)
				break
			}
			times := strings.Split(strings.Split(strings.Split(strings.Split(strings.Split(logline, " ")[1], "=")[1], "(")[1], ":")[0], ".")
			sec, err := strconv.Atoi(times[0])
			if err != nil {
				a.auditErrCh <- util.ErrorWrapFunc(err)
			}
			nsec, err := strconv.Atoi(times[1])
			if err != nil {
				a.auditErrCh <- util.ErrorWrapFunc(err)
			}
			logTime := time.Unix(int64(sec), int64(nsec))
			if !logTime.After(nowTime) {
				continue
			}

			if err := a.parseAuditLog(logline); err != nil {
				a.auditErrCh <- errors.Wrap(err, "audit log parse error")
			}
		case err := <-a.logErrCh:
			a.auditErrCh <- errors.Wrap(err, "read logfile error")
		}
	}
}