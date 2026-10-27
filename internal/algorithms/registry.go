package algorithms

// All returns algorithms in stable UI/reporting order.
func All() []ConvexHullAlgorithm {
	return []ConvexHullAlgorithm{GrahamScan{}, JarvisMarch{}, QuickHull{}}
}
