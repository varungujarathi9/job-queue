package models

type Job struct {
	ID     int    `json:"ID"`
	Type   string `json:"Type"`
	Status string `json:"Status"`
}

type Node struct {
	val  *Job
	next *Node
}

type JobQueue struct {
	head *Node
}

func (queue *JobQueue) Insert(job *Job) {
	newNode := &Node{val: job}
	if queue.head == nil {
		queue.head = newNode
	} else {
		curr := queue.head
		for curr.next != nil {
			curr = curr.next
		}
		curr.next = newNode
	}
}

func (queue *JobQueue) Poll() *Job {
	if queue.head == nil {
		return nil
	}
	first := queue.head
	queue.head = queue.head.next
	return first.val
}
