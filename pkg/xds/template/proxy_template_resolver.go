package template

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	model "github.com/kumahq/kuma/pkg/core/xds"
)

var (
	templateResolverLog = core.Log.WithName("proxy-template-resolver")
)

type ProxyTemplateResolver interface {
	GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate
}

type SimpleProxyTemplateResolver struct {
	ReadOnlyResourceManager manager.ReadOnlyResourceManager
}

func (r *SimpleProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	log := templateResolverLog.WithValues("dataplane", core_model.MetaToResourceKey(proxy.Dataplane.Meta))
	ctx := context.Background()
	templateList := &mesh_core.ProxyTemplateResourceList{}
	if err := r.ReadOnlyResourceManager.List(ctx, templateList, core_store.ListByMesh(proxy.Dataplane.Meta.GetMesh())); err != nil {
		templateResolverLog.Error(err, "failed to list ProxyTemplates")
		return nil
	}

	policies := make([]core_policy.DataplanePolicy, len(templateList.Items))
	for i, proxyTemplate := range templateList.Items {
		policies[i] = proxyTemplate
	}

	if bestMatchTemplate := core_policy.SelectDataplanePolicy(proxy.Dataplane, policies); bestMatchTemplate != nil {
		log.V(2).Info("found the best matching ProxyTemplate", "proxytemplate", core_model.MetaToResourceKey(bestMatchTemplate.GetMeta()))
		return bestMatchTemplate.(*mesh_core.ProxyTemplateResource).Spec
	}

	log.V(2).Info("no matching ProxyTemplate")
	return nil
}

type StaticProxyTemplateResolver struct {
	Template *mesh_proto.ProxyTemplate
}

func (r *StaticProxyTemplateResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	return r.Template
}

type sequentialResolver []ProxyTemplateResolver

func (s sequentialResolver) GetTemplate(proxy *model.Proxy) *mesh_proto.ProxyTemplate {
	for _, r := range s {
		if t := r.GetTemplate(proxy); t != nil {
			return t
		}
	}

	return nil
}

// SequentialResolver returns a new ProxyTemplate resolver that applies
// each of the resolvers given as arguments in turn. The result of the
// first successful resolver is returned.
func SequentialResolver(r ...ProxyTemplateResolver) ProxyTemplateResolver {
	return sequentialResolver(r)
}
