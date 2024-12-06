/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package host implements the host pattern for multi-agent system.
package host

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// MultiAgent is a host multi-agent system.
// A host agent is responsible for deciding which specialist to 'hand off' the task to.
// One or more specialist agents are responsible for completing the task.
type MultiAgent struct {
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
}

func (ma *MultiAgent) Generate(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
	composeOptions := agent.GetComposeOptions(opts...)

	handler := convertCallbacks(opts...)
	if handler != nil {
		composeOptions = append(composeOptions, compose.WithCallbacks(handler).DesignateNode(hostName))
	}

	return ma.runnable.Invoke(ctx, input, composeOptions...)
}

func (ma *MultiAgent) Stream(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	composeOptions := agent.GetComposeOptions(opts...)

	handler := convertCallbacks(opts...)
	if handler != nil {
		composeOptions = append(composeOptions, compose.WithCallbacks(handler).DesignateNode(hostName))
	}

	return ma.runnable.Stream(ctx, input, composeOptions...)
}

// MultiAgentConfig is the config for host multi-agent system.
type MultiAgentConfig struct {
	Host        Host
	Specialists []*Specialist

	Name string // the name of the host multi agent
}

func (conf *MultiAgentConfig) validate() error {
	if conf == nil {
		return errors.New("host multi agent config is nil")
	}

	if conf.Host.ChatModel == nil {
		return errors.New("host multi agent host ChatModel is nil")
	}

	if len(conf.Specialists) == 0 {
		return errors.New("host multi agent specialists are empty")
	}

	if len(conf.Host.SystemPrompt) == 0 {
		conf.Host.SystemPrompt = defaultHostPrompt
	}

	for _, s := range conf.Specialists {
		if s.ChatModel == nil && s.Invokable == nil && s.Streamable == nil {
			return fmt.Errorf("specialist %s has no chat model or Invokable or Streamable", s.Name)
		}

		if err := s.AgentMeta.validate(); err != nil {
			return err
		}
	}

	if len(conf.Name) == 0 {
		conf.Name = "host multi agent"
	}

	return nil
}

// AgentMeta is the meta information of an agent within a multi-agent system.
type AgentMeta struct {
	Name        string // the name of the agent, should be unique within multi-agent system
	IntendedUse string // the intended use-case of the agent, used as the reason for the multi-agent system to hand over control to this agent
}

func (am AgentMeta) validate() error {
	if len(am.Name) == 0 {
		return errors.New("agent meta name is empty")
	}

	if len(am.IntendedUse) == 0 {
		return errors.New("agent meta intended use is empty")
	}

	return nil
}

// Host is the host agent within a multi-agent system.
// Currently, it can only be a model.ChatModel.
type Host struct {
	ChatModel    model.ChatModel
	SystemPrompt string
}

// Specialist is a specialist agent within a host multi-agent system.
// It can be a model.ChatModel or any Invokable and/or Streamable, such as react.Agent.
// ChatModel and (Invokable / Streamable) are mutually exclusive, only one should be provided.
// If Invokable is provided but not Streamable, then the Specialist will be compose.InvokableLambda.
// If Streamable is provided but not Invokable, then the Specialist will be compose.StreamableLambda.
// if Both Invokable and Streamable is provided, then the Specialist will be compose.AnyLambda.
type Specialist struct {
	AgentMeta

	ChatModel    model.ChatModel
	SystemPrompt string

	Invokable  compose.Invoke[[]*schema.Message, *schema.Message, agent.AgentOption]
	Streamable compose.Stream[[]*schema.Message, *schema.Message, agent.AgentOption]
}
