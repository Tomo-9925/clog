package judge

import (
	"fmt"
)

// audit はプロセス起動のタイミングじゃないと順序がおかしくなる？

type containerInfo struct {
	isChecked   bool
	isContainer bool
	cid         string
	ppid        int
	children    map[int]struct{}
}
type containerCache map[int]*containerInfo

func (c containerCache) newContainerInfo() *containerInfo {
	info := containerInfo{isChecked: false}
	info.children = make(map[int]struct{})
	return &info
}

func (c containerCache) add(pid int, ppid int) error {
	if info, ok := c[pid]; ok {
		if info.isChecked {
			return nil
		}
	} else {
		c[pid] = c.newContainerInfo()
	}

	if info, ok := c[ppid]; ok {
		if info.isChecked {
			c[pid].isChecked = true
			c[pid].isContainer = c[ppid].isContainer
			c[pid].cid = c[ppid].cid
			c[pid].ppid = ppid
			c.addChild(ppid, pid)
			return nil
		}
	}
	// auditを上から順に辿っているので、親プロセスのほうが先に処理されているはず
	// ppidの情報がないのはおかしい
	return fmt.Errorf("unkown ppid: %v, pid: %v", ppid, pid)
}

func (c containerCache) addManually(pid int, isContainer bool, cid string, ppid int) {
	if _, ok := c[pid]; !ok {
		c[pid] = c.newContainerInfo()
	}
	c[pid].isChecked = true
	c[pid].isContainer = isContainer
	c[pid].cid = cid
	c[pid].ppid = ppid
	c.addChild(ppid, pid)
	// logrus.Infof("chache is : %v", getKeys(c))
}

func (c containerCache) addChild(ppid, pid int) {
	if _, ok := c[ppid]; !ok {
		c[ppid] = c.newContainerInfo()
	}
	c[ppid].children[pid] = struct{}{}
}

// func getKeys(m map[int]*containerInfo) []int {
// 	l := make([]int, 0, 100)
// 	for key, _ := range m {
// 		l = append(l, key)
// 	}
// 	sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
// 	return l
// }
