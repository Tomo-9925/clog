package logread

const (
	AddSyscallFlag = uint(1)
	AddExecFlag    = uint(1) << iota
	AddCwdFlag
)

var (
	logpath = "/var/log/audit/audit.log"
)
