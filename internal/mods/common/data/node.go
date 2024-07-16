package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/cachex"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const (
	NodeNs                = "node_ns"
	prefix                = "/node"
	nodeInfoKey           = "node_info_m4"
	nodeStateKey          = "node_state:%s"
	nodeStatusKeepLength  = 200
	nodeStatusQueryLength = 120
	nodeStatusKeepTimes   = 86400 //  1day in seconds
)

type nodeRepository struct {
	etcdClient  *clientv3.Client
	redisClient cachex.Cache
}

func NewNodeRepository(etcdClient *clientv3.Client, cache cachex.Cache) repo.NodeRepository {
	return &nodeRepository{
		etcdClient:  etcdClient,
		redisClient: cache,
	}
}

// GetAllNodeInfo get all node info
func (n *nodeRepository) GetAllNodeInfo(ctx context.Context) ([]*repo.Node, error) {
	response, err := n.etcdClient.Get(ctx, fmt.Sprintf("%s/%s", prefix, nodeInfoKey), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	nodeInfos := make([]*repo.Node, 0)
	for _, kv := range response.Kvs {
		var info repo.Node
		err = json.Unmarshal(kv.Value, &info)
		if err == nil {
			nodeInfos = append(nodeInfos, &info)
		}
	}
	return nodeInfos, nil
}

// UpdateNodeInfoList update node info list
func (n *nodeRepository) UpdateNodeInfoList(ctx context.Context, infos [][2]*repo.Node) error {
	conditions := make([]clientv3.Cmp, 0)
	dos := make([]clientv3.Op, 0)
	for _, infoPair := range infos {
		key := fmt.Sprintf("%s/%s/%s", prefix, nodeInfoKey, infoPair[1].Addr)
		if infoPair[0] == nil {
			conditions = append(conditions, clientv3.Compare(clientv3.Version(key), "=", 0))
		} else {
			oldInfoValue, err := json.Marshal(infoPair[0])
			if err != nil {
				return fmt.Errorf("failed to marshal NodeInfo: %v", err)
			}
			oldInfo := string(oldInfoValue)
			conditions = append(conditions, clientv3.Compare(clientv3.Value(key), "=", oldInfo))
		}
		newNodeValue, err := json.Marshal(infoPair[1])
		if err != nil {
			return fmt.Errorf("failed to marshal NodeInfo: %v", err)
		}
		dos = append(dos, clientv3.OpPut(key, string(newNodeValue)))
	}
	txn := n.etcdClient.Txn(ctx).If(conditions...).Then(dos...)
	resp, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	if !resp.Succeeded {
		return fmt.Errorf("transaction failed: %v", resp.Responses)
	}
	return nil
}

// UpdateNodeCostNum update node cost num
func (n *nodeRepository) UpdateNodeCostNum(ctx context.Context, addr string, stressType, num int) error {
	key := fmt.Sprintf("%s/%s/%s", prefix, nodeInfoKey, addr)

	// Get the current node information
	getResp, err := n.etcdClient.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get current node info: %v", err)
	}
	if len(getResp.Kvs) == 0 {
		return fmt.Errorf("node info not found for address: %s", addr)
	}

	// Unmarshal the current node info
	var currentInfo repo.Node
	err = json.Unmarshal(getResp.Kvs[0].Value, &currentInfo)
	if err != nil {
		return fmt.Errorf("failed to unmarshal current node info: %v", err)
	}

	// Create the updated node info
	updatedInfo := currentInfo
	if stressType == int(enums.Rps) {
		updatedInfo.RpsCost -= num
	} else {
		updatedInfo.GoroutineCost -= num
	}
	updatedInfo.RunningTaskCount -= 1

	// Marshal the updated node info
	updatedInfoValue, err := json.Marshal(updatedInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated node info: %v", err)
	}

	// Marshal the current node info to compare
	currentInfoValue, err := json.Marshal(currentInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal current node info: %v", err)
	}

	// Create a transaction with conditions
	txn := n.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(key), "=", string(currentInfoValue))).
		Then(clientv3.OpPut(key, string(updatedInfoValue)))
	// Commit the transaction
	txnResp, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	if !txnResp.Succeeded {
		return fmt.Errorf("transaction failed: concurrent update detected")
	}
	return nil
}

// DeleteNodeInfo delete node info
func (n *nodeRepository) DeleteNodeInfo(ctx context.Context, addr string) error {
	_, err := n.etcdClient.Delete(ctx, fmt.Sprintf("%s/%s/%s", prefix, nodeInfoKey, addr))
	return err
}

// UpdateNodeQuota updates the quota for a specific node.
//
// ctx: the context in which the update operation is performed.
// addr: the address of the node.
// rpsQuota: the new RPS quota for the node.
// goroutineQuota: the new goroutine quota for the node.
// error: returns an error if the update operation fails.
// UpdateNodeQuota updates the quota for a specific node.
func (n *nodeRepository) UpdateNodeQuota(ctx context.Context, addr string, rpsQuota, goroutineQuota int) error {
	key := fmt.Sprintf("%s/%s/%s", prefix, nodeInfoKey, addr)

	// Get the current node information
	getResp, err := n.etcdClient.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get current node info: %v", err)
	}

	var currentInfo repo.Node
	var updatedInfo repo.Node

	if len(getResp.Kvs) == 0 {
		// If no existing node info, initialize both currentInfo and updatedInfo with default values
		updatedInfo = repo.Node{
			Addr:             addr,
			RpsQuota:         rpsQuota,
			GoroutineQuota:   goroutineQuota,
			RpsCost:          0,
			GoroutineCost:    0,
			RunningTaskCount: 0,
		}
	} else {
		// Unmarshal the current node info
		err = json.Unmarshal(getResp.Kvs[0].Value, &currentInfo)
		if err != nil {
			return fmt.Errorf("failed to unmarshal current node info: %v", err)
		}

		// Create the updated node info with new quotas
		updatedInfo = currentInfo
		updatedInfo.RpsQuota = rpsQuota
		updatedInfo.GoroutineQuota = goroutineQuota
	}

	// Marshal the updated node info
	updatedInfoValue, err := json.Marshal(updatedInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated node info: %v", err)
	}

	if len(getResp.Kvs) == 0 {
		// If no existing node info, directly put the new value
		_, err = n.etcdClient.Put(ctx, key, string(updatedInfoValue))
		if err != nil {
			return fmt.Errorf("failed to put new node info: %v", err)
		}
	} else {
		// Marshal the current node info to compare
		currentInfoValue, err := json.Marshal(currentInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal current node info: %v", err)
		}

		// Create a transaction with conditions
		txn := n.etcdClient.Txn(ctx).
			If(clientv3.Compare(clientv3.Value(key), "=", string(currentInfoValue))).
			Then(clientv3.OpPut(key, string(updatedInfoValue)))

		// Commit the transaction
		txnResp, err := txn.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}
		if !txnResp.Succeeded {
			return fmt.Errorf("transaction failed: concurrent update detected")
		}
	}
	return nil
}

// ReportNodeInfo report node info
func (n *nodeRepository) ReportNodeInfo(ctx context.Context, state *repo.NodeState) error {
	key := fmt.Sprintf(nodeStateKey, state.Addr)

	// Serialize the node state to JSON
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal NodeState: %v", err)
	}

	// Create a pipeline to execute multiple commands atomically
	pipe := n.redisClient.TxPipeline()

	// Push the new data to the Redis list
	pipe.LPush(ctx, key, data)

	// Trim the list to keep only the latest 200 entries
	pipe.LTrim(ctx, key, 0, nodeStatusKeepLength-1)

	// Set the TTL to 1 day
	pipe.Expire(ctx, key, nodeStatusKeepTimes*time.Second)

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %v", err)
	}
	return nil
}

// BatchGetNodeState batch get node state
func (n *nodeRepository) BatchGetNodeState(ctx context.Context, addrs []string) ([]*repo.NodeState, error) {
	pipe := n.redisClient.Pipeline()
	results := make([]*redis.StringSliceCmd, len(addrs))
	keys := make([]string, len(addrs))

	// Create the pipeline commands
	for i, addr := range addrs {
		key := fmt.Sprintf(nodeStateKey, addr)
		results[i] = pipe.LRange(ctx, key, 0, nodeStatusQueryLength-1)
		keys[i] = key
	}

	// Execute the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Redis pipeline: %v", err)
	}

	// Collect results
	nodeStates := make([]*repo.NodeState, 0)
	for i, cmd := range results {
		if err := cmd.Err(); err != nil {
			return nil, fmt.Errorf("failed to get node state for key %s: %v", keys[i], err)
		}
		values, err := cmd.Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get result for key %s: %v", keys[i], err)
		}
		for _, value := range values {
			var state repo.NodeState
			err = json.Unmarshal([]byte(value), &state)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal node state for key %s: %v", keys[i], err)
			}
			nodeStates = append(nodeStates, &state)
		}
	}

	return nodeStates, nil
}

// DeleteNodeStateInfo deletes node state information.
//
// ctx: The context in which the function is being called.
// addr: The address of the node state information to be deleted.
// error: An error, if any, that occurred during the deletion process.
func (n *nodeRepository) DeleteNodeStateInfo(ctx context.Context, addr string) error {
	n.redisClient.Delete(ctx, NodeNs, fmt.Sprintf(nodeStateKey, addr))
	return nil
}
