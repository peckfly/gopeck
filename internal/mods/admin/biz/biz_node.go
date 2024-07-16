package biz

import (
	"context"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/consts"
	"github.com/peckfly/gopeck/pkg/registry"
	"github.com/spf13/cast"
)

type NodeUsecase struct {
	nodeRepository repo.NodeRepository
	discovery      registry.Discovery
}

// NewNodeUsecase creates a new NodeUsecase.
//
// Parameters:
//
//	nodeRepository: repo.NodeRepository
//	discovery: registry.Discovery
//
// Return type: *NodeUsecase
func NewNodeUsecase(nodeRepository repo.NodeRepository, discovery registry.Discovery) *NodeUsecase {
	return &NodeUsecase{
		nodeRepository: nodeRepository,
		discovery:      discovery,
	}
}

// QueryAllNodes retrieves all nodes using the provided context.
// It returns a NodeListResult pointer and an error.
// return rps、goroutine cost/total, status and node executing task count
func (s *NodeUsecase) QueryAllNodes(ctx context.Context) (*NodeListResult, error) {
	services, err := s.discovery.GetService(ctx, consts.Pecker)
	if err != nil {
		return nil, err
	}
	nodeInfos, err := s.nodeRepository.GetAllNodeInfo(ctx)
	if err != nil {
		return nil, err
	}
	var nodeItems []*NodeResultItem
	for _, service := range services {
		nodeInfo := &repo.Node{
			RpsCost:          0,
			GoroutineCost:    0,
			RunningTaskCount: 0,
			GoroutineQuota:   cast.ToInt(service.Metadata[consts.MaxConcurrencyNum]),
			RpsQuota:         cast.ToInt(service.Metadata[consts.MaxRpsNum]),
		}
		for _, info := range nodeInfos {
			if service.Addr == info.Addr {
				nodeInfo = info
				break
			}
		}
		nodeItems = append(nodeItems, &NodeResultItem{
			RpsCost:          nodeInfo.RpsCost,
			GoroutineCost:    nodeInfo.GoroutineCost,
			RunningTaskCount: nodeInfo.RunningTaskCount,
			Addr:             service.Addr,
			GoroutineQuota:   nodeInfo.GoroutineQuota,
			RpsQuota:         nodeInfo.RpsQuota,
		})
	}
	return &NodeListResult{nodeItems}, nil
}

// QueryNodes retrieves all nodes using the provided context.
// return node cpu/mem in graph
func (s *NodeUsecase) QueryNodes(ctx context.Context, nodeQuery *NodeQuery) (*NodeQueryResult, error) {
	nodeStates, err := s.nodeRepository.BatchGetNodeState(ctx, []string{nodeQuery.Addr})
	if err != nil {
		return nil, err
	}
	// reverse the order
	for i, j := 0, len(nodeStates)-1; i < j; i, j = i+1, j-1 {
		nodeStates[i], nodeStates[j] = nodeStates[j], nodeStates[i]
	}
	return &NodeQueryResult{nodeStates}, nil
}

// UpdateNodeQuota updates the quota for a node.
//
// ctx: the context for the operation.
// item: the form containing the updated node information.
// Returns two empty interfaces.
func (s *NodeUsecase) UpdateNodeQuota(ctx context.Context, updateNodeForm *UpdateNodeForm) (any, error) {
	err := s.nodeRepository.UpdateNodeQuota(ctx, updateNodeForm.Addr, updateNodeForm.RpsQuota, updateNodeForm.GoroutineQuota)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
