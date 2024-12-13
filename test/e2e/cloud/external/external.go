package external

// NodeManager defines operations related to node lifecycle management.
type NodeManager interface {
	// DeleteNodeIfNotExists ensures that nodes not present in the cloud provider are deleted from the API server.
	DeleteNodeIfNotExists(nodeName string) error

	// RebootNode performs an ordered clean reboot of the specified node.
	RebootNode(nodeName string) error

	// EnsureNodeFunctionality validates that a node is functional after restart.
	EnsureNodeFunctionality(nodeName string) error
}

// UpgradeManager defines operations for managing cluster and master upgrades.
type UpgradeManager interface {
	// UpgradeMaster upgrades the master node and ensures the cluster remains functional.
	UpgradeMaster(version string) error

	// UpgradeCluster upgrades all cluster nodes and ensures the cluster remains functional.
	UpgradeCluster(version string) error

	// DowngradeCluster downgrades the cluster to a specified version and ensures it remains functional.
	DowngradeCluster(version string) error
}

// PodManager defines operations for managing and validating pod functionality.
type PodManager interface {
	// GetPodCount retrieves the number of pods running and ready in the cluster.
	GetPodCount() (int, error)

	// ValidatePodCount ensures the number of pods running and ready matches the expected count.
	ValidatePodCount(expectedCount int) error

	// EnsurePodFunctionality validates that all pods are running and ready after a cluster event (e.g., restart, upgrade).
	EnsurePodFunctionality() error
}

// LoadBalancerManager defines operations for managing and testing LoadBalancer behavior.
type LoadBalancerManager interface {
	// ChangeTypeAndPorts updates the type and ports of a LoadBalancer service.
	ChangeTypeAndPorts(serviceName string, newType string, ports []int) error

	// ValidateSessionAffinity ensures session affinity works for LoadBalancer services with Local traffic policy.
	ValidateSessionAffinity(serviceName string) error

	// CleanupFinalizer validates the cleanup of LoadBalancer finalizers for services.
	CleanupFinalizer(serviceName string) error

	// CreateWithoutNodePort creates a LoadBalancer service without a NodePort and validates its functionality.
	CreateWithoutNodePort(serviceName string) error

	// PreserveUDPTrafficAcrossNodes ensures UDP traffic is preserved when server pods cycle across nodes.
	PreserveUDPTrafficAcrossNodes(serviceName string) error

	// PreserveUDPTrafficSameNode ensures UDP traffic is preserved when server pods cycle on the same node.
	PreserveUDPTrafficSameNode(serviceName string) error

	// NoDisruptionDuringRollingUpdate ensures no connectivity disruption during rolling updates.
	NoDisruptionDuringRollingUpdate(serviceName string, externalTrafficPolicy string) error

	// ValidateExternalTrafficPolicyLocal validates ExternalTrafficPolicy: Local for LoadBalancer services.
	ValidateExternalTrafficPolicyLocal(serviceName string) error

	// TargetAllNodesWithEndpoints ensures all nodes with endpoints are targeted for ExternalTrafficPolicy: Local.
	TargetAllNodesWithEndpoints(serviceName string) error
}

// ServiceManager defines operations for managing and testing Kubernetes Services.
type ServiceManager interface {
	// SecureMasterService ensures the master service is secure.
	SecureMasterService() error

	// MultiportEndpoints validates services serve multiport endpoints from pods.
	MultiportEndpoints(serviceName string) error

	// UpdatePorts ensures services are updated after adding or deleting ports.
	UpdatePorts(serviceName string, ports []int) error

	// PreserveSourcePodIP ensures the source pod IP is preserved for traffic through service cluster IP.
	PreserveSourcePodIP(serviceName string) error

	// AllowHairpinTraffic ensures pods can hairpin back to themselves through services.
	AllowHairpinTraffic(serviceName string) error

	// ServiceLifecycle tests creating, updating, and deleting services.
	ServiceLifecycle(serviceName string) error

	// ValidateServiceAfterRestart ensures services work after restarting kube-proxy or the API server.
	ValidateServiceAfterRestart(serviceName string, component string) error

	// NodePortService validates the creation of a functioning NodePort service.
	NodePortService(serviceName string) error

	// ExternalIPService validates connectivity to a service via ExternalIP.
	ExternalIPService(serviceName string, externalIP string) error

	// PreventNodePortCollisions ensures no NodePort collisions occur.
	PreventNodePortCollisions(serviceName string) error

	// ValidateSessionAffinity validates session affinity for various service types.
	ValidateSessionAffinity(serviceName string, serviceType string) error

	// TestServiceProxyName validates the implementation of service.kubernetes.io/service-proxy-name.
	TestServiceProxyName(serviceName string) error

	// TestHeadlessService validates the implementation of headless services.
	TestHeadlessService(serviceName string) error

	// ValidateInternalTrafficPolicy ensures internalTrafficPolicy settings work as expected.
	ValidateInternalTrafficPolicy(serviceName string, policy string) error

	// ValidateExternalTrafficPolicy ensures externalTrafficPolicy settings work as expected.
	ValidateExternalTrafficPolicy(serviceName string, policy string) error

	// LifecycleEndpoints tests the lifecycle of endpoints for a service.
	LifecycleEndpoints(serviceName string) error

	// ValidateNodePortAndHealthCheckNodePort validates NodePort and HealthCheckNodePort behavior.
	ValidateNodePortAndHealthCheckNodePort(serviceName string) error
}

// CloudControllerManagerValidate aggregates interfaces to manage nodes, upgrades, pods, LoadBalancers, and services in the cluster.
type CloudControllerManagerValidate interface {
	NodeManager
	UpgradeManager
	PodManager
	LoadBalancerManager
	ServiceManager
}
