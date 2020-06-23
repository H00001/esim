package kafka

import (
	"github.com/Shopify/sarama"
)

type balanceStrategy struct {
	name   string
	coreFn func(plan sarama.BalanceStrategyPlan, memberIDs []string, topic string, partitions []int32)
}

// Name implements BalanceStrategy.
func (s *balanceStrategy) Name() string { return s.name }

// Balance implements BalanceStrategy.
func (s *balanceStrategy) Plan(members map[string]sarama.ConsumerGroupMemberMetadata, topics map[string][]int32) (sarama.BalanceStrategyPlan, error) {
	// Build members by topic map
	mbt := make(map[string][]string)
	for memberID, meta := range members {
		for _, topic := range meta.Topics {
			mbt[topic] = append(mbt[topic], memberID)
		}
	}

	// Assemble plan
	plan := make(sarama.BalanceStrategyPlan, len(members))
	for topic, memberIDs := range mbt {
		s.coreFn(plan, memberIDs, topic, topics[topic])
	}
	return plan, nil
}
