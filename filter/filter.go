package filter

import (
	"raptor/executor"
	"raptor/proto"
)

var Filters = make(map[string]Filter)

type Filter interface {
	Filter(instance *proto.JobInstance) error
}

func NewChain(instance proto.JobInstance) *Chain {
	chain := &Chain{instance: &instance}
	for _, f := range instance.Config.PreFilter {
		chain.withPreFilter(f)
	}
	chain.withPreFilter(MetricsPreFilterKey)
	chain.withPostFilter(MetricsPostFilterKey)
	for _, f := range instance.Config.PostFilter {
		chain.withPreFilter(f)
	}
	chain.withExecutor(instance.Config.Executor)
	return chain
}

type Chain struct {
	preFilter []Filter
	postFilter []Filter
	executor executor.Executor
	instance *proto.JobInstance
}

func (c *Chain) Do() error {
	for _, f := range c.preFilter {
		if err := f.Filter(c.instance); err != nil {
			return err
		}
	}

	if err := c.executor.Execute(*c.instance); err != nil {
		return err
	}

	for _, f := range c.postFilter {
		if err := f.Filter(c.instance); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) withPreFilter(filter string) *Chain {
	c.preFilter = append(c.preFilter, Filters[filter])
	return c
}

func (c *Chain) withPostFilter(filter string) *Chain {
	c.postFilter = append(c.postFilter, Filters[filter])
	return c
}

func (c *Chain) withExecutor(exec string) *Chain {
	if exec == "" {
		exec = executor.HttpExecutorKey
	}
	c.executor = executor.Executors[exec]
	return c
}



