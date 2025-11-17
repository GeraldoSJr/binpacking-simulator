package item

type Item struct {
	ID              string
	CPU             int64
	Memory          int64
	ApplicationName string
}

func NewItem(id string, vcpus int64, memory int64, appName string) *Item {
	return &Item{
		ID:              id,
		CPU:             vcpus,
		Memory:          memory,
		ApplicationName: appName,
	}
}

type ByCPUAndMemoryDesc []Item

func (s ByCPUAndMemoryDesc) Len() int {
	return len(s)
}

func (s ByCPUAndMemoryDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByCPUAndMemoryDesc) Less(i, j int) bool {
	if s[i].CPU != s[j].CPU {
		return s[i].CPU > s[j].CPU
	}
	if s[i].Memory != s[j].Memory {
		return s[i].Memory > s[j].Memory
	}
	return s[i].ID > s[j].ID
}
