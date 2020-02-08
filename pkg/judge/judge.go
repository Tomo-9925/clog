package judge

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xapima/conps/pkg/ps"
	"github.com/xapima/conps/pkg/util"
)

type Judge struct {
	cache containerCache
}

func NewJudgeApi() *Judge {
	j := Judge{}
	j.cache = make(containerCache)
	j.setFirstProc()
	return &j
}

func (j *Judge) IsContainerProc(pid int, ppid int) (bool, string, error) {
	if info, ok := j.cache[pid]; ok && info.isChecked {
		return info.isContainer, info.cid, nil
	}
	if ok, cid, err := j.isContainerProcFromCgroup(pid, ppid); err == nil {
		return ok, cid, nil
	}
	if err := j.cache.add(pid, ppid); err != nil {
		return false, "", util.ErrorWrapFunc(err)
	}
	logrus.Debug("cache add pid:", pid)
	return j.cache[pid].isContainer, j.cache[pid].cid, nil
}

func (j *Judge) isContainerProcFromCgroup(pid, ppid int) (bool, string, error) {
	logrus.Debug("FromCgroup pid:", pid)
	pidnamespace, err := ps.GetPidNameSpace(proc, pid)
	if err == ps.PidNameSpaceNotFoundError {
		j.cache.addManually(pid, false, "", ppid)
		logrus.Debug("PidNameSpaceNotFoundError")
		return false, "", nil
	} else if err != nil {
		return false, "", util.ErrorWrapFunc(err)
	}
	logrus.Debug("PidNameSpace:", pidnamespace)
	dir, file := filepath.Split(pidnamespace)
	if !strings.HasPrefix(dir, cgroupContainerPrefix) {
		logrus.Debug("Dont have container prefix")
		j.cache.addManually(pid, false, "", ppid)
		return false, "", nil
	}
	j.cache.addManually(pid, true, file, ppid)
	logrus.Debugf("Pid %d is DockerProcess", pid)
	return true, file, nil
}

func (j *Judge) SetContainerProc(pid int, isContainer bool, cid string, ppid int) {
	j.cache.addManually(pid, isContainer, cid, ppid)
}

// RemoveContainerProc remove arg pid and all children pids.You should call when remove pod.
func (j *Judge) RemoveContainerProc(pid int) error {
	if _, ok := j.cache[pid]; !ok {
		return util.ErrorWrapFunc(fmt.Errorf("unknown pid: %v", pid))
	}
	if len(j.cache[pid].children) != 0 {
		for cpid := range j.cache[pid].children {
			if err := j.RemoveContainerProc(cpid); err != nil {
				return util.ErrorWrapFunc(err)
			}
		}
	}
	delete(j.cache, pid)
	return nil
}

func (j *Judge) RemoveCid1mAfter(pid int) {
	time.Sleep(time.Minute * 1)
	err := j.RemoveContainerProc(pid)
	logrus.Errorf("cant remove containerCache: pid: %v: %v", pid, err)
}

func (j *Judge) setFirstProc() {
	j.SetContainerProc(0, false, "", 0)

	fileinfos, err := ioutil.ReadDir(proc)
	if err != nil {
		logrus.Error(util.ErrorWrapFunc(err))
	}
	for _, fi := range fileinfos {
		if isNum(fi.Name()) {
			pid, err := strconv.Atoi(fi.Name())
			if err != nil {
				logrus.Error(util.ErrorWrapFunc(err))
			}
			ppid, err := ps.PPid(pid)
			if err != nil {
				logrus.Errorf("in first: cant get ppid: pid %d: %v", pid, err)
			}
			ok, cid, err := j.IsContainerProc(pid, ppid)
			if err != nil {
				logrus.Errorf("in the first: cant check isContainr: pid %d: %v", pid, err)
			}
			j.cache.addManually(pid, ok, cid, ppid)
		}
	}
}

type ContainerInfo struct {
	IsChecked   bool
	IsContainer bool
	Cid         string
	PPid        int
	// Children    map[int]struct{}
}

func (j *Judge) GetContainerInfo(pid int) ContainerInfo {
	info := ContainerInfo{}
	info.IsChecked = j.cache[pid].isChecked
	info.IsContainer = j.cache[pid].isContainer
	info.Cid = j.cache[pid].cid
	info.PPid = j.cache[pid].ppid
	// info.Children = j.cache[pid].children
	return info
}
