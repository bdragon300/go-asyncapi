package asyncapi

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type SecurityScheme struct {
	Type             string      `json:"type,omitzero" yaml:"type"`
	Description      string      `json:"description,omitzero" yaml:"description"`
	Name             string      `json:"name,omitzero" yaml:"name"`
	In               string      `json:"in,omitzero" yaml:"in"`
	Scheme           string      `json:"scheme,omitzero" yaml:"scheme"`
	BearerFormat     string      `json:"bearerFormat,omitzero" yaml:"bearerFormat"`
	Flows            *OAuthFlows `json:"flows,omitzero" yaml:"flows"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitzero" yaml:"openIdConnectUrl"`
	Scopes           []string    `json:"scopes,omitzero" yaml:"scopes"`

	XGoName string `json:"x-go-name,omitzero" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (ss SecurityScheme) Compile(ctx *compile.Context) error {
	obj, err := ss.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (ss SecurityScheme) build(ctx *compile.Context, securitySchemeKey string) (common.Artifact, error) {
	if ss.Ref != "" {
		return registerRef(ctx, ss.Ref, securitySchemeKey, nil), nil
	}
	schemeName, _ := lo.Coalesce(ss.XGoName, securitySchemeKey)
	if ss.Type == "" {
		return nil, types.CompileError{Err: fmt.Errorf("security scheme type is empty"), Path: ctx.CurrentRefPointer()}
	}
	res := render.SecurityScheme{
		OriginalName: ctx.GenerateObjName(schemeName, ""),
		SchemeType:   ss.Type,
		Description:  ss.Description,
		Dummy:        ss.XIgnore,
		Params:       ss.buildValues(),
	}

	// List promises for all secured servers and operations
	prm := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		v, ok := item.(*render.Server)
		return ok && v.Visible() && len(v.SecuritySchemePromises) > 0
	}, nil)
	ctx.PutListPromise(prm)
	res.AllSecuredServersPromise = prm

	prm = lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		v, ok := item.(*render.Operation)
		return ok && v.Visible() && len(v.SecuritySchemePromises) > 0
	}, nil)
	ctx.PutListPromise(prm)
	res.AllSecuredOperationsPromise = prm

	return &res, nil
}

func (ss SecurityScheme) buildValues() render.SecuritySchemeParams {
	setOAuthParams := func(v *OAuthFlow) *render.SecuritySchemeOAuthFlowParams {
		if v == nil {
			return nil
		}
		return &render.SecuritySchemeOAuthFlowParams{
			AuthorizationURL: v.AuthorizationURL,
			TokenURL:         v.TokenURL,
			RefreshURL:       v.RefreshURL,
			AvailableScopes:  v.AvailableScopes,
		}
	}

	res := render.SecuritySchemeParams{
		In:               ss.In,
		Scheme:           ss.Scheme,
		BearerFormat:     ss.BearerFormat,
		OpenIDConnectURL: ss.OpenIDConnectURL,
		Scopes:           ss.Scopes,
	}
	if ss.Flows != nil {
		res.Flows = &render.SecuritySchemeOAuthFlowsParams{}

		res.Flows.Implicit = setOAuthParams(ss.Flows.Implicit)
		res.Flows.Password = setOAuthParams(ss.Flows.Password)
		res.Flows.ClientCredentials = setOAuthParams(ss.Flows.ClientCredentials)
		res.Flows.AuthorizationCode = setOAuthParams(ss.Flows.AuthorizationCode)
	}

	return res
}

type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitzero" yaml:"implicit"`
	Password          *OAuthFlow `json:"password,omitzero" yaml:"password"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitzero" yaml:"clientCredentials"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitzero" yaml:"authorizationCode"`
}

type OAuthFlow struct {
	AuthorizationURL string                           `json:"authorizationUrl,omitzero" yaml:"authorizationUrl"`
	TokenURL         string                           `json:"tokenUrl,omitzero" yaml:"tokenUrl"`
	RefreshURL       string                           `json:"refreshUrl,omitzero" yaml:"refreshUrl"`
	AvailableScopes  types.OrderedMap[string, string] `json:"availableScopes,omitzero" yaml:"availableScopes"`
}
