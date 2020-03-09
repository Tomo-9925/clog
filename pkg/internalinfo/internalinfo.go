package internalinfo

import (
	"github.com/xapima/conps/pkg/docker"
	"github.com/xapima/conps/pkg/ps"
	"github.com/xapima/conps/pkg/util"
)

type ContainerInternalApi struct {
	dapi *docker.DockerApi
}

type InternalInfo struct {
	Cid       string
	PasswdMap ps.PasswdMap
	GroupMap  ps.GroupMap
}

func NewContainerInternalApid() (*ContainerInternalApi, error) {
	c := ContainerInternalApi{}
	dapi, err := docker.NewDockerApi()
	if err != nil {
		return nil, util.ErrorWrapFunc(err)
	}
	c.dapi = dapi
	return &c, nil
}

func (c *ContainerInternalApi) GetInternalInfo(cid string) (InternalInfo, error) {
	ii := InternalInfo{Cid: cid}
	if err := c.dapi.SetCid(cid); err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	passwdMap, err := c.dapi.GetCidPasswdMap(cid)
	if err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	ii.PasswdMap = passwdMap

	groupMap, err := c.dapi.GetCidGroupMap(cid)
	if err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	ii.GroupMap = groupMap
	return ii, nil
}
