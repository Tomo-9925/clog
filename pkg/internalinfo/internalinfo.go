package internalinfo

import (
	"github.com/xapima/conps/pkg/docker"
	"github.com/xapima/conps/pkg/util"
)

type ContainerInternalApi struct {
	dapi *docker.DockerApi
}

type InternalInfo struct {
	CidPasswdMap *docker.CidPasswdMap
	CidGroupMap  *docker.CidGroupMap
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
	ii := InternalInfo{}
	if err := c.dapi.SetCid(cid); err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	cidPasswdMap, err := c.dapi.GetCidPasswdMap(cid)
	if err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	ii.CidPasswdMap = cidPasswdMap

	cidGroupMap, err := c.dapi.GetCidGroupMap(cid)
	if err != nil {
		return InternalInfo{}, util.ErrorWrapFunc(err)
	}
	ii.CidGroupMap = cidGroupMap
	return ii, nil
}
