package processing

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestOauthGenerateVirtualService(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	oauthConfig := getOauthConfig()
	oauthStrategy := &oauth{oathkeeperSvc: "test-oathkeeper", oathkeeperSvcPort: uint32(4455)}

	vs := oauthStrategy.generateVirtualService(gate, oauthConfig)

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(int(vs.Spec.HTTP[0].Route[0].Destination.Port.Number), 4455)
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, "/foo")

	assert.Equal(vs.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(vs.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthPrepareVirtualService(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	oauthConfig := getOauthConfig()
	oauthStrategy := &oauth{oathkeeperSvc: "test-oathkeeper", oathkeeperSvcPort: uint32(4455)}

	oldVS := oauthStrategy.generateVirtualService(gate, oauthConfig)

	oldVS.ObjectMeta.Generation = int64(15)
	oldVS.ObjectMeta.Name = "mst"

	newVS := oauthStrategy.prepareVirtualService(gate, oldVS, oauthConfig)

	assert.Equal(newVS.ObjectMeta.Generation, int64(15))

	assert.Equal(len(newVS.Spec.Gateways), 1)
	assert.Equal(newVS.Spec.Gateways[0], apiGateway)

	assert.Equal(len(newVS.Spec.Hosts), 1)
	assert.Equal(newVS.Spec.Hosts[0], serviceHost)

	assert.Equal(len(newVS.Spec.HTTP), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Route), 1)
	assert.Equal(len(newVS.Spec.HTTP[0].Match), 1)
	assert.Equal(newVS.Spec.HTTP[0].Route[0].Destination.Host, "test-oathkeeper")
	assert.Equal(int(newVS.Spec.HTTP[0].Route[0].Destination.Port.Number), 4455)
	assert.Equal(newVS.Spec.HTTP[0].Match[0].URI.Regex, "/foo")

	assert.Equal(newVS.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(newVS.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(newVS.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthGenerateAccessRule(t *testing.T) {
	assert := assert.New(t)

	gate := getGate()
	oauthConfig := getOauthConfig()
	requiredScopes := []byte(`required_scopes: ["write", "read"]`)

	ar := generateAccessRule(gate, &oauthConfig.Paths[0], requiredScopes)

	assert.Equal(len(ar.Spec.Authenticators), 1)
	assert.NotEmpty(ar.Spec.Authenticators[0].Config)
	assert.Equal(string(ar.Spec.Authenticators[0].Config.Raw), string(requiredScopes))

	assert.Equal(len(ar.Spec.Match.Methods), 1)
	assert.Equal(ar.Spec.Match.Methods[0], "GET")
	assert.Equal(ar.Spec.Match.URL, "<http|https>://myService.myDomain.com</foo>")

	assert.Equal(ar.Spec.Authorizer.Name, "allow")
	assert.Empty(ar.Spec.Authorizer.Config)

	assert.Equal(ar.Spec.Upstream.URL, "http://test-service.test-namespace.svc.cluster.local:8080")

	assert.Equal(ar.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(ar.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(ar.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(ar.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestOauthPrepareAccessRule(t *testing.T) {
	assert := assert.New(t)

	oauthStrategy := &oauth{oathkeeperSvc: "test-oathkeeper"}
	gate := getGate()
	oauthConfig := getOauthConfig()
	requiredScopes := []byte(`required_scopes: ["write", "read"]`)

	oldAR := generateAccessRule(gate, &oauthConfig.Paths[0], requiredScopes)

	oldAR.ObjectMeta.Generation = int64(15)
	oldAR.ObjectMeta.Name = "mst"

	newAR := oauthStrategy.prepareAccessRule(gate, oldAR, &oauthConfig.Paths[0], requiredScopes)

	assert.Equal(newAR.ObjectMeta.Generation, int64(15))

	assert.Equal(len(oldAR.Spec.Authenticators), 1)
	assert.NotEmpty(oldAR.Spec.Authenticators[0].Config)
	assert.Equal(string(oldAR.Spec.Authenticators[0].Config.Raw), string(requiredScopes))

	assert.Equal(len(oldAR.Spec.Match.Methods), 1)
	assert.Equal(oldAR.Spec.Match.Methods[0], "GET")
	assert.Equal(oldAR.Spec.Match.URL, "<http|https>://myService.myDomain.com</foo>")

	assert.Equal(oldAR.Spec.Authorizer.Name, "allow")
	assert.Empty(oldAR.Spec.Authorizer.Config)

	assert.Equal(oldAR.Spec.Upstream.URL, "http://test-service.test-namespace.svc.cluster.local:8080")

	assert.Equal(oldAR.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(oldAR.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(oldAR.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(oldAR.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(oldAR.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(oldAR.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func getGate() *gatewayv2alpha1.Gate {
	var apiUID types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	var apiGateway = "some-gateway"
	var serviceName = "test-service"
	var serviceHost = "myService.myDomain.com"
	var servicePort int32 = 8080

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gate",
			UID:       apiUID,
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v2alpha1",
			Kind:       "Gate",
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &apiGateway,
			Service: &gatewayv2alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
		},
	}
}

func getOauthConfig() *gatewayv2alpha1.OauthModeConfig {
	return &gatewayv2alpha1.OauthModeConfig{
		Paths: []gatewayv2alpha1.Option{
			{
				Path:    "/foo",
				Scopes:  []string{"write", "read"},
				Methods: []string{"GET"},
			},
		},
	}
}
