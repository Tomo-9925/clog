package metainfo

import (
	"github.com/docker/docker/api/types"
	"github.com/xapima/conps/pkg/docker"
	"github.com/xapima/conps/pkg/util"
)

type MetainfoApi struct {
	dapi *docker.DockerApi
}

func NewMetainfoApi() (*MetainfoApi, error) {
	mapi := MetainfoApi{}

	dapi, err := docker.NewDockerApi()
	if err != nil {
		return nil, util.ErrorWrapFunc(err)
	}
	mapi.dapi = dapi
	return &mapi, nil
}

// GetMetadata get container metadate from docker daemon. It is not cached, so cache it in the calling function.
func (mapi *MetainfoApi) GetMetadata(cid string) (types.ContainerJSON, error) {
	cJson, err := mapi.dapi.GetInspectWithCid(cid)
	if err != nil {
		return types.ContainerJSON{}, util.ErrorWrapFunc(err)
	}
	return cJson, nil
}
