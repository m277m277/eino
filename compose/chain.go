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

package compose

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/internal/gmap"
	"github.com/cloudwego/eino/internal/gslice"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/utils/generic"
)

// NewChain create a chain with input/output type.
func NewChain[I, O any]() *Chain[I, O] {
	ch := &Chain[I, O]{
		gg: NewGraph[I, O](),
	}

	ch.gg.graph.addNodeChecker = nodeCheckerOfForbidProcessor(baseNodeChecker)
	ch.gg.graph.runtimeGraphKey = defaultGraphKey()

	return ch
}

// Chain is a chain of components.
// Chain nodes can be parallel / branch / sequence components.
// Chain is designed to be used in a builder pattern (should Compile() before use).
// And the interface is `Chain style`, you can use it like: `chain.AppendXX(...).AppendXX(...)`
//
// Normal usage:
//  1. create a chain with input/output type: `chain := NewChain[inputType, outputType]()`
//  2. add components to chainable list:
//     2.1 add components: `chain.AppendChatTemplate(...).AppendChatModel(...).AppendToolsNode(...)`
//     2.2 add parallel or branch node if needed: `chain.AppendParallel()`, `chain.AppendBranch()`
//  3. compile: `r, err := c.Compile()`
//  4. run:
//     4.1 `one input & one output` use `r.Invoke(ctx, input)`
//     4.2 `one input & multi output chunk` use `r.Stream(ctx, input)`
//     4.3 `multi input chunk & one output` use `r.Collect(ctx, inputReader)`
//     4.4 `multi input chunk & multi output chunk` use `r.Transform(ctx, inputReader)`
//
// Using in graph or other chain:
// chain1 := NewChain[inputType, outputType]()
// graph := NewGraph[](runTypePregel)
// graph.AddGraph("key", chain1) // chain is an AnyGraph implementation
//
// // or in another chain:
// chain2 := NewChain[inputType, outputType]()
// chain2.AppendGraph(chain1)
type Chain[I, O any] struct {
	err error

	gg *Graph[I, O]

	namePrefix string
	nodeIdx    int

	preNodeKeys []string
}

// implements AnyGraph.
func (c *Chain[I, O]) compile(ctx context.Context, option *graphCompileOptions) (*composableRunnable, error) {
	if c.err != nil {
		return nil, c.err
	}

	if !c.gg.isFrozen() {
		err := c.addEnds()
		if err != nil {
			return nil, err
		}
	}
	c.gg.compileChecker = wrapCompileChecker(c.gg.compileChecker, func(options *graphCompileOptions) error {
		if len(option.nodeTriggerMode) != 0 && option.nodeTriggerMode != AnyPredecessor {
			return errors.New("only support AnyPredecessor in chain") // dag not support branch
		}

		return nil
	})

	return c.gg.compile(ctx, option)
}

// addEnds add END edge of the chain/graph.
// only run once when compiling.
func (c *Chain[I, O]) addEnds() error {
	if len(c.preNodeKeys) == 0 {
		return fmt.Errorf("pre node keys not set, number of nodes in chain= %d", len(c.gg.nodes))
	}

	for _, nodeKey := range c.preNodeKeys {
		err := c.gg.AddEdge(nodeKey, END)
		if err != nil {
			return err
		}
	}

	return nil
}

// inputType returns the input type of the chain.
// implements AnyGraph.
func (c *Chain[I, O]) inputType() reflect.Type {
	return generic.TypeOf[I]()
}

// outputType returns the output type of the chain.
// implements AnyGraph.
func (c *Chain[I, O]) outputType() reflect.Type {
	return generic.TypeOf[O]()
}

// compositeType returns the composite type of the chain.
// implements AnyGraph.
func (c *Chain[I, O]) component() component {
	return ComponentOfChain
}

// Compile to a Runnable.
// Runnable can be used directly.
// e.g.
//
//		chain := NewChain[string, string]()
//		r, err := chain.Compile()
//		if err != nil {}
//
//	 	r.Invoke(ctx, input) // ping => pong
//		r.Stream(ctx, input) // ping => stream out
//		r.Collect(ctx, inputReader) // stream in => pong
//		r.Transform(ctx, inputReader) // stream in => stream out
func (c *Chain[I, O]) Compile(ctx context.Context, opts ...GraphCompileOption) (Runnable[I, O], error) {
	if c.err != nil {
		return nil, c.err
	}

	opts = append(opts, withComponent(ComponentOfChain))

	if !c.gg.isFrozen() {
		err := c.addEnds()
		if err != nil {
			return nil, err
		}
	}

	c.gg.compileChecker = wrapCompileChecker(c.gg.compileChecker, func(options *graphCompileOptions) error {
		if len(options.nodeTriggerMode) != 0 && options.nodeTriggerMode != AnyPredecessor {
			return errors.New("only support AnyPredecessor in chain") // dag not support branch
		}

		return nil
	})

	tr, err := c.gg.Compile(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// AppendChatModel add a ChatModel node to the chain.
// e.g.
//
//	model, err := openai.NewChatModel(ctx, config)
//	if err != nil {...}
//	chain.AppendChatModel(model)
func (c *Chain[I, O]) AppendChatModel(node model.ChatModel, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toChatModelNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendChatTemplate add a ChatTemplate node to the chain.
// eg.
//
//	chatTemplate, err := prompt.FromMessages(schema.FString, &schema.Message{
//		Role:    schema.System,
//		Content: "You are acting as a {role}.",
//	})
//
//	chain.AppendChatTemplate(chatTemplate)
func (c *Chain[I, O]) AppendChatTemplate(node prompt.ChatTemplate, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toChatTemplateNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendToolsNode add a ToolsNode node to the chain.
// e.g.
//
//	toolsNode, err := tools.NewToolNode(ctx, &tools.ToolsNodeConfig{
//		Tools: []tools.Tool{...},
//	})
//
//	chain.AppendToolsNode(toolsNode)
func (c *Chain[I, O]) AppendToolsNode(node *ToolsNode, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toToolsNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendDocumentTransformer add a DocumentTransformer node to the chain.
// e.g.
//
//	markdownSplitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderSplitterConfig{})
//
//	chain.AppendDocumentTransformer(markdownSplitter)
func (c *Chain[I, O]) AppendDocumentTransformer(node document.Transformer, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toDocumentTransformerNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendLambda add a Lambda node to the chain.
// Lambda is a node that can be used to implement custom logic.
// e.g.
//
//	lambdaNode := compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) (string, error) {...})
//	chain.AppendLambda(lambdaNode)
//
// Note:
// to create a Lambda node, you need to use `compose.AnyLambda` or `compose.InvokableLambda` or `compose.StreamableLambda` or `compose.TransformableLambda`.
// if you want this node has real stream output, you need to use `compose.StreamableLambda` or `compose.TransformableLambda`, for example.
func (c *Chain[I, O]) AppendLambda(node *Lambda, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toLambdaNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendEmbedding add a Embedding node to the chain.
// e.g.
//
//	embedder, err := openai.NewEmbedder(ctx, config)
//	if err != nil {...}
//	chain.AppendEmbedding(embedder)
func (c *Chain[I, O]) AppendEmbedding(node embedding.Embedder, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toEmbeddingNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendRetriever add a Retriever node to the chain.
// e.g.
//
//		retriever, err := vectorstore.NewRetriever(ctx, config)
//		if err != nil {...}
//		chain.AppendRetriever(retriever)
//
//	 or using fornax knowledge as retriever:
//
//		config := fornaxknowledge.Config{...}
//		retriever, err := fornaxknowledge.NewKnowledgeRetriever(ctx, config)
//		if err != nil {...}
//		chain.AppendRetriever(retriever)
func (c *Chain[I, O]) AppendRetriever(node retriever.Retriever, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toRetrieverNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendLoaderSplitter add a LoaderSplitter node to the chain.
// Deprecated: use AppendLoader instead.
func (c *Chain[I, O]) AppendLoaderSplitter(node document.LoaderSplitter, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toLoaderSplitterNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendLoader adds a Loader node to the chain.
// e.g.
//
//	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{})
//	if err != nil {...}
//	chain.AppendLoader(loader)
func (c *Chain[I, O]) AppendLoader(node document.Loader, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toLoaderNode(node, opts...)
	c.addNode(n)
	return c
}

// AppendIndexer add an Indexer node to the chain.
// Indexer is a node that can store documents.
// e.g.
//
//	vectorStoreImpl, err := vikingdb.NewVectorStorer(ctx, vikingdbConfig) // in components/vectorstore/vikingdb/vectorstore.go
//	if err != nil {...}
//
//	config := vectorstore.IndexerConfig{VectorStore: vectorStoreImpl}
//	indexer, err := vectorstore.NewIndexer(ctx, config)
//	if err != nil {...}
//
//	chain.AppendIndexer(indexer)
func (c *Chain[I, O]) AppendIndexer(node indexer.Indexer, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toIndexerNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendBranch add a conditional branch to chain.
// Each branch within the ChainBranch can be an AnyGraph.
// All branches should either lead to END, or converge to another node within the Chain.
// e.g.
//
//	cb := compose.NewChainBranch(conditionFunc)
//	cb.AddChatTemplate("chat_template_key_01", chatTemplate)
//	cb.AddChatTemplate("chat_template_key_02", chatTemplate2)
//	chain.AppendBranch(cb)
func (c *Chain[I, O]) AppendBranch(b *ChainBranch) *Chain[I, O] { // nolint: byted_s_too_many_lines_in_func
	if b == nil {
		c.reportError(fmt.Errorf("append branch invalid, branch is nil"))
		return c
	}

	if b.err != nil {
		c.reportError(fmt.Errorf("append branch error: %w", b.err))
		return c
	}

	if len(b.key2BranchNode) == 0 {
		c.reportError(fmt.Errorf("append branch invalid, nodeList is empty"))
		return c
	}

	if len(b.key2BranchNode) == 1 {
		c.reportError(fmt.Errorf("append branch invalid, nodeList length = 1"))
		return c
	}

	var startNode string
	if len(c.preNodeKeys) == 0 { // branch appended directly to START
		startNode = START
	} else if len(c.preNodeKeys) == 1 {
		startNode = c.preNodeKeys[0]
	} else {
		c.reportError(fmt.Errorf("append branch invalid, multiple previous nodes: %v ", c.preNodeKeys))
		return c
	}

	pName := c.nextNodeKey("Branch")
	key2NodeKey := make(map[string]string, len(b.key2BranchNode))

	for key := range b.key2BranchNode {
		node := b.key2BranchNode[key]
		nodeKey := fmt.Sprintf("%s[%s]_%s", pName, key, node.getNodeName())
		if err := c.gg.addNode(nodeKey, node); err != nil {
			c.reportError(fmt.Errorf("add branch node[%s] to chain failed: %w", nodeKey, err))
			return c
		}

		key2NodeKey[key] = nodeKey
	}

	condition := &composableRunnable{
		i:                 b.condition.i,
		t:                 b.condition.t,
		inputType:         b.condition.inputType,
		inputStreamFilter: b.condition.inputStreamFilter,
		outputType:        b.condition.outputType,
		optionType:        b.condition.optionType,
		isPassthrough:     b.condition.isPassthrough,
		meta:              b.condition.meta,
		nodeInfo:          b.condition.nodeInfo,
	}

	invokeCon := func(ctx context.Context, in any, opts ...any) (endNode any, err error) {
		endKey, err := b.condition.i(ctx, in, opts...)
		if err != nil {
			return "", err
		}

		endStr, ok := endKey.(string)
		if !ok {
			return "", fmt.Errorf("chain branch result not string, got: %T", endKey)
		}

		nodeKey, ok := key2NodeKey[endStr]
		if !ok {
			return "", fmt.Errorf("chain branch result not in added keys: %s", endStr)
		}

		return nodeKey, nil
	}
	condition.i = invokeCon

	transformCon := func(ctx context.Context, sr streamReader, opts ...any) (streamReader, error) {
		iEndStream, err := b.condition.t(ctx, sr, opts...)
		if err != nil {
			return nil, err
		}

		if iEndStream.getChunkType() != reflect.TypeOf("") {
			return nil, fmt.Errorf("chain branch result not string, got: %v", iEndStream.getChunkType())
		}

		endStream, ok := unpackStreamReader[string](iEndStream)
		if !ok {
			return nil, fmt.Errorf("unpack stream reader not ok")
		}

		endStr, err := concatStreamReader(endStream)
		if err != nil {
			return nil, err
		}

		nodeKey, ok := key2NodeKey[endStr]
		if !ok {
			return nil, fmt.Errorf("chain branch result not in added keys: %s", endStr)
		}

		return packStreamReader(schema.StreamReaderFromArray([]string{nodeKey})), nil
	}
	condition.t = transformCon

	gBranch := &GraphBranch{
		condition: condition,
		endNodes: gslice.ToMap(gmap.Values(key2NodeKey), func(k string) (string, bool) {
			return k, true
		}),
	}

	if err := c.gg.AddBranch(startNode, gBranch); err != nil {
		c.reportError(fmt.Errorf("chain append branch failed: %w", err))
		return c
	}

	c.preNodeKeys = gmap.Values(key2NodeKey)

	return c
}

// AppendParallel add a Parallel structure (multiple concurrent nodes) to the chain.
// e.g.
//
//	parallel := compose.NewParallel()
//	parallel.AddChatModel("openai", model1) // => "openai": *schema.Message{}
//	parallel.AddChatModel("maas", model2) // => "maas": *schema.Message{}
//
//	chain.AppendParallel(parallel) // => multiple concurrent nodes are added to the Chain
//
//	The next node in the chain is either an END, or a node which accepts a map[string]any, where keys are `openai` `maas` as specified above.
func (c *Chain[I, O]) AppendParallel(p *Parallel) *Chain[I, O] {
	if p == nil {
		c.reportError(fmt.Errorf("append parallel invalid, parallel is nil"))
		return c
	}

	if p.err != nil {
		c.reportError(fmt.Errorf("append parallel invalid, parallel error: %w", p.err))
		return c
	}

	if len(p.nodes) <= 1 {
		c.reportError(fmt.Errorf("append parallel invalid, not enough nodes, count = %d", len(p.nodes)))
		return c
	}

	var startNode string
	if len(c.preNodeKeys) == 0 { // parallel appended directly to START
		startNode = START
	} else if len(c.preNodeKeys) == 1 {
		startNode = c.preNodeKeys[0]
	} else {
		c.reportError(fmt.Errorf("append parallel invalid, multiple previous nodes: %v ", c.preNodeKeys))
		return c
	}

	pName := c.nextNodeKey("Parallel")
	var nodeKeys []string

	for i := range p.nodes {
		node := p.nodes[i]
		nodeKey := fmt.Sprintf("%s[%d]_%s", pName, i, node.getNodeName())
		if err := c.gg.addNode(nodeKey, node); err != nil {
			c.reportError(fmt.Errorf("add parallel node[%s] to chain failed: %w", nodeKey, err))
			return c
		}
		if err := c.gg.AddEdge(startNode, nodeKey); err != nil {
			c.reportError(fmt.Errorf("add parallel edge[%s]-[%s] to chain failed: %w", startNode, nodeKey, err))
			return c
		}
		nodeKeys = append(nodeKeys, nodeKey)
	}

	c.preNodeKeys = nodeKeys

	return c
}

// AppendGraph add a AnyGraph node to the chain.
// AnyGraph can be a chain or a graph.
// e.g.
//
//	graph := compose.NewGraph[string, string]()
//	chain.AppendGraph(graph)
func (c *Chain[I, O]) AppendGraph(node AnyGraph, opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toAnyGraphNode(node, opts...)

	c.addNode(n)
	return c
}

// AppendPassthrough add a Passthrough node to the chain.
// Could be used to connect multiple ChainBranch or Parallel.
// e.g.
//
//	chain.AppendPassthrough()
func (c *Chain[I, O]) AppendPassthrough(opts ...GraphAddNodeOpt) *Chain[I, O] {
	n := toPassthroughNode(opts...)

	c.addNode(n)
	return c
}

// nextNodeKey.
// get the next node key for the chain.
// e.g. "Chain[1]_ChatModel" => represent the second node of the chain, and is a ChatModel node.
// e.g. "Chain[2]_NameByUser" => represent the third node of the chain, and the node name is set by user of `NameByUser`.
func (c *Chain[I, O]) nextNodeKey(name string) string {
	if c.namePrefix == "" {
		c.namePrefix = string(ComponentOfChain)
	}
	fullKey := fmt.Sprintf("%s[%d]_%s", c.namePrefix, c.nodeIdx, name)
	c.nodeIdx++
	return fullKey
}

// reportError.
// save the first error in the chain.
func (c *Chain[I, O]) reportError(err error) {
	if c.err == nil {
		c.err = err
	}
}

// addNode.
// add a node to the chain.
func (c *Chain[I, O]) addNode(node *graphNode) {
	if c.err != nil {
		return
	}

	if node == nil {
		c.reportError(fmt.Errorf("chain add node invalid, node is nil"))
		return
	}

	nodeKey := c.nextNodeKey(node.getNodeName())
	if node.nodeInfo.key != "" {
		nodeKey = node.nodeInfo.key
	}
	err := c.gg.addNode(nodeKey, node)
	c.reportError(err)

	if len(c.preNodeKeys) == 0 {
		c.preNodeKeys = append(c.preNodeKeys, START)
	}

	for _, preNodeKey := range c.preNodeKeys {
		err := c.gg.AddEdge(preNodeKey, nodeKey)
		if err != nil {
			c.reportError(err)
			return
		}
	}

	c.preNodeKeys = []string{nodeKey}
}
