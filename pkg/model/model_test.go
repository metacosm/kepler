/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	collector_metric "github.com/sustainable-computing-io/kepler/pkg/collector/metric"
	"github.com/sustainable-computing-io/kepler/pkg/power/components/source"
)

var _ = Describe("Test Model Unit", func() {
	var (
		containersMetrics map[string]*collector_metric.ContainerMetrics
		nodeMetrics       collector_metric.NodeMetrics

		machineSensorID = "sensor0"
		machineSocketID = 0

		systemFeatures = []string{"cpu_architecture"}
		systemValues   = []string{"Sandy Bridge"}
	)

	BeforeEach(func() {
		source.SystemCollectionSupported = false // disable the system power collection to use the prediction power model
		setCollectorMetrics()
		containersMetrics = createMockContainersMetrics()
		nodeMetrics = createMockNodeMetrics(containersMetrics)
	})

	// Currently, the model server test models only have data for the DynComponentModelWeight. We cannot get weights for the AbsModelWeight, AbsComponentModelWeight and DynModelWeight
	// Therefore, we can only test this the DynComponentModelWeight component
	// TODO: the make the usage of this different models more transparent, it is currently very hard to know what is going on...
	It("Get container power with no dependency and no node power ", func() {
		// getEstimatorMetrics
		InitEstimateFunctions(usageMetrics, systemFeatures, systemValues)
		Expect(ContainerComponentPowerModelValid).To(Equal(true))

		// update container and node metrics
		componentsEnergies := make(map[int]source.NodeComponentsEnergy)
		componentsEnergies[machineSocketID] = source.NodeComponentsEnergy{
			Pkg:  0,
			Core: 0,
			DRAM: 0,
		}
		nodeMetrics.AddNodeComponentsEnergy(componentsEnergies)
		nodePlatformEnergy := map[string]float64{}
		nodePlatformEnergy[machineSensorID] = 0 // empty
		nodeMetrics.AddLastestPlatformEnergy(nodePlatformEnergy)

		// calculate container energy consumption
		UpdateContainerEnergy(containersMetrics, nodeMetrics)
		// Unit test use is reported by default settings through LR model
		// and following will be reported so EnergyInPkg.Curr will be 9512
		Expect(containersMetrics["containerA"].EnergyInPkg.Curr).To(Equal(uint64(9512)))
	})

	It("Get container power with no dependency but with total node power ", func() {
		// getEstimatorMetrics
		InitEstimateFunctions(usageMetrics, systemFeatures, systemValues)
		Expect(ContainerComponentPowerModelValid).To(Equal(true))

		// update container and node metrics
		componentsEnergies := make(map[int]source.NodeComponentsEnergy)
		componentsEnergies[machineSocketID] = source.NodeComponentsEnergy{
			Pkg:  0,
			Core: 0,
			DRAM: 0,
		}
		nodeMetrics.AddNodeComponentsEnergy(componentsEnergies)
		nodePlatformEnergy := map[string]float64{}
		nodePlatformEnergy[machineSensorID] = 10
		nodeMetrics.AddLastestPlatformEnergy(nodePlatformEnergy)

		// calculate container energy consumption
		UpdateContainerEnergy(containersMetrics, nodeMetrics)
		// Unit test use is reported by default settings through LR model
		// and following will be reported so EnergyInPkg.Curr will be 9512
		Expect(containersMetrics["containerA"].EnergyInPkg.Curr).To(Equal(uint64(9512)))
	})

	It("Get container power with no dependency but with all node power ", func() {
		// getEstimatorMetrics
		InitEstimateFunctions(usageMetrics, systemFeatures, systemValues)
		Expect(ContainerComponentPowerModelValid).To(Equal(true))

		// update container and node metrics
		componentsEnergies := make(map[int]source.NodeComponentsEnergy)
		// the NodeComponentsEnergy is the aggregated energy consumption of the node components
		// then, the components energy consumption is added to the in the nodeMetrics as Agg data
		// this means that, to have a Curr value, we must have at least two Agg data (to have Agg diff)
		// therefore, we need to add two values for NodeComponentsEnergy to have energy values to test
		componentsEnergies[machineSocketID] = source.NodeComponentsEnergy{
			Pkg:  10,
			Core: 10,
			DRAM: 10,
		}
		nodeMetrics.AddNodeComponentsEnergy(componentsEnergies)
		componentsEnergies[machineSocketID] = source.NodeComponentsEnergy{
			Pkg:  18,
			Core: 15,
			DRAM: 11,
		}
		nodeMetrics.AddNodeComponentsEnergy(componentsEnergies)
		nodePlatformEnergy := map[string]float64{}
		nodePlatformEnergy[machineSensorID] = 10
		nodeMetrics.AddLastestPlatformEnergy(nodePlatformEnergy)
		nodePlatformEnergy[machineSensorID] = 15
		nodeMetrics.AddLastestPlatformEnergy(nodePlatformEnergy)

		// calculate container energy consumption

		UpdateContainerEnergy(containersMetrics, nodeMetrics)
		// Unit test use is reported by default settings through LR model
		// and following will be reported so EnergyInPkg.Curr will be 9512
		Expect(containersMetrics["containerA"].EnergyInPkg.Curr).To(Equal(uint64(9512)))
	})
})
