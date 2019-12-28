package monitor

import (
    log "github.com/sirupsen/logrus"
    "github.com/sobitada/thor/pooltool"
    "time"
)

type ActionContext struct {
    BlockHeightMap     map[string]uint32
    MaximumBlockHeight uint32
    UpToDateNodes      []string
}

type Action interface {
    execute(nodes []Node, context ActionContext)
}

type ShutDownWithBlockLagAction struct{}

func (action ShutDownWithBlockLagAction) execute(nodes []Node, context ActionContext) {
    log.Infof("Maximum last block height '%v' reported by %v.", context.MaximumBlockHeight, context.UpToDateNodes)
    for p := range nodes {
        peer := nodes[p]
        peerBlockHeight, found := context.BlockHeightMap[peer.Name]
        if found {
            if peerBlockHeight < (context.MaximumBlockHeight - peer.MaxBlockLag) {
                log.Warnf("[%s] Pool has fallen behind %v blocks.", peer.Name, context.MaximumBlockHeight-peerBlockHeight)
                go shutDownNode(peer)
            }
        }
    }
}

// shuts down the peer
func shutDownNode(node Node) {
    _ = node.API.Shutdown()
    time.Sleep(time.Duration(200) * time.Millisecond)
    _ = node.API.Shutdown()
}

type PostLastTipToPoolToolAction struct {
    PoolID      string
    UserID      string
    GenesisHash string
}

func (action PostLastTipToPoolToolAction) execute(nodes []Node, context ActionContext) {
    go pooltool.PostLatestTip(context.MaximumBlockHeight, action.PoolID, action.UserID, action.GenesisHash)
}
