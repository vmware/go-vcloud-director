package govcd

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const cciLabelKubeConfig = "Kube Config"

type KubeConfigValues struct {
	ContextName   string
	ClusterName   string
	ClusterServer string
	UserName      string
	Token         *jwt.Token
}

// GetKubeConfig retrieves a representation of KubeConfig.
// Org name is mandatory. Project and Supervisor Namespace names are optional
func (cciClient *CciClient) GetKubeConfig(orgName, projectName, supervisorNamespaceName string) (*clientcmdapi.Config, *KubeConfigValues, error) {
	if orgName == "" {
		return nil, nil, fmt.Errorf("Org name is mandatory for %s", cciLabelKubeConfig)
	}

	clusterServerUrl, err := cciClient.GetCciUrl()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting CCI Url: %s", err)
	}
	cciHost := clusterServerUrl.Host
	contextName := orgName
	clusterName := fmt.Sprintf("%s:%s", orgName, cciHost)

	clusterServer := clusterServerUrl.String()

	// if Project Name and Supervisor Namespace were specified - set the context for it
	if projectName != "" && supervisorNamespaceName != "" {
		supervisorNamespace, err := cciClient.GetSupervisorNamespaceByName(projectName, supervisorNamespaceName)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading %s: %s", cciLabelSupervisorNamespace, err)
		}
		readyStatus := false
		for _, condition := range supervisorNamespace.SupervisorNamespace.Status.Conditions {
			if strings.ToLower(condition.Type) == "ready" {
				if strings.ToLower(condition.Status) == "true" {
					readyStatus = true
				}
				break
			}
		}
		if !readyStatus {
			return nil, nil, fmt.Errorf("%s %s is not in a ready status", cciLabelSupervisorNamespace, supervisorNamespaceName)
		}
		if supervisorNamespace.SupervisorNamespace.Status.NamespaceEndpointURL == "" {
			return nil, nil, fmt.Errorf("unable to retrieve the endpoint URL for %s %s", cciLabelSupervisorNamespace, supervisorNamespaceName)
		}
		clusterName = fmt.Sprintf("%s:%s@%s", orgName, supervisorNamespaceName, cciHost)
		clusterServer = supervisorNamespace.SupervisorNamespace.Status.NamespaceEndpointURL
		contextName = fmt.Sprintf("%s:%s:%s", orgName, supervisorNamespaceName, projectName)
	}

	token, _, err := new(jwt.Parser).ParseUnverified(cciClient.VCDClient.Client.VCDToken, jwt.MapClaims{})
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing JWT token: %s", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("could not parse claims from JWT token")
	}
	preferredUsername, ok := claims["preferred_username"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("could not parse preferred username from JWT token claims")
	}
	username := fmt.Sprintf("%s:%s@%s", orgName, preferredUsername, cciHost)

	kubeconfig := &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: clientcmdapi.SchemeGroupVersion.Version,
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				InsecureSkipTLSVerify: cciClient.VCDClient.Client.InsecureSkipVerify,
				Server:                clusterServer,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: username,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			username: {
				Token: token.Raw,
			},
		},
		CurrentContext: contextName,
	}

	if projectName != "" && supervisorNamespaceName != "" {
		kubeconfig.Contexts[contextName].Namespace = supervisorNamespaceName
	}

	r := &KubeConfigValues{
		ContextName:   contextName,
		ClusterName:   clusterName,
		ClusterServer: clusterServer,
		UserName:      username,
		Token:         token,
	}

	return kubeconfig, r, nil
}
