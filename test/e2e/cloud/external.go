package cloud

// CloudProvider defines an interface for testing multiple cloud providers.
type CloudProvider interface {
	// Name returns the name of the cloud provider.
	Name() string

	// DeleteNodes returns a function to delete nodes and a boolean indicating support.
	DeleteNodes() (func(nodes []string) error, bool)

	// CreateInstance returns a function to create instances and a boolean indicating support.
	CreateInstance() (func(name string, config InstanceConfig) error, bool)

	// LoadBalancer returns a function to manage load balancers and a boolean indicating support.
	LoadBalancer() (func(lbName string, action string) error, bool)

	// UpgradeMaster ensures master upgrades maintain a functioning cluster.
	UpgradeMaster() (func() error, bool)

	// UpgradeCluster ensures cluster upgrades maintain a functioning cluster.
	UpgradeCluster() (func() error, bool)

	// DowngradeCluster ensures downgrades maintain a functioning cluster.
	DowngradeCluster() (func() error, bool)

	// RebootNodes reboots each node and ensures functionality upon restart.
	RebootNodes() (func(nodes []string) error, bool)

	// VerifyPodCount ensures the same number of pods are running and ready after restart.
	VerifyPodCount() (func() error, bool)

	// ServiceTests enumerates possible service tests related to networking.
	ServiceTests() (func() error, bool)
}

// InstanceConfig holds the configuration for creating instances.
type InstanceConfig struct {
	MachineType string
	Region      string
	Zone        string
}

// Test function that checks if the cloud provider creates Instance.
func TestCreateInstance(provider CloudProvider) {
	if instanceCreator, supported := provider.CreateInstance(); !supported {
		skipTest("CreateInstance not supported by " + provider.Name())
	} else {
		if err := instanceCreator("test", InstanceConfig{}); err != nil {
			logError("Failed to create instance: ", err)
		}
	}
}

// Test function that checks if the cloud provider supports node deletion.
func TestDeleteNodes(provider CloudProvider, nodes []string) {
	if nodeDeleter, supported := provider.DeleteNodes(); !supported {
		skipTest("DeleteNodes not supported by " + provider.Name())
	} else {
		if err := nodeDeleter(nodes); err != nil {
			logError("Failed to delete nodes: ", err)
		}
	}
}

// Test function that checks if the cloud provider supports master upgrade.
func TestUpgradeMaster(provider CloudProvider) {
	if upgrader, supported := provider.UpgradeMaster(); !supported {
		skipTest("UpgradeMaster not supported by " + provider.Name())
	} else {
		if err := upgrader(); err != nil {
			logError("Failed to upgrade master: ", err)
		}
	}
}

// Test function that checks if the cloud provider supports cluster upgrade.
func TestUpgradeCluster(provider CloudProvider) {
	if upgrader, supported := provider.UpgradeCluster(); !supported {
		skipTest("UpgradeCluster not supported by " + provider.Name())
	} else {
		if err := upgrader(); err != nil {
			logError("Failed to upgrade cluster: ", err)
		}
	}
}

// Test function that checks if the cloud provider supports cluster downgrade.
func TestDowngradeCluster(provider CloudProvider) {
	if downgrader, supported := provider.DowngradeCluster(); !supported {
		skipTest("DowngradeCluster not supported by " + provider.Name())
	} else {
		if err := downgrader(); err != nil {
			logError("Failed to downgrade cluster: ", err)
		}
	}
}

// Test function that checks if the cloud provider supports rebooting nodes.
func TestRebootNodes(provider CloudProvider, nodes []string) {
	if rebooter, supported := provider.RebootNodes(); !supported {
		skipTest("RebootNodes not supported by " + provider.Name())
	} else {
		if err := rebooter(nodes); err != nil {
			logError("Failed to reboot nodes: ", err)
		}
	}
}

// Test function that verifies pod count after restart.
func TestVerifyPodCount(provider CloudProvider) {
	if verifier, supported := provider.VerifyPodCount(); !supported {
		skipTest("VerifyPodCount not supported by " + provider.Name())
	} else {
		if err := verifier(); err != nil {
			logError("Failed to verify pod count: ", err)
		}
	}
}

func skipTest(reason string) {
	println("Skipping test: ", reason)
}

func logError(msg string, err error) {
	println(msg, err.Error())
}
