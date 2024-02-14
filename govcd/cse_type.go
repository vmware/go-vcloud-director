package govcd

import (
	"embed"
	semver "github.com/hashicorp/go-version"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"time"
)

// CseKubernetesCluster is a type for managing an existing Kubernetes cluster created by the Container Service Extension (CSE)
type CseKubernetesCluster struct {
	CseClusterSettings
	ID                         string
	Etag                       string
	KubernetesVersion          semver.Version
	TkgVersion                 semver.Version
	CapvcdVersion              semver.Version
	ClusterResourceSetBindings []string
	CpiVersion                 semver.Version
	CsiVersion                 semver.Version
	State                      string
	Events                     []CseClusterEvent

	client            *Client
	capvcdType        *types.Capvcd
	supportedUpgrades []*types.VAppTemplate // Caches the vApp templates that can be used to upgrade a cluster.
}

// CseClusterEvent is an event that has occurred during the lifetime of a Container Service Extension (CSE) Kubernetes cluster.
type CseClusterEvent struct {
	Name         string
	Type         string
	ResourceId   string
	ResourceName string
	OccurredAt   time.Time
	Details      string
}

// CseClusterSettings defines the required configuration of a Container Service Extension (CSE) Kubernetes cluster.
type CseClusterSettings struct {
	CseVersion              semver.Version
	Name                    string
	OrganizationId          string
	VdcId                   string
	NetworkId               string
	KubernetesTemplateOvaId string
	ControlPlane            CseControlPlaneSettings
	WorkerPools             []CseWorkerPoolSettings
	DefaultStorageClass     *CseDefaultStorageClassSettings // Optional
	Owner                   string                          // Optional, if not set will pick the current session user from the VCDClient
	ApiToken                string
	NodeHealthCheck         bool
	PodCidr                 string
	ServiceCidr             string
	SshPublicKey            string
	VirtualIpSubnet         string
	AutoRepairOnErrors      bool
}

// CseControlPlaneSettings defines the required configuration of a Control Plane of a Container Service Extension (CSE) Kubernetes cluster.
type CseControlPlaneSettings struct {
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	StorageProfileId  string // Optional
	Ip                string // Optional
}

// CseWorkerPoolSettings defines the required configuration of a Worker Pool of a Container Service Extension (CSE) Kubernetes cluster.
type CseWorkerPoolSettings struct {
	Name              string
	MachineCount      int
	DiskSizeGi        int
	SizingPolicyId    string // Optional
	PlacementPolicyId string // Optional
	VGpuPolicyId      string // Optional
	StorageProfileId  string // Optional
}

// CseDefaultStorageClassSettings defines the required configuration of a Default Storage Class of a Container Service Extension (CSE) Kubernetes cluster.
type CseDefaultStorageClassSettings struct {
	StorageProfileId string
	Name             string
	ReclaimPolicy    string // Must be either "delete" or "retain"
	Filesystem       string // Must be either "ext4" or "xfs"
}

// CseClusterUpdateInput defines the required configuration that a Container Service Extension (CSE) Kubernetes cluster needs in order to be updated.
type CseClusterUpdateInput struct {
	KubernetesTemplateOvaId *string
	ControlPlane            *CseControlPlaneUpdateInput
	WorkerPools             *map[string]CseWorkerPoolUpdateInput // Maps a node pool name with its contents
	NewWorkerPools          *[]CseWorkerPoolSettings
	NodeHealthCheck         *bool
	AutoRepairOnErrors      *bool

	// Private fields that are computed, not requested to the consumer of this struct
	vcdKeConfigVersion string
	clusterName        string
	cseVersion         semver.Version
}

// CseControlPlaneUpdateInput defines the required configuration that the Control Plane of the Container Service Extension (CSE) Kubernetes cluster
// needs in order to be updated.
type CseControlPlaneUpdateInput struct {
	MachineCount int
}

// CseWorkerPoolUpdateInput defines the required configuration that a Worker Pool of the Container Service Extension (CSE) Kubernetes cluster
// needs in order to be updated.
type CseWorkerPoolUpdateInput struct {
	MachineCount int
}

// cseClusterSettingsInternal defines the required arguments that are required by the CSE Server used internally to specify
// a Kubernetes cluster. These are not set by the user, but instead they are computed from a valid
// CseClusterSettings object in the cseClusterSettingsToInternal method. These fields are then
// inserted in Go templates to render a final JSON that is valid to be used as the cluster Runtime Defined Entity (RDE) payload.
//
// The main difference between CseClusterSettings and this structure is that the first one uses IDs and this one uses names, among
// other differences like the computed TkgVersionBundle.
type cseClusterSettingsInternal struct {
	CseVersion                semver.Version
	Name                      string
	OrganizationName          string
	VdcName                   string
	NetworkName               string
	KubernetesTemplateOvaName string
	TkgVersionBundle          tkgVersionBundle
	CatalogName               string
	RdeType                   *types.DefinedEntityType
	ControlPlane              cseControlPlaneSettingsInternal
	WorkerPools               []cseWorkerPoolSettingsInternal
	DefaultStorageClass       cseDefaultStorageClassInternal
	VcdKeConfig               vcdKeConfig
	Owner                     string
	ApiToken                  string
	VcdUrl                    string
	VirtualIpSubnet           string
	SshPublicKey              string
	PodCidr                   string
	ServiceCidr               string
	AutoRepairOnErrors        bool
}

// tkgVersionBundle is a type that contains all the versions of the components of
// a Kubernetes cluster that can be obtained with the Kubernetes Template OVA name, downloaded
// from VMware Customer connect:
// https://customerconnect.vmware.com/downloads/details?downloadGroup=TKG-240&productId=1400
type tkgVersionBundle struct {
	EtcdVersion       string
	CoreDnsVersion    string
	TkgVersion        string
	TkrVersion        string
	KubernetesVersion string
}

// cseControlPlaneSettingsInternal defines the Control Plane inside cseClusterSettingsInternal
type cseControlPlaneSettingsInternal struct {
	MachineCount        int
	DiskSizeGi          int
	SizingPolicyName    string
	PlacementPolicyName string
	StorageProfileName  string
	Ip                  string
}

// cseWorkerPoolSettingsInternal defines a Worker Pool inside cseClusterSettingsInternal
type cseWorkerPoolSettingsInternal struct {
	Name                string
	MachineCount        int
	DiskSizeGi          int
	SizingPolicyName    string
	PlacementPolicyName string
	VGpuPolicyName      string
	StorageProfileName  string
}

// cseDefaultStorageClassInternal defines a Default Storage Class inside cseClusterSettingsInternal
type cseDefaultStorageClassInternal struct {
	StorageProfileName     string
	Name                   string
	UseDeleteReclaimPolicy bool
	Filesystem             string
}

// vcdKeConfig is a type that contains only the required and relevant fields from the VCDKEConfig (CSE Server) configuration,
// such as the Machine Health Check settings.
type vcdKeConfig struct {
	MaxUnhealthyNodesPercentage float64
	NodeStartupTimeout          string
	NodeNotReadyTimeout         string
	NodeUnknownTimeout          string
	ContainerRegistryUrl        string
}

// cseComponentsVersions is a type that registers the versions of the subcomponents of a specific CSE Version
type cseComponentsVersions struct {
	VcdKeConfigRdeTypeVersion string
	CapvcdRdeTypeVersion      string
	CseInterfaceVersion       string
}

// This collection of files contains all the Go Templates and resources required for the Container Service Extension (CSE) methods
// to work.
//
//go:embed cse
var cseFiles embed.FS
