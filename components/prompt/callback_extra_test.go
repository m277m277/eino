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

package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudwego/eino/schema"
)

func TestConvPrompt(t *testing.T) {
	assert.NotNil(t, ConvCallbackInput(&CallbackInput{}))
	assert.NotNil(t, ConvCallbackInput(map[string]any{}))
	assert.Nil(t, ConvCallbackInput("asd"))

	assert.NotNil(t, ConvCallbackOutput(&CallbackOutput{}))
	assert.NotNil(t, ConvCallbackOutput([]*schema.Message{}))
	assert.Nil(t, ConvCallbackOutput("asd"))
}
