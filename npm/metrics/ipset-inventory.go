package metrics

import (
	"github.com/Azure/azure-container-networking/npm/util"
	"github.com/prometheus/client_golang/prometheus"
)

var ipsetInventoryMap = make(map[string]int)

// GetIPSetInventory returns the number of entries in an IPSet, or 0 if the set doesn't exist.
func GetIPSetInventory(setName string) int {
	val, exists := ipsetInventoryMap[setName]
	if exists {
		return val
	}
	return 0
}

// SetIPSetInventory sets the number of entries in an IPSet and updates a Prometheus metric. When val is 0, the set will be deleted from the IPSetInventory metric.
func SetIPSetInventory(setName string, val int) {
	_, exists := ipsetInventoryMap[setName]
	if exists || val != 0 {
		ipsetInventoryMap[setName] = val
		updateIPSetInventory(setName)
	}
}

// IncIPSetInventory increases the number of entries in an IPSet and updates a Prometheus metric.
func IncIPSetInventory(setName string) {
	_, exists := ipsetInventoryMap[setName]
	if !exists {
		ipsetInventoryMap[setName] = 0
	}
	ipsetInventoryMap[setName]++
	updateIPSetInventory(setName)
}

// DecIPSetInventory decreases the number of entries in an IPSet and updates a Prometheus metric. If the set has 0 members after this operation, it will be deleted from the IPSetInventory metric.
func DecIPSetInventory(setName string) {
	_, exists := ipsetInventoryMap[setName]
	if exists {
		ipsetInventoryMap[setName]--
		updateIPSetInventory(setName)
	}
}

// GetIPSetInventoryLabels returns the labels for the IPSetInventory GaugeVec for a given setName.
func GetIPSetInventoryLabels(setName string) prometheus.Labels {
	return prometheus.Labels{SetNameLabel: setName, SetHashLabel: util.GetHashedName(setName)}
}

// updateIPSetInventory modifies the IPSetInventory metric. A set with 0 members will be deleted from the metric and the ipsetInventoryMap.
func updateIPSetInventory(setName string) {
	labels := GetIPSetInventoryLabels(setName)
	if ipsetInventoryMap[setName] == 0 {
		IPSetInventory.Delete(labels)
		delete(ipsetInventoryMap, setName)
	} else {
		val := float64(ipsetInventoryMap[setName])
		IPSetInventory.With(labels).Set(val)
	}
}
