package main

type counters struct {
	totalResources  int
	taggedResources int
	totalFiles      int
	taggedFiles     int
}

func (self *counters) Add(other counters) {
	self.totalResources += other.totalResources
	self.taggedResources += other.taggedResources
	self.totalFiles += other.totalFiles
	self.taggedFiles += other.taggedFiles
}
