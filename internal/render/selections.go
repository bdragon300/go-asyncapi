package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type TemplateSelections struct {
	Criteria any // TODO: params from config
	Selections []common.Renderer
}

func (r TemplateSelections) SelectLangs() []common.Renderer {
	return lo.Filter(r.Selections, func(item common.Renderer, _ int) bool {
		return item.Kind() == common.ObjectKindLang
	})
}

func (r TemplateSelections) SelectSchemas() []*lang.GoStruct {
	return selectObjects[*lang.GoStruct](r.Selections, common.ObjectKindSchema)
}

func (r TemplateSelections) SelectServers() []*ProtoServer {
	return selectObjects[*ProtoServer](r.Selections, common.ObjectKindServer)
}

func (r TemplateSelections) SelectServerVariables() []*ServerVariable {
	return selectObjects[*ServerVariable](r.Selections, common.ObjectKindServerVariable)
}

func (r TemplateSelections) SelectChannels() []*ProtoChannel {
	return selectObjects[*ProtoChannel](r.Selections, common.ObjectKindChannel)
}

func (r TemplateSelections) SelectMessages() []*ProtoMessage {
	return selectObjects[*ProtoMessage](r.Selections, common.ObjectKindMessage)
}

func (r TemplateSelections) SelectParameters() []*Parameter {
	return selectObjects[*Parameter](r.Selections, common.ObjectKindParameter)
}

func (r TemplateSelections) SelectCorrelationIDs() []*CorrelationID {
	return selectObjects[*CorrelationID](r.Selections, common.ObjectKindCorrelationID)
}

func (r TemplateSelections) SelectServerBindings() []*Bindings {
	return selectObjects[*Bindings](r.Selections, common.ObjectKindServerBindings)
}

func (r TemplateSelections) SelectChannelBindings() []*Bindings {
	return selectObjects[*Bindings](r.Selections, common.ObjectKindChannelBindings)
}

func (r TemplateSelections) SelectOperationBindings() []*Bindings {
	return selectObjects[*Bindings](r.Selections, common.ObjectKindOperationBindings)
}

func (r TemplateSelections) SelectMessageBindings() []*Bindings {
	return selectObjects[*Bindings](r.Selections, common.ObjectKindMessageBindings)
}

func selectObjects[T common.Renderer](selections []common.Renderer, kind common.ObjectKind) []T {
	return lo.FilterMap(selections, func(item common.Renderer, _ int) (T, bool) {
		if item.Kind() == kind {
			return item.(T), true
		}
		return lo.Empty[T](), false
	})
}
