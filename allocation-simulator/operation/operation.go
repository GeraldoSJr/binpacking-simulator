package operation

type OperationType int

const (
	Insertion OperationType = iota
	Deletion
)

func (op OperationType) String() string {
	if op == Insertion {
		return "insertion"
	}
	if op == Deletion {
		return "deletion"
	}
	return "unknown"
}
