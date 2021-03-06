package processing

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
)

func getGate4JWT() *gatewayv2alpha1.Gate {
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

func getJWTConfig() *gatewayv2alpha1.JWTModeConfig {
	return &gatewayv2alpha1.JWTModeConfig{
		Issuer: "http://dex.someDomain.local",
		Mode: gatewayv2alpha1.InternalConfig{
			Name:   gatewayv2alpha1.JWTAll,
			Config: &runtime.RawExtension{},
		},
	}
}

func TestGenerateAuthenticationPolicy(t *testing.T) {
	assert := assert.New(t)

	jwtStrategy := &jwt{JWKSURI: "http://dex-service.namespace.svc.cluster.local:5556/keys"}
	ap := jwtStrategy.generateAuthenticationPolicy(getGate4JWT(), getJWTConfig())

	assert.Equal(ap.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(ap.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(ap.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(ap.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))
	assert.Equal(ap.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(ap.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(len(ap.Spec.Targets), 1)
	assert.Equal(ap.Spec.Targets[0].Name, "test-service")
	assert.Equal(ap.Spec.PrincipalBinding, authenticationv1alpha1.PrincipalBindingUserOrigin)
	assert.Equal(len(ap.Spec.Peers), 1)
	assert.Equal(*ap.Spec.Peers[0].Mtls, authenticationv1alpha1.MutualTLS{})
	assert.Equal(len(ap.Spec.Origins), 1)
	assert.Equal(ap.Spec.Origins[0].Jwt.Issuer, "http://dex.someDomain.local")
	assert.Equal(ap.Spec.Origins[0].Jwt.JwksURI, "http://dex-service.namespace.svc.cluster.local:5556/keys")
}

func TestOauthGenerateVirtualService4JWT(t *testing.T) {
	assert := assert.New(t)

	jwtStrategy := &jwt{JWKSURI: "http://dex-service.namespace.svc.cluster.local:5556/keys"}
	vs := jwtStrategy.generateVirtualService(getGate4JWT())

	assert.Equal(len(vs.Spec.Gateways), 1)
	assert.Equal(vs.Spec.Gateways[0], apiGateway)

	assert.Equal(len(vs.Spec.Hosts), 1)
	assert.Equal(vs.Spec.Hosts[0], serviceHost)

	assert.Equal(len(vs.Spec.HTTP), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(vs.Spec.HTTP[0].Route[0].Destination.Host, "test-service.test-namespace.svc.cluster.local")
	assert.Equal(int(vs.Spec.HTTP[0].Route[0].Destination.Port.Number), 8080)
	assert.Equal(vs.Spec.HTTP[0].Match[0].URI.Regex, "/.*")

	assert.Equal(vs.ObjectMeta.Name, "test-gate-test-service")
	assert.Equal(vs.ObjectMeta.Namespace, "test-namespace")

	assert.Equal(vs.ObjectMeta.OwnerReferences[0].APIVersion, "gateway.kyma-project.io/v2alpha1")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Kind, "Gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].Name, "test-gate")
	assert.Equal(vs.ObjectMeta.OwnerReferences[0].UID, types.UID("eab0f1c8-c417-11e9-bf11-4ac644044351"))

}

func TestPrepareAuthenticationPolicy(t *testing.T) {
	assert := assert.New(t)

	jwtStrategy := &jwt{JWKSURI: "http://dex-service.namespace.svc.cluster.local:5556/keys"}
	gate := getGate4JWT()
	jwtConfig := getJWTConfig()
	jwtConfig.Issuer = "http://someIssuer.someDomain.com"

	currentAP := jwtStrategy.generateAuthenticationPolicy(gate, getJWTConfig())
	currentAP.ObjectMeta.Generation = int64(42)

	newAP := jwtStrategy.prepareAuthenticationPolicy(gate, jwtConfig, currentAP)

	assert.Equal(newAP.ObjectMeta.Generation, int64(42))
	assert.Equal(newAP.Spec.Origins[0].Jwt.Issuer, jwtConfig.Issuer)
}
