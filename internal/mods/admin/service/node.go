package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type NodeService struct {
	uc *biz.NodeUsecase
}

func NewNodeService(uc *biz.NodeUsecase) *NodeService {
	return &NodeService{uc: uc}
}

// QueryAllNodes retrieves all nodes based on the query parameters.
//
// Takes in a gin Context as a parameter and does not return anything.
func (s *NodeService) QueryAllNodes(c *gin.Context) {
	nodes, err := s.uc.QueryAllNodes(c.Request.Context())
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, nodes)
}

// QueryNodes is a function to query nodes.
//
// It takes a gin Context as a parameter and does not return any value.
func (s *NodeService) QueryNodes(c *gin.Context) {
	var params biz.NodeQuery
	if err := common.ParseQuery(c, &params); err != nil {
		common.ResError(c, err)
		return
	}
	nodes, err := s.uc.QueryNodes(c.Request.Context(), &params)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, nodes)
}

// UpdateNodeQuota updates the quota for a node.
//
// context *gin.Context
func (s *NodeService) UpdateNodeQuota(c *gin.Context) {
	item := new(biz.UpdateNodeForm)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	}
	result, err := s.uc.UpdateNodeQuota(c.Request.Context(), item)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, result)
}
